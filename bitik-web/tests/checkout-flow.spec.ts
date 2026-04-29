import { expect, test } from "@playwright/test"
import { envelope, mockJson, routeApiMatch, stubLoggedInSession } from "./helpers/api-mock"

test("checkout flow places order (mocked)", async ({ page }) => {
  await stubLoggedInSession(page, ["buyer"], { email: "buyer@example.com" })

  await page.route(routeApiMatch("/api/v1/buyer/addresses"), async (route) => {
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
  await page.route(routeApiMatch("/api/v1/buyer/checkout/sessions"), async (route) => {
    await mockJson(route, envelope({ id: "chk-1", checkout_session_id: "chk-1" }), 201)
  })
  await page.route(routeApiMatch("/api/v1/buyer/checkout/sessions/chk-1"), async (route) => {
    const url = route.request().url()
    if (url.endsWith("/place-order")) {
      await mockJson(route, envelope({ order_id: "ord-1", payment_status: "paid" }), 201)
      return
    }
    if (route.request().method() === "GET") {
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
      return
    }
    await mockJson(route, envelope({ ok: true }))
  })

  await page.goto("/checkout")
  await expect(page.getByRole("heading", { name: "Checkout" })).toBeVisible({ timeout: 15_000 })
  await page.getByRole("button", { name: "Place order" }).click()
  await expect(page).toHaveURL("/checkout/success/ord-1")
})

