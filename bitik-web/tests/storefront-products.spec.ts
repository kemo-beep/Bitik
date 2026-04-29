import { expect, test } from "@playwright/test"
import AxeBuilder from "@axe-core/playwright"
import { envelope, mockJson, routeApiMatch, stubAuthBootstrap } from "./helpers/api-mock"

test("products list paginates (mocked)", async ({ page }) => {
  await stubAuthBootstrap(page)

  await page.route(routeApiMatch("/api/v1/public/products"), async (route) => {
    const url = new URL(route.request().url())
    const pageNum = Number(url.searchParams.get("page") ?? "1")
    const items =
      pageNum === 2
        ? [
            {
              id: "00000000-0000-0000-0000-000000000102",
              seller_id: "00000000-0000-0000-0000-000000000202",
              name: "Page 2 Product",
              slug: "page-2-product",
              min_price_cents: 2000,
              max_price_cents: 2000,
              currency: "MMK",
              rating: "0",
              review_count: 0,
            },
          ]
        : [
            {
              id: "00000000-0000-0000-0000-000000000101",
              seller_id: "00000000-0000-0000-0000-000000000201",
              name: "Page 1 Product",
              slug: "page-1-product",
              min_price_cents: 1000,
              max_price_cents: 1500,
              currency: "MMK",
              rating: "4.2",
              review_count: 12,
            },
          ]

    await mockJson(
      route,
      envelope({
        items,
        pagination: { page: pageNum, per_page: 24, total: 48, total_pages: 2 },
      })
    )
  })

  await page.goto("/products")
  await expect(page.getByRole("heading", { name: "Products" })).toBeVisible({ timeout: 15_000 })
  await expect(page.getByText("Page 1 Product")).toBeVisible()

  await page.getByLabel("Go to next page").click()
  await expect(page).toHaveURL(/page=2/)
  await expect(page.getByText("Page 2 Product")).toBeVisible()

  const results = await new AxeBuilder({ page }).include("main").analyze()
  expect(results.violations).toEqual([])
})

