import { expect, test } from "@playwright/test"
import { envelope, mockJson, routeApiMatch, stubAuthBootstrap } from "./helpers/api-mock"

test("stored locale updates document lang and direction", async ({ page }) => {
  await stubAuthBootstrap(page)
  await page.goto("/")
  await expect(page.locator("html")).toHaveAttribute("dir", "ltr")
  await page.evaluate(() => window.localStorage.setItem("bitik.locale.v1", "ar"))
  await page.reload()
  await expect(page.locator("html")).toHaveAttribute("dir", "rtl")
  await expect(page.locator("html")).toHaveAttribute("lang", "ar")
  await page.evaluate(() => window.localStorage.setItem("bitik.locale.v1", "fr"))
  await page.reload()
  await expect(page.locator("html")).toHaveAttribute("dir", "ltr")
  await expect(page.locator("html")).toHaveAttribute("lang", "fr")
})

test("analytics ingestion fires for add-to-cart interaction (@critical)", async ({ page }) => {
  await stubAuthBootstrap(page)
  await page.route(routeApiMatch("/api/v1/public/products/p-1"), async (route) => {
    const url = route.request().url()
    if (url.includes("/reviews")) {
      await mockJson(route, envelope({ items: [] }))
      return
    }
    if (url.includes("/related")) {
      await mockJson(route, envelope([]))
      return
    }
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
  await page.route(routeApiMatch("/api/v1/public/sellers/"), async (route) => {
    await mockJson(route, envelope({ id: "s-1", shop_name: "Seller One" }))
  })

  const events: unknown[] = []
  await page.route(routeApiMatch("/api/v1/analytics/events"), async (route) => {
    const payload = route.request().postDataJSON()
    events.push(payload)
    await mockJson(route, envelope({}))
  })

  await page.goto("/products/p-1")
  await page.getByRole("button", { name: "Add to cart (stub)" }).click()
  await expect.poll(() => events.length, { timeout: 15_000 }).toBeGreaterThan(0)
})

