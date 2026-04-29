import { expect, test } from "@playwright/test"
import AxeBuilder from "@axe-core/playwright"
import { envelope, mockJson, stubAuthBootstrap } from "./helpers/api-mock"

test("search updates URL and renders results (mocked)", async ({ page }) => {
  await stubAuthBootstrap(page)

  await page.route("**/api/v1/public/products**", async (route) => {
    const url = new URL(route.request().url())
    const q = url.searchParams.get("q") ?? ""
    await mockJson(
      route,
      envelope({
        items: q
          ? [
              {
                id: "00000000-0000-0000-0000-000000000501",
                seller_id: "00000000-0000-0000-0000-000000000601",
                name: `Result for ${q}`,
                slug: `result-${q}`,
                min_price_cents: 1000,
                max_price_cents: 1000,
                currency: "MMK",
                rating: "0",
                review_count: 0,
              },
            ]
          : [],
        pagination: { page: 1, per_page: 24, total: 1, total_pages: 1 },
      })
    )
  })

  await page.goto("/search")
  await page.getByRole("textbox", { name: "Search products" }).fill("shoes")
  await page.getByRole("button", { name: "Search" }).click()

  await expect(page).toHaveURL(/q=shoes/)
  await expect(page.getByText("Result for shoes")).toBeVisible({ timeout: 15_000 })

  const results = await new AxeBuilder({ page }).include("main").analyze()
  expect(results.violations).toEqual([])
})

