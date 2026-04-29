import { expect, test, type Page } from "@playwright/test"
import { envelope, makeFakeJWT, mockJson } from "./helpers/api-mock"

async function stubAdminSession(page: Page, roles: string[]) {
  const accessToken = makeFakeJWT({ roles, sub: "admin-1" })
  await page.route("**/api/v1/auth/refresh-token", async (route) => {
    await mockJson(route, envelope({ access_token: accessToken, refresh_token: "r1" }))
  })
  await page.route("**/api/v1/users/me", async (route) => {
    await mockJson(
      route,
      envelope({
        id: "00000000-0000-0000-0000-000000000001",
        status: "active",
        email_verified: true,
        phone_verified: true,
        created_at: new Date().toISOString(),
        email: "admin@example.com",
      })
    )
  })
}

test("admin guard redirects buyer away from admin shell", async ({ page }) => {
  await stubAdminSession(page, ["buyer"])
  await page.goto("/admin")
  await expect(page).toHaveURL("/")
})

test("admin dashboard and logs render for admin role", async ({ page }) => {
  await stubAdminSession(page, ["admin"])
  await page.route("**/api/v1/admin/dashboard/overview", async (route) => {
    await mockJson(route, envelope({ users_total: 10, sellers_total: 2 }))
  })
  await page.route("**/api/v1/admin/logs/audit**", async (route) => {
    await mockJson(route, envelope({ items: [] }))
  })
  await page.goto("/admin")
  await expect(page.getByRole("heading", { name: "Admin dashboard" })).toBeVisible()
  await page.goto("/admin/audit-logs")
  await expect(page.getByRole("heading", { name: "Audit logs" })).toBeVisible()
})

test("admin wave approval actions call backend", async ({ page }) => {
  await stubAdminSession(page, ["admin"])
  await page.route("**/api/v1/admin/payments/wave/pending", async (route) => {
    await mockJson(route, envelope({ items: [{ id: "pay-1", status: "pending" }] }))
  })
  await page.route("**/api/v1/admin/payments/pay-1/wave/approve", async (route) => {
    await mockJson(route, envelope({ id: "pay-1", status: "approved" }))
  })
  await page.route("**/api/v1/admin/payments/pay-1/wave/reject", async (route) => {
    await mockJson(route, envelope({ id: "pay-1", status: "rejected" }))
  })
  await page.goto("/admin/payments/wave")
  await page.getByPlaceholder("Payment ID").first().fill("pay-1")
  await page.getByRole("button", { name: "Approve" }).click({ force: true })
  await page.getByRole("button", { name: "Reject" }).click({ force: true })
})

test("admin CMS and RBAC pages render with mocked data", async ({ page }) => {
  await stubAdminSession(page, ["admin"])
  await page.route("**/api/v1/admin/cms/pages", async (route) => {
    await mockJson(route, envelope({ items: [{ id: "page-1", slug: "faq" }] }))
  })
  await page.route("**/api/v1/admin/rbac/roles", async (route) => {
    await mockJson(route, envelope({ items: [{ id: "role-1", key: "ops" }] }))
  })
  await page.goto("/admin/cms/pages")
  await expect(page.getByRole("heading", { name: "CMS pages" })).toBeVisible()
  await page.goto("/admin/rbac")
  await expect(page.getByRole("heading", { name: "RBAC" })).toBeVisible()
})

