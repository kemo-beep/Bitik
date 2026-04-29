import { expect, test } from "@playwright/test"
import {
  envelope,
  makeFakeJWT,
  mockJson,
  routeApiMatch,
  stubAuthBootstrap,
  stubLoggedInSession,
} from "./helpers/api-mock"

function stubSellerBootstrap(page: Parameters<typeof stubLoggedInSession>[0]) {
  return stubLoggedInSession(page, ["seller"], {
    id: "00000000-0000-0000-0000-000000000111",
    email: "seller@example.com",
    sub: "seller-1",
  })
}

function stubAdminSession(page: Parameters<typeof stubLoggedInSession>[0]) {
  return stubLoggedInSession(page, ["admin"], {
    id: "00000000-0000-0000-0000-000000000001",
    email: "admin@example.com",
    sub: "admin-1",
  })
}

test("register submits auth register (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  const token = makeFakeJWT({ roles: ["buyer"], sub: "new-user" })
  let registered = false
  await page.route(routeApiMatch("/api/v1/auth/register"), async (route) => {
    registered = true
    await mockJson(route, envelope({ access_token: token, refresh_token: "r-new" }))
  })
  await page.route(routeApiMatch("/api/v1/users/me"), async (route) => {
    await mockJson(
      route,
      envelope({
        id: "00000000-0000-0000-0000-000000000222",
        status: "active",
        email_verified: false,
        phone_verified: false,
        created_at: new Date().toISOString(),
        email: "new@example.com",
      })
    )
  })

  await page.goto("/register")
  await page.getByLabel("Email").fill("new@example.com")
  await page.getByLabel("Password").fill("passwordpassword")
  await page.getByRole("button", { name: "Create account" }).click()
  await expect.poll(() => registered).toBe(true)
  await expect(page).toHaveURL("/")
})

test("storefront home and search load (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  await page.route(routeApiMatch("/api/v1/public/home"), async (route) => {
    await mockJson(route, envelope({ banners: [], sections: [], products: [] }))
  })
  await page.goto("/")
  await expect(page.getByRole("banner")).toBeVisible({ timeout: 15_000 })

  await page.route(routeApiMatch("/api/v1/public/products"), async (route) => {
    await mockJson(route, envelope({ items: [], pagination: { total_pages: 1 } }))
  })
  await page.goto("/search?q=test")
  await expect(page.getByRole("heading", { name: "Search" })).toBeVisible()
})

test("checkout with POD pre-selected places order (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  await page.route(routeApiMatch("/api/v1/buyer/addresses"), async (route) => {
    await mockJson(
      route,
      envelope([
        {
          id: "addr-1",
          full_name: "Buyer",
          address_line1: "1 St",
          phone: "+1",
          country: "US",
          is_default: true,
        },
      ])
    )
  })
  await page.route(routeApiMatch("/api/v1/buyer/checkout/sessions"), async (route) => {
    await mockJson(route, envelope({ id: "chk-pod", checkout_session_id: "chk-pod" }), 201)
  })
  let placed = false
  await page.route(routeApiMatch("/api/v1/buyer/checkout/sessions/chk-pod"), async (route) => {
    const url = route.request().url()
    if (url.endsWith("/place-order")) {
      placed = true
      await mockJson(route, envelope({ order_id: "ord-pod", payment_status: "paid" }), 201)
      return
    }
    if (route.request().method() === "GET") {
      await mockJson(
        route,
        envelope({
          id: "chk-pod",
          shipping_address_id: "addr-1",
          shipping_method: "standard",
          payment_method: "pod",
          summary: { total_amount: 5000, currency: "MMK" },
        })
      )
      return
    }
    await mockJson(route, envelope({ ok: true }))
  })

  await page.goto("/checkout")
  await page.getByRole("button", { name: "Place order" }).click()
  await expect.poll(() => placed).toBe(true)
})

test("Wave manual pending page (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  await page.goto("/checkout/pending?order_id=ord-pending-1")
  await expect(page.getByRole("heading", { name: "Awaiting manual confirmation" })).toBeVisible()
  await expect(page.getByText(/ord-pending-1/)).toBeVisible()
})

