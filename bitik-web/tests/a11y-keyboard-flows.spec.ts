import { expect, test } from "@playwright/test"
import AxeBuilder from "@axe-core/playwright"
import { envelope, makeFakeJWT, mockJson, routeApiMatch } from "./helpers/api-mock"

async function stubBuyer(page: import("@playwright/test").Page) {
  const accessToken = makeFakeJWT({ roles: ["buyer"], sub: "buyer-1" })
  await page.route(routeApiMatch("/api/v1/auth/refresh-token"), async (route) => {
    await mockJson(route, envelope({ access_token: accessToken, refresh_token: "r1" }))
  })
  await page.route(routeApiMatch("/api/v1/users/me"), async (route) => {
    await mockJson(
      route,
      envelope({
        id: "00000000-0000-0000-0000-000000000001",
        status: "active",
        email_verified: true,
        phone_verified: true,
        created_at: new Date().toISOString(),
      })
    )
  })
}

test("addresses dialog supports keyboard open/focus/escape", async ({ page }) => {
  await stubBuyer(page)
  await page.route(routeApiMatch("/api/v1/users/me/addresses"), async (route) => {
    await mockJson(route, envelope([]))
  })

  await page.goto("/account/addresses")
  const addButton = page.getByRole("button", { name: "Add address" })
  await addButton.focus()
  await page.keyboard.press("Enter")
  await expect(page.getByRole("dialog")).toBeVisible()
  await expect(page.getByLabel("Full name")).toBeFocused()
  await page.keyboard.press("Escape")
  await expect(page.getByRole("dialog")).toBeHidden()
  await expect(addButton).toBeFocused()
})

test("search form labels and contrast are axe-clean", async ({ page }) => {
  await page.route(routeApiMatch("/api/v1/public/products"), async (route) => {
    await mockJson(route, envelope({ items: [], pagination: { total_pages: 1 } }))
  })
  await page.goto("/search")
  await expect(page.getByRole("heading", { name: "Search" })).toBeVisible()
  const results = await new AxeBuilder({ page }).include("main").analyze()
  expect(results.violations).toEqual([])
})