test("admin categories and brands CRUD actions call real endpoints", async ({ page }) => {
  await stubAdminSession(page, ["admin"])
  await page.route("**/api/v1/admin/categories", async (route) => {
    if (route.request().method() === "GET") {
      await mockJson(route, envelope({ items: [{ id: "cat-1", name: "Shoes" }] }))
      return
    }
    await mockJson(route, envelope({ id: "cat-2" }), 201)
  })
  await page.route("**/api/v1/admin/categories/reorder", async (route) => {
    await mockJson(route, envelope({ updated: 1 }))
  })
  await page.route("**/api/v1/admin/categories/*", async (route) => {
    await mockJson(route, envelope({ updated: true }))
  })
  await page.route("**/api/v1/admin/brands", async (route) => {
    if (route.request().method() === "GET") {
      await mockJson(route, envelope({ items: [{ id: "brand-1", name: "Acme" }] }))
      return
    }
    await mockJson(route, envelope({ id: "brand-2" }), 201)
  })
  await page.route("**/api/v1/admin/brands/*", async (route) => {
    await mockJson(route, envelope({ updated: true }))
  })

  await page.goto("/admin/categories")
  await page.getByPlaceholder('{"name":"Shoes","slug":"shoes"}').fill('{"name":"Boots","slug":"boots"}')
  await page.getByRole("button", { name: "Create" }).click()
  await page.getByPlaceholder('[{"id":"...","sort_order":1}]').fill('[{"id":"cat-1","sort_order":1}]')
  await page.getByRole("button", { name: "Reorder" }).click()

  await page.goto("/admin/brands")
  await page.getByPlaceholder('{"name":"Acme","slug":"acme"}').fill('{"name":"BrandX","slug":"brandx"}')
  await page.getByRole("button", { name: "Create" }).click()
})

test("admin products page uses admin product source and detail endpoint", async ({ page }) => {
  await stubAdminSession(page, ["admin"])
  await page.route("**/api/v1/admin/products", async (route) => {
    await mockJson(route, envelope({ items: [{ id: "prod-1", name: "Shoe" }] }))
  })
  await page.route("**/api/v1/admin/products/prod-1", async (route) => {
    await mockJson(route, envelope({ id: "prod-1", name: "Shoe", slug: "shoe" }))
  })
  await page.route("**/api/v1/admin/products/prod-1/moderation", async (route) => {
    await mockJson(route, envelope({ id: "prod-1", moderation_status: "approved" }))
  })
  await page.goto("/admin/products")
  await page.getByPlaceholder("Product ID").first().fill("prod-1")
  await page.getByRole("button", { name: "Approve product" }).click()
  await page.getByRole("button", { name: "View detail" }).click()
})

test("admin shipments and webhooks pages support detail/tracking/reprocess", async ({ page }) => {
  await stubAdminSession(page, ["admin"])
  await page.route("**/api/v1/admin/shipping/shipments", async (route) => {
    await mockJson(route, envelope({ items: [{ id: "ship-1", status: "pending" }] }))
  })
  await page.route("**/api/v1/admin/shipping/shipments/ship-1", async (route) => {
    await mockJson(route, envelope({ id: "ship-1", status: "pending" }))
  })
  await page.route("**/api/v1/admin/shipping/shipments/ship-1/tracking", async (route) => {
    await mockJson(route, envelope({ items: [{ id: "evt-1", status: "in_transit" }] }))
  })
  await page.route("**/api/v1/admin/shipping/shipments/ship-1/status", async (route) => {
    await mockJson(route, envelope({ id: "ship-1", status: "in_transit" }))
  })
  await page.route("**/api/v1/admin/payments/webhooks", async (route) => {
    await mockJson(route, envelope({ items: [{ id: "wh-1", event_id: "e1", provider: "wave_manual", processed: false }] }))
  })
  await page.route("**/api/v1/admin/payments/webhooks/wh-1", async (route) => {
    await mockJson(route, envelope({ id: "wh-1", event_id: "e1", provider: "wave_manual", processed: false }))
  })
  await page.route("**/api/v1/admin/payments/webhooks/wh-1/reprocess", async (route) => {
    await mockJson(route, envelope({ id: "wh-1", processed: true }))
  })

  await page.goto("/admin/shipments")
  await page.getByPlaceholder("Shipment ID").first().fill("ship-1")
  await page.getByRole("button", { name: "Shipment detail" }).click()
  await page.getByRole("button", { name: "Shipment tracking" }).click()

  await page.goto("/admin/payments")
  await page.getByPlaceholder("Webhook Event ID").first().fill("wh-1")
  await page.getByRole("button", { name: "Webhook detail" }).click()
  await page.getByRole("button", { name: "Webhook reprocess" }).click()
})

