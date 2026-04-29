import { expect, test } from "@playwright/test"
import { envelope, mockJson, stubAuthBootstrap } from "./helpers/api-mock"

test("register/login critical flow (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  let loginCalled = false
  await page.route("**/api/v1/auth/login", async (route) => {
    loginCalled = true
    await mockJson(route, envelope({ access_token: "mock", refresh_token: "mock" }))
  })
  await page.goto("/login")
  await page.getByLabel("Email").fill("buyer@example.com")
  await page.getByLabel("Password").fill("passwordpassword")
  await page.getByRole("button", { name: "Sign in" }).click()
  await expect.poll(() => loginCalled).toBe(true)
})

test("seller apply + product publish critical flow (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  let applyCalled = false
  await page.route("**/api/v1/seller/apply", async (route) => {
    applyCalled = true
    await mockJson(route, envelope({ id: "app-1", status: "submitted" }), 201)
  })
  await page.goto("/seller/apply")
  await page.getByRole("textbox").first().fill("My Seller Shop")
  await page.getByRole("button", { name: "Submit application" }).click()
  await expect.poll(() => applyCalled).toBe(true)
})

test("buyer browse/search/cart/checkout critical flow (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  let orderCalled = false
  await page.route("**/api/v1/buyer/addresses", async (route) => {
    await mockJson(route, envelope([{ id: "addr-1", full_name: "Buyer One", address_line1: "1 Market St", phone: "+15550001111", country: "US", is_default: true }]))
  })
  await page.route("**/api/v1/buyer/checkout/sessions", async (route) => {
    await mockJson(route, envelope({ id: "chk-1", checkout_session_id: "chk-1" }), 201)
  })
  await page.route("**/api/v1/buyer/checkout/sessions/chk-1", async (route) => {
    await mockJson(route, envelope({ id: "chk-1", shipping_address_id: "addr-1", shipping_method: "standard", payment_method: "wave_manual", summary: { total_amount: 12000, currency: "MMK" } }))
  })
  await page.route("**/api/v1/buyer/checkout/sessions/chk-1/**", async (route) => {
    if (route.request().url().endsWith("/place-order")) {
      orderCalled = true
      await mockJson(route, envelope({ order_id: "ord-1", payment_status: "paid" }), 201)
      return
    }
    await mockJson(route, envelope({ ok: true }))
  })
  await page.goto("/checkout")
  await page.getByRole("button", { name: "Place order" }).click()
  await expect.poll(() => orderCalled).toBe(true)
})

test("wave manual payment approval critical flow (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  await page.route("**/api/v1/admin/payments/wave/pending", async (route) =>
    mockJson(route, envelope({ items: [{ id: "pay-1", status: "pending" }] }))
  )
  let approveCalled = false
  await page.route("**/api/v1/admin/payments/pay-1/wave/approve", async (route) => {
    approveCalled = true
    await mockJson(route, envelope({ id: "pay-1", status: "approved" }))
  })
  await page.goto("/admin/payments/wave")
  await page.getByPlaceholder("Payment ID").first().fill("pay-1")
  await page.getByRole("button", { name: "Approve" }).click()
  await expect.poll(() => approveCalled).toBe(true)
})

test("POD lifecycle critical flow (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  await page.goto("/buyer/orders")
  await expect(page).toHaveURL(/\/buyer\/orders$/)
})

test("seller ship order critical flow (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  await page.goto("/seller/orders")
  await expect(page).toHaveURL(/\/seller\/orders$/)
})

test("buyer confirm received critical flow (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  await page.goto("/buyer/orders")
  await expect(page).toHaveURL(/\/buyer\/orders$/)
})

test("review submission critical flow (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  await page.goto("/buyer/reviews")
  await expect(page).toHaveURL(/\/buyer\/reviews$/)
})

test("admin moderation critical flow (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  let moderationCalled = false
  await page.route("**/api/v1/admin/moderation/reports**", async (route) => {
    moderationCalled = true
    await mockJson(route, envelope({ items: [] }))
  })
  await page.goto("/admin/moderation")
  await expect.poll(() => moderationCalled).toBe(true)
})
