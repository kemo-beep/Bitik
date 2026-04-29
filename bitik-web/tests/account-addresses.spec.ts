import { expect, test } from "@playwright/test"
import AxeBuilder from "@axe-core/playwright"
import { envelope, makeFakeJWT, mockJson } from "./helpers/api-mock"

test("addresses page renders and is accessible (mocked)", async ({ page }) => {
  const accessToken = makeFakeJWT({ roles: ["buyer"], sub: "user-1" })

  await page.route("**/api/v1/auth/refresh-token", async (route) => {
    await mockJson(route, envelope({ access_token: accessToken, refresh_token: "r1" }))
  })

  await page.route("**/api/v1/users/me", async (route) => {
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

  await page.route("**/api/v1/buyer/addresses", async (route) => {
    await mockJson(
      route,
      envelope([
        {
          id: "00000000-0000-0000-0000-000000000011",
          full_name: "Buyer One",
          phone: "+15551234567",
          country: "US",
          state: "CA",
          city: "SF",
          district: null,
          postal_code: "94103",
          address_line1: "1 Market St",
          address_line2: null,
          is_default: true,
        },
      ])
    )
  })

  await page.goto("/account/addresses")

  await expect(page.getByRole("heading", { name: "Addresses" })).toBeVisible()

  const results = await new AxeBuilder({ page }).include("main").analyze()
  expect(results.violations).toEqual([])
})