test("buyer order detail shows tracking section (@critical)", async ({ page }) => {
  await stubLoggedInSession(page, ["buyer"], { email: "buyer@example.com" })
  await page.route(routeApiMatch("/api/v1/buyer/orders/ord-1"), async (route) => {
    await mockJson(route, envelope({ id: "ord-1", status: "shipped" }))
  })
  await page.route(routeApiMatch("/api/v1/buyer/orders/ord-1/items"), async (route) => {
    await mockJson(route, envelope({ items: [] }))
  })
  await page.route(routeApiMatch("/api/v1/buyer/orders/ord-1/status-history"), async (route) => {
    await mockJson(route, envelope({ items: [] }))
  })
  await page.route(routeApiMatch("/api/v1/buyer/orders/ord-1/tracking"), async (route) => {
    await mockJson(route, envelope({ carrier: "TestCarrier", events: [] }))
  })
  await page.route(routeApiMatch("/api/v1/buyer/orders/ord-1/invoice"), async (route) => {
    await mockJson(route, envelope({}))
  })

  await page.goto("/account/orders/ord-1")
  await expect(page.getByRole("heading", { name: "Order detail" })).toBeVisible({ timeout: 15_000 })
  await expect(page.getByRole("heading", { name: "Tracking" })).toBeVisible()
})

test("seller creates product from new product page (@critical)", async ({ page }) => {
  await stubSellerBootstrap(page)
  let created = false
  await page.route(routeApiMatch("/api/v1/seller/products"), async (route) => {
    if (route.request().method() === "POST") {
      created = true
      await mockJson(route, envelope({ id: "prod-new", name: "Widget", slug: "widget" }), 201)
      return
    }
    await mockJson(route, envelope({ items: [] }))
  })

  await page.goto("/seller/products/new")
  await expect(page.getByRole("heading", { name: "New product" })).toBeVisible({ timeout: 15_000 })
  await page.getByPlaceholder("Name").fill("Widget")
  await page.getByPlaceholder("Slug").fill("widget")
  await page.getByRole("button", { name: "Create product" }).click()
  await expect.poll(() => created).toBe(true)
})

test("seller ships order from order detail (@critical)", async ({ page }) => {
  await stubSellerBootstrap(page)
  let shipped = false
  await page.route(routeApiMatch("/api/v1/seller/orders/ord-s1"), async (route) => {
    const url = route.request().url()
    if (url.includes("/orders/ord-s1/ship") && route.request().method() === "POST") {
      shipped = true
      await mockJson(route, envelope({ ok: true }))
      return
    }
    if (url.includes("/shipments")) {
      await mockJson(route, envelope({ items: [] }))
      return
    }
    if (url.includes("/items")) {
      await mockJson(route, envelope({ items: [] }))
      return
    }
    await mockJson(route, envelope({ id: "ord-s1", status: "paid" }))
  })

  await page.goto("/seller/orders/ord-s1")
  await expect(page.getByRole("heading", { name: "Order detail" })).toBeVisible({ timeout: 15_000 })
  await page.getByRole("button", { name: "Ship" }).first().click()
  await expect.poll(() => shipped).toBe(true)
})

test("admin approves seller application (@critical)", async ({ page }) => {
  await stubAdminSession(page)
  await page.route(routeApiMatch("/api/v1/admin/seller-applications"), async (route) => {
    await mockJson(route, envelope({ items: [{ id: "app-99", status: "pending" }] }))
  })
  let reviewed = false
  await page.route(routeApiMatch("/api/v1/admin/seller-applications/app-99/review"), async (route) => {
    reviewed = true
    await mockJson(route, envelope({ id: "app-99", decision: "approved" }))
  })

  await page.goto("/admin/sellers")
  await expect(page.getByRole("heading", { name: "Sellers" })).toBeVisible({ timeout: 15_000 })
  await page.getByPlaceholder("Application ID").first().fill("app-99")
  await page.getByRole("button", { name: "Approve application" }).click({ force: true })
  await expect.poll(() => reviewed).toBe(true)
})
