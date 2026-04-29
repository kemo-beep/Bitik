import { expect, test, type Page } from "@playwright/test"
import { envelope, makeFakeJWT, mockJson, stubAuthBootstrap } from "./helpers/api-mock"

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

test("checkout place order via keyboard focus and Enter (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)

  await page.route("**/api/v1/buyer/addresses", async (route) => {
    await mockJson(
      route,
      envelope([
        {
          id: "addr-1",
          full_name: "Buyer One",
          address_line1: "1 Market St",
          phone: "+15550001111",
          country: "US",
          is_default: true,
        },
      ])
    )
  })
  await page.route("**/api/v1/buyer/checkout/sessions", async (route) => {
    await mockJson(route, envelope({ id: "chk-1", checkout_session_id: "chk-1" }), 201)
  })
  await page.route("**/api/v1/buyer/checkout/sessions/chk-1", async (route) => {
    await mockJson(
      route,
      envelope({
        id: "chk-1",
        shipping_address_id: "addr-1",
        shipping_method: "standard",
        payment_method: "wave_manual",
        summary: { total_amount: 12000, currency: "MMK" },
      })
    )
  })
  await page.route("**/api/v1/buyer/checkout/sessions/chk-1/**", async (route) => {
    const url = route.request().url()
    if (url.endsWith("/place-order")) {
      await mockJson(route, envelope({ order_id: "ord-kb-1", payment_status: "paid" }), 201)
      return
    }
    await mockJson(route, envelope({ ok: true }))
  })

  await page.goto("/checkout")
  await expect(page.getByRole("heading", { name: "Checkout" })).toBeVisible({ timeout: 15_000 })

  const placeOrder = page.getByRole("button", { name: "Place order" })
  await expect(placeOrder).toBeEnabled({ timeout: 15_000 })
  await placeOrder.focus()
  await page.keyboard.press("Enter")
  await expect(page).toHaveURL("/checkout/success/ord-kb-1")
})

test("admin moderation Get report via keyboard (@critical)", async ({ page }) => {
  await stubAdminSession(page, ["admin"])

  await page.route("**/api/v1/admin/moderation/reports", async (route) => {
    await mockJson(route, envelope({ items: [{ id: "rep-1", status: "open" }] }))
  })

  await page.route("**/api/v1/admin/moderation/reports/rep-1", async (route) => {
    await mockJson(route, envelope({ id: "rep-1", status: "open", reason: "spam" }))
  })

  await page.goto("/admin/moderation")
  await expect(page.getByRole("heading", { name: "Moderation" })).toBeVisible({ timeout: 15_000 })

  const reportField = page.getByPlaceholder("Report ID").first()
  await reportField.focus()
  await page.keyboard.insertText("rep-1")

  const getReport = page.getByRole("button", { name: "Get report" })
  const detailReq = page.waitForRequest(
    (r) => r.method() === "GET" && r.url().includes("/api/v1/admin/moderation/reports/rep-1")
  )
  await getReport.focus()
  await page.keyboard.press("Enter")
  await detailReq
})
