import { expect, test } from "@playwright/test"
import { envelope, mockJson, stubAuthBootstrap } from "./helpers/api-mock"

test("payment edge cases show rejected state and allow cancel (mocked)", async ({ page }) => {
  await stubAuthBootstrap(page)

  await page.route("**/api/v1/buyer/payments/create-intent", async (route) => {
    await mockJson(route, envelope({ payment_id: "pay-1" }))
  })
  await page.route("**/api/v1/buyer/payments/pay-1", async (route) => {
    await mockJson(route, envelope({ id: "pay-1", status: "rejected" }))
  })
  await page.route("**/api/v1/buyer/payments/pay-1/cancel", async (route) => {
    await mockJson(route, envelope({ id: "pay-1", status: "cancelled" }))
  })
  await page.route("**/api/v1/buyer/payments/confirm", async (route) => {
    await mockJson(route, envelope({ id: "pay-1", status: "pending_manual" }))
  })

  await page.goto("/checkout/payment?order_id=ord-1")
  await expect(page.getByRole("heading", { name: "Payment" })).toBeVisible({ timeout: 15_000 })

  await page.getByRole("button", { name: "Create payment intent" }).click()
  await expect(page.getByText(/Payment rejected|Payment failed|Payment timeout/)).toBeVisible()

  await page.getByRole("button", { name: "Cancel payment" }).click()
})

