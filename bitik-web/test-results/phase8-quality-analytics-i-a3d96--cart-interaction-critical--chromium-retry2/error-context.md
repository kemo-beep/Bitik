# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: phase8-quality.spec.ts >> analytics ingestion fires for add-to-cart interaction (@critical)
- Location: tests/phase8-quality.spec.ts:12:1

# Error details

```
Error: expect(received).toBeGreaterThan(expected)

Expected: > 0
Received:   0

Call Log:
- Timeout 15000ms exceeded while waiting on the predicate
```

# Page snapshot

```yaml
- generic [ref=e1]:
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
          - generic [ref=e12]: Search Bitik
          - generic [ref=e13]:
            - img
            - searchbox "Search Bitik" [ref=e14]
        - generic [ref=e15]:
          - generic [ref=e16]:
            - generic [ref=e17]: Language
            - combobox "Language" [ref=e18]:
              - option "EN" [selected]
              - option "AR"
          - link "Wishlist" [ref=e19] [cursor=pointer]:
            - /url: /wishlist
            - img [ref=e20]
          - link "Notifications" [ref=e22] [cursor=pointer]:
            - /url: /account/notifications
            - img [ref=e23]
          - link "Cart" [ref=e26] [cursor=pointer]:
            - /url: /cart
            - img [ref=e27]
          - button "Sign in" [ref=e31] [cursor=pointer]:
            - img
            - text: Sign in
    - main [ref=e32]:
      - generic [ref=e33]:
        - generic [ref=e34]:
          - img "No images" [ref=e35]
          - generic [ref=e36]:
            - generic [ref=e37]:
              - heading "Product A" [level=1] [ref=e38]
              - generic [ref=e39]: No rating · 0 reviews
            - generic [ref=e40]: MMK 10
            - generic [ref=e41]:
              - button "Add to cart (stub)" [active] [ref=e42]
              - button "Wishlist (stub)" [disabled]
            - generic [ref=e43]:
              - generic [ref=e45]: Seller
              - generic [ref=e46]:
                - generic [ref=e47]:
                  - generic [ref=e48]: Seller One
                  - generic [ref=e49]: Ships fast · Secure payments
                - button "Visit shop" [ref=e50] [cursor=pointer]
            - generic [ref=e51]:
              - heading "Shipping & payment" [level=2] [ref=e52]
              - list [ref=e53]:
                - listitem [ref=e54]: Delivery estimates shown at checkout (stub).
                - listitem [ref=e55]: Cashless payments supported (stub).
                - listitem [ref=e56]: Returns and refunds policy (stub).
        - generic [ref=e57]:
          - heading "Reviews" [level=2] [ref=e58]
          - generic [ref=e60]: No reviews yet.
        - generic [ref=e61]:
          - generic [ref=e62]:
            - heading "Related products" [level=2] [ref=e63]
            - link "Browse more" [ref=e64] [cursor=pointer]:
              - /url: /products
          - generic [ref=e66]: No related products.
    - contentinfo [ref=e67]:
      - generic [ref=e68]:
        - navigation "Shop" [ref=e69]:
          - heading "Shop" [level=2] [ref=e70]
          - list [ref=e71]:
            - listitem [ref=e72]:
              - link "Categories" [ref=e73] [cursor=pointer]:
                - /url: /categories
            - listitem [ref=e74]:
              - link "Brands" [ref=e75] [cursor=pointer]:
                - /url: /brands
            - listitem [ref=e76]:
              - link "All products" [ref=e77] [cursor=pointer]:
                - /url: /products
            - listitem [ref=e78]:
              - link "Search" [ref=e79] [cursor=pointer]:
                - /url: /search
        - navigation "Account" [ref=e80]:
          - heading "Account" [level=2] [ref=e81]
          - list [ref=e82]:
            - listitem [ref=e83]:
              - link "Sign in" [ref=e84] [cursor=pointer]:
                - /url: /login
            - listitem [ref=e85]:
              - link "Register" [ref=e86] [cursor=pointer]:
                - /url: /register
            - listitem [ref=e87]:
              - link "Orders" [ref=e88] [cursor=pointer]:
                - /url: /account/orders
            - listitem [ref=e89]:
              - link "Wishlist" [ref=e90] [cursor=pointer]:
                - /url: /wishlist
        - navigation "Sell on Bitik" [ref=e91]:
          - heading "Sell on Bitik" [level=2] [ref=e92]
          - list [ref=e93]:
            - listitem [ref=e94]:
              - link "Apply as seller" [ref=e95] [cursor=pointer]:
                - /url: /seller/apply
            - listitem [ref=e96]:
              - link "Seller center" [ref=e97] [cursor=pointer]:
                - /url: /seller
        - navigation "Help" [ref=e98]:
          - heading "Help" [level=2] [ref=e99]
          - list [ref=e100]:
            - listitem [ref=e101]:
              - link "FAQ" [ref=e102] [cursor=pointer]:
                - /url: /p/page/faq
            - listitem [ref=e103]:
              - link "Shipping" [ref=e104] [cursor=pointer]:
                - /url: /p/page/shipping
            - listitem [ref=e105]:
              - link "Returns" [ref=e106] [cursor=pointer]:
                - /url: /p/page/returns
            - listitem [ref=e107]:
              - link "Contact" [ref=e108] [cursor=pointer]:
                - /url: /p/page/contact
      - generic [ref=e110]:
        - paragraph [ref=e111]: © 2026 Bitik. All rights reserved.
        - generic [ref=e112]:
          - link "Terms" [ref=e113] [cursor=pointer]:
            - /url: /p/page/terms
          - link "Privacy" [ref=e114] [cursor=pointer]:
            - /url: /p/page/privacy
  - region "Notifications alt+T"
  - alert [ref=e116]
```

