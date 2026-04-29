import { expect, test } from "@playwright/test"
import AxeBuilder from "@axe-core/playwright"
import { envelope, mockJson, routeApiMatch, stubAuthBootstrap } from "./helpers/api-mock"

test("home renders and is accessible (mocked)", async ({ page }) => {
  await stubAuthBootstrap(page)

  await page.route(routeApiMatch("/api/v1/public/home"), async (route) => {
    await mockJson(
      route,
      envelope({
        banners: [
          {
            id: "b1",
            title: "Welcome to Bitik",
            image_url: "http://localhost:8080/static/banner.jpg",
            link_url: "/products",
            placement: "home",
            sort_order: 1,
          },
        ],
        sections: [{ key: "featured", value: { title: "Featured" }, description: "Featured section" }],
        products: [
          {
            id: "00000000-0000-0000-0000-000000000101",
            seller_id: "00000000-0000-0000-0000-000000000201",
            name: "Sample Product",
            slug: "sample-product",
            min_price_cents: 1000,
            max_price_cents: 1500,
            currency: "MMK",
            rating: "4.2",
            review_count: 12,
          },
        ],
      })
    )
  })

  await page.goto("/")
  await expect(page.getByRole("heading", { name: /popular products/i })).toBeVisible({ timeout: 15_000 })

  const results = await new AxeBuilder({ page }).include("main").analyze()
  expect(results.violations).toEqual([])
})

