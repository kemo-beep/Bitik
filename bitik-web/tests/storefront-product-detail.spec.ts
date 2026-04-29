import { expect, test } from "@playwright/test"
import AxeBuilder from "@axe-core/playwright"
import { envelope, mockJson, stubAuthBootstrap } from "./helpers/api-mock"

test("product detail loads (mocked)", async ({ page }) => {
  await stubAuthBootstrap(page)

  const productId = "00000000-0000-0000-0000-000000000301"
  const sellerId = "00000000-0000-0000-0000-000000000401"

  await page.route("**/api/v1/public/products/**", async (route) => {
    const url = route.request().url()
    // Handle the base product detail request; let review/related handlers handle their own routes.
    if (url.includes(`/public/products/${productId}`) && !url.includes("/reviews") && !url.includes("/related")) {
      await mockJson(
        route,
        envelope({
          id: productId,
          seller_id: sellerId,
          name: "Detail Product",
          slug: "detail-product",
          description: "A detailed description.",
          min_price_cents: 1000,
          max_price_cents: 2000,
          currency: "MMK",
          rating: "4.5",
          review_count: 3,
          images: [
            { id: "i1", url: "http://localhost:8080/static/p.jpg", alt_text: "p", sort_order: 1, is_primary: true },
          ],
          variants: [
            { id: "v1", sku: "SKU-1", name: "Variant A", price_cents: 1500, currency: "MMK", weight_grams: 100 },
          ],
        })
      )
      return
    }
    await route.fallback()
  })

  await page.route(`**/api/v1/public/products/${productId}/reviews**`, async (route) => {
    await mockJson(route, envelope({ items: [{ id: "r1", rating: 5, title: "Great", body: "Nice.", created_at: new Date().toISOString() }] }))
  })

  await page.route(`**/api/v1/public/products/${productId}/related**`, async (route) => {
    await mockJson(
      route,
      envelope([
        {
          id: "00000000-0000-0000-0000-000000000302",
          seller_id: sellerId,
          name: "Related Product",
          slug: "related-product",
          min_price_cents: 900,
          max_price_cents: 1100,
          currency: "MMK",
          rating: "0",
          review_count: 0,
        },
      ])
    )
  })

  await page.route(`**/api/v1/public/sellers/${sellerId}**`, async (route) => {
    await mockJson(
      route,
      envelope({
        id: sellerId,
        shop_name: "Great Shop",
        slug: "great-shop",
        description: "Shop description",
        logo_url: null,
        banner_url: null,
        rating: "4.1",
        total_sales: 123,
      })
    )
  })

  await page.goto(`/products/${productId}`)
  await expect(page.getByRole("heading", { name: "Detail Product" })).toBeVisible({ timeout: 15_000 })
  await expect(page.getByText("Related products")).toBeVisible()

  const results = await new AxeBuilder({ page }).include("main").analyze()
  expect(results.violations).toEqual([])
})