# Test source

```ts
  1  | import { expect, test } from "@playwright/test"
  2  | import { envelope, mockJson, stubAuthBootstrap } from "./helpers/api-mock"
  3  | 
  4  | test("language switcher toggles document direction", async ({ page }) => {
  5  |   await stubAuthBootstrap(page)
  6  |   await page.goto("/")
  7  |   await expect(page.locator("html")).toHaveAttribute("dir", "ltr")
  8  |   await page.getByLabel("Language").selectOption("ar")
  9  |   await expect(page.locator("html")).toHaveAttribute("dir", "rtl")
  10 | })
  11 | 
  12 | test("analytics ingestion fires for add-to-cart interaction (@critical)", async ({ page }) => {
  13 |   await stubAuthBootstrap(page)
  14 |   await page.route("**/api/v1/public/products/*", async (route) => {
  15 |     await mockJson(
  16 |       route,
  17 |       envelope({
  18 |         id: "p-1",
  19 |         name: "Product A",
  20 |         seller_id: "s-1",
  21 |         min_price_cents: 1000,
  22 |         max_price_cents: 1000,
  23 |         currency: "MMK",
  24 |         images: [],
  25 |       })
  26 |     )
  27 |   })
  28 |   await page.route("**/api/v1/public/products/*/reviews**", async (route) => {
  29 |     await mockJson(route, envelope({ items: [] }))
  30 |   })
  31 |   await page.route("**/api/v1/public/products/*/related**", async (route) => {
  32 |     await mockJson(route, envelope([]))
  33 |   })
  34 |   await page.route("**/api/v1/public/sellers/*", async (route) => {
  35 |     await mockJson(route, envelope({ id: "s-1", shop_name: "Seller One" }))
  36 |   })
  37 | 
  38 |   const events: unknown[] = []
  39 |   await page.route("**/api/v1/analytics/events", async (route) => {
  40 |     const payload = route.request().postDataJSON()
  41 |     events.push(payload)
  42 |     await mockJson(route, envelope({}))
  43 |   })
  44 | 
  45 |   await page.goto("/products/p-1")
  46 |   await page.getByRole("button", { name: "Add to cart (stub)" }).click()
> 47 |   await expect.poll(() => events.length, { timeout: 15_000 }).toBeGreaterThan(0)
     |   ^ Error: expect(received).toBeGreaterThan(expected)
  48 | })
  49 | 
  50 | 
```