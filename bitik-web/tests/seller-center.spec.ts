import { expect, test, type Page } from "@playwright/test"
import { envelope, makeFakeJWT, mockJson, routeApiMatch } from "./helpers/api-mock"

async function stubSellerBootstrap(page: Page, roles: string[] = ["seller"]) {
  const accessToken = makeFakeJWT({ roles, sub: "seller-1" })
  await page.route(routeApiMatch("/api/v1/auth/refresh-token"), async (route) => {
    await mockJson(route, envelope({ access_token: accessToken, refresh_token: "r1" }))
  })
  await page.route(routeApiMatch("/api/v1/users/me"), async (route) => {
    await mockJson(
      route,
      envelope({
        id: "00000000-0000-0000-0000-000000000111",
        status: "active",
        email_verified: true,
        phone_verified: true,
        created_at: new Date().toISOString(),
        email: "seller@example.com",
      })
    )
  })
}

test("seller onboarding flow renders and submits application (mocked)", async ({ page }) => {
  await stubSellerBootstrap(page, ["buyer"])
  await page.route(routeApiMatch("/api/v1/seller/apply"), async (route) => {
    await mockJson(route, envelope({ id: "app-1", status: "submitted" }), 201)
  })
  await page.goto("/seller/apply")
  await expect(page.getByRole("heading", { name: "Apply as seller" })).toBeVisible({ timeout: 15_000 })
  await page.getByRole("textbox").first().fill("My Seller Shop")
  await page.getByRole("button", { name: "Submit application" }).click()
  await expect(page.getByText("submitted")).toBeVisible()
})

test("seller product management list actions render (mocked)", async ({ page }) => {
  await stubSellerBootstrap(page, ["seller"])
  await page.route(routeApiMatch("/api/v1/seller/products"), async (route) => {
    await mockJson(
      route,
      envelope({
        items: [{ id: "prod-1", name: "Seller Product 1", status: "draft" }],
      })
    )
  })
  await page.route(routeApiMatch("/api/v1/seller/products/prod-1/"), async (route) => {
    await mockJson(route, envelope({ ok: true }))
  })
  await page.route(routeApiMatch("/api/v1/seller/products/prod-1"), async (route) => {
    if (route.request().method() === "DELETE") {
      await mockJson(route, envelope({ ok: true }))
      return
    }
    await mockJson(route, envelope({ id: "prod-1" }))
  })

  await page.goto("/seller/products")
  await expect(page.getByText("Seller Product 1")).toBeVisible({ timeout: 15_000 })
  await page.getByRole("button", { name: "Publish", exact: true }).click()
  await page.getByRole("button", { name: "Unpublish", exact: true }).click()
  await page.getByRole("button", { name: "Duplicate" }).click()
})

test("seller wallet payout request works (mocked)", async ({ page }) => {
  await stubSellerBootstrap(page, ["seller"])
  await page.route(routeApiMatch("/api/v1/seller/wallet"), async (route) => {
    await mockJson(route, envelope({ available_balance_cents: 123000 }))
  })
  await page.route(routeApiMatch("/api/v1/seller/wallet/transactions"), async (route) => {
    await mockJson(route, envelope({ items: [] }))
  })
  await page.route(routeApiMatch("/api/v1/seller/payouts"), async (route) => {
    await mockJson(route, envelope({ items: [] }))
  })
  await page.route(routeApiMatch("/api/v1/seller/payouts/request"), async (route) => {
    await mockJson(route, envelope({ id: "payout-1", status: "pending" }))
  })
  await page.route(routeApiMatch("/api/v1/seller/bank-accounts"), async (route) => {
    if (route.request().method() === "POST") {
      await mockJson(route, envelope({ id: "bank-1", bank_name: "Bitik Bank", account_number_masked: "****1234" }), 201)
      return
    }
    await mockJson(route, envelope({ items: [{ id: "bank-1", bank_name: "Bitik Bank", account_number_masked: "****1234" }] }))
  })
  await page.route(routeApiMatch("/api/v1/seller/bank-accounts/"), async (route) => {
    await mockJson(route, envelope({ ok: true }))
  })

  await page.goto("/seller/wallet")
  await expect(page.getByRole("heading", { name: "Wallet and payouts" })).toBeVisible({ timeout: 15_000 })
  await page.getByPlaceholder("Bank name").fill("Bitik Bank")
  await page.getByPlaceholder("Account name").fill("Seller One")
  await page.getByPlaceholder("Account number").fill("123456789")
  await page.getByRole("button", { name: "Add account" }).click()
  await page.getByPlaceholder("Amount (cents)").fill("10000")
  await page.getByRole("button", { name: "Request payout" }).click()
  await expect(page.getByText("Payout request")).toBeVisible()
})

test("seller inventory low-stock and adjust actions work (mocked)", async ({ page }) => {
  await stubSellerBootstrap(page, ["seller"])
  await page.route(routeApiMatch("/api/v1/seller/inventory/low-stock"), async (route) => {
    await mockJson(route, envelope({ items: [{ id: "inv-low-1", quantity: 1 }] }))
  })
  await page.route(routeApiMatch("/api/v1/seller/inventory"), async (route) => {
    await mockJson(route, envelope({ items: [{ id: "inv-1", quantity: 12 }] }))
  })
  await page.route(routeApiMatch("/api/v1/seller/inventory/inv-1/adjust"), async (route) => {
    await mockJson(route, envelope({ id: "inv-1", quantity: 13 }))
  })
  await page.route(routeApiMatch("/api/v1/seller/inventory/bulk-update"), async (route) => {
    await mockJson(route, envelope({ updated: 1 }))
  })

  await page.goto("/seller/inventory")
  await expect(page.getByRole("heading", { name: "Inventory", exact: true })).toBeVisible()
  await page.getByRole("button", { name: "Adjust" }).click()
  await expect(page.getByText("inv-low-1")).toBeVisible()
})

test("seller order transitions and shipping actions work (mocked)", async ({ page }) => {
  await stubSellerBootstrap(page, ["seller"])
  await page.route(routeApiMatch("/api/v1/seller/orders"), async (route) => {
    await mockJson(route, envelope({ items: [{ id: "ord-1", status: "pending" }] }))
  })
  await page.route(routeApiMatch("/api/v1/seller/orders/ord-1"), async (route) => {
    await mockJson(route, envelope({ id: "ord-1", status: "pending" }))
  })
  await page.route(routeApiMatch("/api/v1/seller/orders/ord-1/items"), async (route) => {
    await mockJson(route, envelope({ items: [{ id: "item-1" }] }))
  })
  await page.route(routeApiMatch("/api/v1/seller/orders/ord-1/shipments"), async (route) => {
    await mockJson(route, envelope({ items: [{ id: "ship-1" }] }))
  })
  await page.route(routeApiMatch("/api/v1/seller/orders/ord-1/"), async (route) => {
    await mockJson(route, envelope({ ok: true }))
  })
  await page.route(routeApiMatch("/api/v1/seller/shipments/ship-1/"), async (route) => {
    await mockJson(route, envelope({ ok: true }))
  })

  await page.goto("/seller/orders")
  await page.getByRole("link", { name: "Order ord-1" }).click()
  await expect(page.getByRole("heading", { name: "Order detail" })).toBeVisible()
  await page.getByRole("button", { name: "Accept" }).click()
  await page.getByRole("button", { name: "Pack" }).click()
  await page.getByRole("button", { name: "Ship" }).click()
})

