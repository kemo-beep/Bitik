import { expect, test } from "@playwright/test"
import { envelope, mockJson, stubAuthBootstrap } from "./helpers/api-mock"

test("language switcher toggles document direction", async ({ page }) => {
  await stubAuthBootstrap(page)
  await page.goto("/")
  await expect(page.locator("html")).toHaveAttribute("dir", "ltr")
  await page.getByLabel("Language").selectOption("ar")
  await expect(page.locator("html")).toHaveAttribute("dir", "rtl")
})

test("analytics ingestion fires for add-to-cart interaction (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  await page.route("**/api/v1/public/products/*", async (route) => {
    await mockJson(
      route,
      envelope({
        id: "p-1",
        name: "Product A",
        seller_id: "s-1",
        min_price_cents: 1000,
        max_price_cents: 1000,
        currency: "MMK",
        images: [],
      })
    )
  })
  await page.route("**/api/v1/public/products/*/reviews**", async (route) => {
    await mockJson(route, envelope({ items: [] }))
  })
  await page.route("**/api/v1/public/products/*/related**", async (route) => {
    await mockJson(route, envelope([]))
  })
  await page.route("**/api/v1/public/sellers/*", async (route) => {
    await mockJson(route, envelope({ id: "s-1", shop_name: "Seller One" }))
  })

  const events: unknown[] = []
  await page.route("**/api/v1/analytics/events", async (route) => {
    const payload = route.request().postDataJSON()
    events.push(payload)
    await mockJson(route, envelope({}))
  })

  await page.goto("/products/p-1")
  await page.getByRole("button", { name: "Add to cart (stub)" }).click()
  await expect.poll(() => events.length, { timeout: 15_000 }).toBeGreaterThan(0)
})

