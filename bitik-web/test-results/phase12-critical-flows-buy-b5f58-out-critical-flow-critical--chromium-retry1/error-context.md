# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: phase12-critical-flows.spec.ts >> buyer browse/search/cart/checkout critical flow (@critical)
- Location: tests/phase12-critical-flows.spec.ts:31:1

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: locator.click: Test timeout of 30000ms exceeded.
Call log:
  - waiting for getByRole('button', { name: 'Place order' })
    - locator resolved to <button disabled tabindex="0" type="button" data-disabled="" data-slot="button" class="group/button inline-flex shrink-0 items-center justify-center rounded-lg border border-transparent bg-clip-padding text-sm font-medium whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-inva…>Place order</button>
  - attempting click action
    2 × waiting for element to be visible, enabled and stable
      - element is not enabled
    - retrying click action
    - waiting 20ms
    2 × waiting for element to be visible, enabled and stable
      - element is not enabled
    - retrying click action
      - waiting 100ms
    59 × waiting for element to be visible, enabled and stable
       - element is not enabled
     - retrying click action
       - waiting 500ms

```

# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - link "Skip to main content" [ref=e2] [cursor=pointer]:
    - /url: "#main-content"
  - generic [ref=e3]:
    - banner [ref=e4]:
      - generic [ref=e5]:
        - link "Bitik home" [ref=e6] [cursor=pointer]:
          - /url: /
          - text: Bitik
        - navigation "Primary" [ref=e7]:
          - link "Categories" [ref=e8] [cursor=pointer]:
            - /url: /categories
          - link "Brands" [ref=e9] [cursor=pointer]:
            - /url: /brands
          - link "Products" [ref=e10] [cursor=pointer]:
            - /url: /products
        - search [ref=e11]:
          - text: Search Bitik
          - generic [ref=e12]:
            - img [ref=e13]
            - searchbox "Search Bitik" [ref=e16]
        - generic [ref=e17]:
          - generic [ref=e18]:
            - text: Language
            - combobox "Language" [ref=e19]:
              - option "EN" [selected]
              - option "AR"
          - link "Wishlist" [ref=e20] [cursor=pointer]:
            - /url: /wishlist
            - img [ref=e21]
          - link "Notifications" [ref=e23] [cursor=pointer]:
            - /url: /account/notifications
            - img [ref=e24]
          - link "Cart" [ref=e27] [cursor=pointer]:
            - /url: /cart
            - img [ref=e28]
          - button "Sign in" [ref=e32] [cursor=pointer]:
            - img [ref=e33]
            - text: Sign in
    - main [ref=e36]:
      - generic [ref=e37]:
        - heading "Checkout" [level=1] [ref=e38]
        - paragraph [ref=e39]: Review details, validate, and place order.
        - generic [ref=e40]:
          - heading "Address" [level=2] [ref=e42]
          - generic [ref=e43]:
            - heading "Shipping" [level=2] [ref=e44]
            - generic [ref=e45]:
              - button "standard" [ref=e46]
              - button "express" [ref=e47]
          - generic [ref=e48]:
            - heading "Payment method" [level=2] [ref=e49]
            - generic [ref=e50]:
              - button "Wave manual" [ref=e51]
              - button "POD" [ref=e52]
            - paragraph [ref=e53]: POD availability depends on backend eligibility checks.
          - generic [ref=e54]:
            - heading "Voucher" [level=2] [ref=e55]
            - generic [ref=e56]:
              - textbox "Voucher code" [ref=e57]
              - button "Apply" [disabled] [ref=e58]
        - generic [ref=e59]:
          - heading "Order summary" [level=2] [ref=e60]
          - generic [ref=e61]: "{}"
          - generic [ref=e62]:
            - button "Validate" [disabled] [ref=e63]
            - button "Place order" [disabled] [ref=e64]
    - contentinfo [ref=e65]:
      - generic [ref=e66]:
        - navigation "Shop" [ref=e67]:
          - heading "Shop" [level=2] [ref=e68]
          - list [ref=e69]:
            - listitem [ref=e70]:
              - link "Categories" [ref=e71] [cursor=pointer]:
                - /url: /categories
            - listitem [ref=e72]:
              - link "Brands" [ref=e73] [cursor=pointer]:
                - /url: /brands
            - listitem [ref=e74]:
              - link "All products" [ref=e75] [cursor=pointer]:
                - /url: /products
            - listitem [ref=e76]:
              - link "Search" [ref=e77] [cursor=pointer]:
                - /url: /search
        - navigation "Account" [ref=e78]:
          - heading "Account" [level=2] [ref=e79]
          - list [ref=e80]:
            - listitem [ref=e81]:
              - link "Sign in" [ref=e82] [cursor=pointer]:
                - /url: /login
            - listitem [ref=e83]:
              - link "Register" [ref=e84] [cursor=pointer]:
                - /url: /register
            - listitem [ref=e85]:
              - link "Orders" [ref=e86] [cursor=pointer]:
                - /url: /account/orders
            - listitem [ref=e87]:
              - link "Wishlist" [ref=e88] [cursor=pointer]:
                - /url: /wishlist
        - navigation "Sell on Bitik" [ref=e89]:
          - heading "Sell on Bitik" [level=2] [ref=e90]
          - list [ref=e91]:
            - listitem [ref=e92]:
              - link "Apply as seller" [ref=e93] [cursor=pointer]:
                - /url: /seller/apply
            - listitem [ref=e94]:
              - link "Seller center" [ref=e95] [cursor=pointer]:
                - /url: /seller
        - navigation "Help" [ref=e96]:
          - heading "Help" [level=2] [ref=e97]
          - list [ref=e98]:
            - listitem [ref=e99]:
              - link "FAQ" [ref=e100] [cursor=pointer]:
                - /url: /p/page/faq
            - listitem [ref=e101]:
              - link "Shipping" [ref=e102] [cursor=pointer]:
                - /url: /p/page/shipping
            - listitem [ref=e103]:
              - link "Returns" [ref=e104] [cursor=pointer]:
                - /url: /p/page/returns
            - listitem [ref=e105]:
              - link "Contact" [ref=e106] [cursor=pointer]:
                - /url: /p/page/contact
      - generic [ref=e108]:
        - paragraph [ref=e109]: © 2026 Bitik. All rights reserved.
        - generic [ref=e110]:
          - link "Terms" [ref=e111] [cursor=pointer]:
            - /url: /p/page/terms
          - link "Privacy" [ref=e112] [cursor=pointer]:
            - /url: /p/page/privacy
  - region "Notifications alt+T"
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
> 52  |   await page.getByRole("button", { name: "Place order" }).click()
      |                                                           ^ Error: locator.click: Test timeout of 30000ms exceeded.
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
  67  |   await page.getByPlaceholder("Payment ID").first().fill("pay-1")
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