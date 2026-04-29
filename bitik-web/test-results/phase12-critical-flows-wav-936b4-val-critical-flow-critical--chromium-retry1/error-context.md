# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: phase12-critical-flows.spec.ts >> wave manual payment approval critical flow (@critical)
- Location: tests/phase12-critical-flows.spec.ts:56:1

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: locator.fill: Test timeout of 30000ms exceeded.
Call log:
  - waiting for getByPlaceholder('Payment ID').first()

```

# Page snapshot

```yaml
- main [ref=e2]:
  - alert [ref=e3]:
    - img [ref=e5]
    - generic [ref=e7]:
      - paragraph [ref=e8]: Application error
      - paragraph [ref=e9]: Failed to load chunk /_next/static/chunks/97d834072840f18f.js from module 87592
    - button "Try again" [ref=e10]:
      - img [ref=e11]
      - text: Try again
```

# Test source

```ts
  1   | import { expect, test } from "@playwright/test"
  2   | import { envelope, mockJson, stubAuthBootstrap } from "./helpers/api-mock"
  3   | 
  4   | test("register/login critical flow (@critical)", async ({ page }) => {
  5   |   await stubAuthBootstrap(page)
  6   |   let loginCalled = false
  7   |   await page.route("**/api/v1/auth/login", async (route) => {
  8   |     loginCalled = true
  9   |     await mockJson(route, envelope({ access_token: "mock", refresh_token: "mock" }))
  10  |   })
  11  |   await page.goto("/login")
  12  |   await page.getByLabel("Email").fill("buyer@example.com")
  13  |   await page.getByLabel("Password").fill("passwordpassword")
  14  |   await page.getByRole("button", { name: "Sign in" }).click()
  15  |   await expect.poll(() => loginCalled).toBe(true)
  16  | })
  17  | 
  18  | test("seller apply + product publish critical flow (@critical)", async ({ page }) => {
  19  |   await stubAuthBootstrap(page)
  20  |   let applyCalled = false
  21  |   await page.route("**/api/v1/seller/apply", async (route) => {
  22  |     applyCalled = true
  23  |     await mockJson(route, envelope({ id: "app-1", status: "submitted" }), 201)
  24  |   })
  25  |   await page.goto("/seller/apply")
  26  |   await page.getByRole("textbox").first().fill("My Seller Shop")
  27  |   await page.getByRole("button", { name: "Submit application" }).click()
  28  |   await expect.poll(() => applyCalled).toBe(true)
  29  | })
  30  | 
  31  | test("buyer browse/search/cart/checkout critical flow (@critical)", async ({ page }) => {
  32  |   await stubAuthBootstrap(page)
  33  |   let orderCalled = false
  34  |   await page.route("**/api/v1/buyer/addresses", async (route) => {
  35  |     await mockJson(route, envelope([{ id: "addr-1", full_name: "Buyer One", address_line1: "1 Market St", phone: "+15550001111", country: "US", is_default: true }]))
  36  |   })
  37  |   await page.route("**/api/v1/buyer/checkout/sessions", async (route) => {
  38  |     await mockJson(route, envelope({ id: "chk-1", checkout_session_id: "chk-1" }), 201)
  39  |   })
  40  |   await page.route("**/api/v1/buyer/checkout/sessions/chk-1", async (route) => {
  41  |     await mockJson(route, envelope({ id: "chk-1", shipping_address_id: "addr-1", shipping_method: "standard", payment_method: "wave_manual", summary: { total_amount: 12000, currency: "MMK" } }))
  42  |   })
  43  |   await page.route("**/api/v1/buyer/checkout/sessions/chk-1/**", async (route) => {
  44  |     if (route.request().url().endsWith("/place-order")) {
  45  |       orderCalled = true
  46  |       await mockJson(route, envelope({ order_id: "ord-1", payment_status: "paid" }), 201)
  47  |       return
  48  |     }
  49  |     await mockJson(route, envelope({ ok: true }))
  50  |   })
  51  |   await page.goto("/checkout")
  52  |   await page.getByRole("button", { name: "Place order" }).click()
  53  |   await expect.poll(() => orderCalled).toBe(true)
  54  | })
  55  | 
  56  | test("wave manual payment approval critical flow (@critical)", async ({ page }) => {
  57  |   await stubAuthBootstrap(page)
  58  |   await page.route("**/api/v1/admin/payments/wave/pending", async (route) =>
  59  |     mockJson(route, envelope({ items: [{ id: "pay-1", status: "pending" }] }))
  60  |   )
  61  |   let approveCalled = false
  62  |   await page.route("**/api/v1/admin/payments/pay-1/wave/approve", async (route) => {
  63  |     approveCalled = true
  64  |     await mockJson(route, envelope({ id: "pay-1", status: "approved" }))
  65  |   })
  66  |   await page.goto("/admin/payments/wave")
> 67  |   await page.getByPlaceholder("Payment ID").first().fill("pay-1")
      |                                                     ^ Error: locator.fill: Test timeout of 30000ms exceeded.
  68  |   await page.getByRole("button", { name: "Approve" }).click()
  69  |   await expect.poll(() => approveCalled).toBe(true)
  70  | })
  71  | 
  72  | test("POD lifecycle critical flow (@critical)", async ({ page }) => {
  73  |   await stubAuthBootstrap(page)
  74  |   await page.goto("/buyer/orders")
  75  |   await expect(page).toHaveURL(/\/buyer\/orders$/)
  76  | })
  77  | 
  78  | test("seller ship order critical flow (@critical)", async ({ page }) => {
  79  |   await stubAuthBootstrap(page)
  80  |   await page.goto("/seller/orders")
  81  |   await expect(page).toHaveURL(/\/seller\/orders$/)
  82  | })
  83  | 
  84  | test("buyer confirm received critical flow (@critical)", async ({ page }) => {
  85  |   await stubAuthBootstrap(page)
  86  |   await page.goto("/buyer/orders")
  87  |   await expect(page).toHaveURL(/\/buyer\/orders$/)
  88  | })
  89  | 
  90  | test("review submission critical flow (@critical)", async ({ page }) => {
  91  |   await stubAuthBootstrap(page)
  92  |   await page.goto("/buyer/reviews")
  93  |   await expect(page).toHaveURL(/\/buyer\/reviews$/)
  94  | })
  95  | 
  96  | test("admin moderation critical flow (@critical)", async ({ page }) => {
  97  |   await stubAuthBootstrap(page)
  98  |   let moderationCalled = false
  99  |   await page.route("**/api/v1/admin/moderation/reports**", async (route) => {
  100 |     moderationCalled = true
  101 |     await mockJson(route, envelope({ items: [] }))
  102 |   })
  103 |   await page.goto("/admin/moderation")
  104 |   await expect.poll(() => moderationCalled).toBe(true)
  105 | })
  106 | 
```