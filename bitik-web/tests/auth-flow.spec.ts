import { expect, test } from "@playwright/test"
import AxeBuilder from "@axe-core/playwright"
import { envelope, makeFakeJWT, mockJson, routeApiMatch, stubAuthBootstrap } from "./helpers/api-mock"

test("login form is accessible (@smoke)", async ({ page }) => {
  await stubAuthBootstrap(page)
  await page.goto("/login")
  await expect(page.getByRole("button", { name: "Sign in" })).toBeVisible({ timeout: 15_000 })

  const results = await new AxeBuilder({ page }).include("main").analyze()
  expect(results.violations).toEqual([])
})

test("login -> me bootstrap works (mocked)", async ({ page }) => {
  await stubAuthBootstrap(page)

  const accessToken = makeFakeJWT({ roles: ["buyer"], sub: "user-1" })

  await page.route(routeApiMatch("/api/v1/auth/login"), async (route) => {
    await mockJson(route, envelope({ access_token: accessToken, refresh_token: "r1" }))
  })
  await page.route(routeApiMatch("/api/v1/users/me"), async (route) => {
    await mockJson(
      route,
      envelope({
        id: "00000000-0000-0000-0000-000000000001",
        status: "active",
        email_verified: true,
        phone_verified: false,
        created_at: new Date().toISOString(),
        email: "buyer@example.com",
      })
    )
  })

  await page.goto("/login")
  await page.getByLabel("Email").fill("buyer@example.com")
  await page.getByLabel("Password").fill("passwordpassword")
  await page.getByRole("button", { name: "Sign in" }).click()

  await expect(page).toHaveURL("/")
})

