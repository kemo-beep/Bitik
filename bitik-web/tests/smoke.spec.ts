import { expect, test } from "@playwright/test"

test("home loads (@smoke)", async ({ page }) => {
  await page.goto("/")
  await expect(page).toHaveTitle(/Bitik/i)
  await expect(page.getByRole("banner")).toBeVisible()
})

