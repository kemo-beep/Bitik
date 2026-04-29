# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: phase9-matrix.spec.ts >> Wave manual pending page (@critical)
- Location: tests/phase9-matrix.spec.ts:137:1

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: getByText('ord-pending-1')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for getByText('ord-pending-1')

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
        - heading "Awaiting manual confirmation" [level=1] [ref=e34]
        - paragraph [ref=e35]: Your payment is pending review. This usually takes a short while.
        - link "View my orders" [ref=e37] [cursor=pointer]:
          - /url: /account/orders
    - contentinfo [ref=e38]:
      - generic [ref=e39]:
        - navigation "Shop" [ref=e40]:
          - heading "Shop" [level=2] [ref=e41]
          - list [ref=e42]:
            - listitem [ref=e43]:
              - link "Categories" [ref=e44] [cursor=pointer]:
                - /url: /categories
            - listitem [ref=e45]:
              - link "Brands" [ref=e46] [cursor=pointer]:
                - /url: /brands
            - listitem [ref=e47]:
              - link "All products" [ref=e48] [cursor=pointer]:
                - /url: /products
            - listitem [ref=e49]:
              - link "Search" [ref=e50] [cursor=pointer]:
                - /url: /search
        - navigation "Account" [ref=e51]:
          - heading "Account" [level=2] [ref=e52]
          - list [ref=e53]:
            - listitem [ref=e54]:
              - link "Sign in" [ref=e55] [cursor=pointer]:
                - /url: /login
            - listitem [ref=e56]:
              - link "Register" [ref=e57] [cursor=pointer]:
                - /url: /register
            - listitem [ref=e58]:
              - link "Orders" [ref=e59] [cursor=pointer]:
                - /url: /account/orders
            - listitem [ref=e60]:
              - link "Wishlist" [ref=e61] [cursor=pointer]:
                - /url: /wishlist
        - navigation "Sell on Bitik" [ref=e62]:
          - heading "Sell on Bitik" [level=2] [ref=e63]
          - list [ref=e64]:
            - listitem [ref=e65]:
              - link "Apply as seller" [ref=e66] [cursor=pointer]:
                - /url: /seller/apply
            - listitem [ref=e67]:
              - link "Seller center" [ref=e68] [cursor=pointer]:
                - /url: /seller
        - navigation "Help" [ref=e69]:
          - heading "Help" [level=2] [ref=e70]
          - list [ref=e71]:
            - listitem [ref=e72]:
              - link "FAQ" [ref=e73] [cursor=pointer]:
                - /url: /p/page/faq
            - listitem [ref=e74]:
              - link "Shipping" [ref=e75] [cursor=pointer]:
                - /url: /p/page/shipping
            - listitem [ref=e76]:
              - link "Returns" [ref=e77] [cursor=pointer]:
                - /url: /p/page/returns
            - listitem [ref=e78]:
              - link "Contact" [ref=e79] [cursor=pointer]:
                - /url: /p/page/contact
      - generic [ref=e81]:
        - paragraph [ref=e82]: © 2026 Bitik. All rights reserved.
        - generic [ref=e83]:
          - link "Terms" [ref=e84] [cursor=pointer]:
            - /url: /p/page/terms
          - link "Privacy" [ref=e85] [cursor=pointer]:
            - /url: /p/page/privacy
  - region "Notifications alt+T"
  - alert [ref=e87]
```

# Test source

```ts
  41  |   })
  42  | }
  43  | 
  44  | test("register submits auth register (@critical)", async ({ page }) => {
  45  |   await stubAuthBootstrap(page)
  46  |   const token = makeFakeJWT({ roles: ["buyer"], sub: "new-user" })
  47  |   let registered = false
  48  |   await page.route("**/api/v1/auth/register", async (route) => {
  49  |     registered = true
  50  |     await mockJson(route, envelope({ access_token: token, refresh_token: "r-new" }))
  51  |   })
  52  |   await page.route("**/api/v1/users/me", async (route) => {
  53  |     await mockJson(
  54  |       route,
  55  |       envelope({
  56  |         id: "00000000-0000-0000-0000-000000000222",
  57  |         status: "active",
  58  |         email_verified: false,
  59  |         phone_verified: false,
  60  |         created_at: new Date().toISOString(),
  61  |         email: "new@example.com",
  62  |       })
  63  |     )
  64  |   })
  65  | 
  66  |   await page.goto("/register")
  67  |   await page.getByLabel("Email").fill("new@example.com")
  68  |   await page.getByLabel("Password").fill("passwordpassword")
  69  |   await page.getByRole("button", { name: "Create account" }).click()
  70  |   await expect.poll(() => registered).toBe(true)
  71  |   await expect(page).toHaveURL("/")
  72  | })
  73  | 
  74  | test("storefront home and search load (@critical)", async ({ page }) => {
  75  |   await stubAuthBootstrap(page)
  76  |   await page.route("**/api/v1/public/home", async (route) => {
  77  |     await mockJson(route, envelope({ banners: [], sections: [], products: [] }))
  78  |   })
  79  |   await page.goto("/")
  80  |   await expect(page.getByRole("banner")).toBeVisible({ timeout: 15_000 })
  81  | 
  82  |   await page.route("**/api/v1/public/products**", async (route) => {
  83  |     await mockJson(route, envelope({ items: [], pagination: { total_pages: 1 } }))
  84  |   })
  85  |   await page.goto("/search?q=test")
  86  |   await expect(page.getByRole("heading", { name: "Search" })).toBeVisible()
  87  | })
  88  | 
  89  | test("checkout with POD pre-selected places order (@critical)", async ({ page }) => {
  90  |   await stubAuthBootstrap(page)
  91  |   await page.route("**/api/v1/buyer/addresses", async (route) => {
  92  |     await mockJson(
  93  |       route,
  94  |       envelope([
  95  |         {
  96  |           id: "addr-1",
  97  |           full_name: "Buyer",
  98  |           address_line1: "1 St",
  99  |           phone: "+1",
  100 |           country: "US",
  101 |           is_default: true,
  102 |         },
  103 |       ])
  104 |     )
  105 |   })
  106 |   await page.route("**/api/v1/buyer/checkout/sessions", async (route) => {
  107 |     await mockJson(route, envelope({ id: "chk-pod", checkout_session_id: "chk-pod" }), 201)
  108 |   })
  109 |   await page.route("**/api/v1/buyer/checkout/sessions/chk-pod", async (route) => {
  110 |     await mockJson(
  111 |       route,
  112 |       envelope({
  113 |         id: "chk-pod",
  114 |         shipping_address_id: "addr-1",
  115 |         shipping_method: "standard",
  116 |         payment_method: "pod",
  117 |         summary: { total_amount: 5000, currency: "MMK" },
  118 |       })
  119 |     )
  120 |   })
  121 |   let placed = false
  122 |   await page.route("**/api/v1/buyer/checkout/sessions/chk-pod/**", async (route) => {
  123 |     const url = route.request().url()
  124 |     if (url.endsWith("/place-order")) {
  125 |       placed = true
  126 |       await mockJson(route, envelope({ order_id: "ord-pod", payment_status: "paid" }), 201)
  127 |       return
  128 |     }
  129 |     await mockJson(route, envelope({ ok: true }))
  130 |   })
  131 | 
  132 |   await page.goto("/checkout")
  133 |   await page.getByRole("button", { name: "Place order" }).click()
  134 |   await expect.poll(() => placed).toBe(true)
  135 | })
  136 | 
  137 | test("Wave manual pending page (@critical)", async ({ page }) => {
  138 |   await stubAuthBootstrap(page)
  139 |   await page.goto("/checkout/pending?order_id=ord-pending-1")
  140 |   await expect(page.getByRole("heading", { name: "Awaiting manual confirmation" })).toBeVisible()
> 141 |   await expect(page.getByText("ord-pending-1")).toBeVisible()
      |                                                 ^ Error: expect(locator).toBeVisible() failed
  142 | })
  143 | 
  144 | test("buyer order detail shows tracking section (@critical)", async ({ page }) => {
  145 |   await stubAuthBootstrap(page)
  146 |   await page.route("**/api/v1/buyer/orders/ord-1", async (route) => {
  147 |     await mockJson(route, envelope({ id: "ord-1", status: "shipped" }))
  148 |   })
  149 |   await page.route("**/api/v1/buyer/orders/ord-1/items", async (route) => {
  150 |     await mockJson(route, envelope({ items: [] }))
  151 |   })
  152 |   await page.route("**/api/v1/buyer/orders/ord-1/status-history", async (route) => {
  153 |     await mockJson(route, envelope({ items: [] }))
  154 |   })
  155 |   await page.route("**/api/v1/buyer/orders/ord-1/tracking", async (route) => {
  156 |     await mockJson(route, envelope({ carrier: "TestCarrier", events: [] }))
  157 |   })
  158 |   await page.route("**/api/v1/buyer/orders/ord-1/invoice", async (route) => {
  159 |     await mockJson(route, envelope({}))
  160 |   })
  161 | 
  162 |   await page.goto("/account/orders/ord-1")
  163 |   await expect(page.getByRole("heading", { name: "Order detail" })).toBeVisible({ timeout: 15_000 })
  164 |   await expect(page.getByRole("heading", { name: "Tracking" })).toBeVisible()
  165 | })
  166 | 
  167 | test("seller creates product from new product page (@critical)", async ({ page }) => {
  168 |   await stubSellerBootstrap(page, ["seller"])
  169 |   let created = false
  170 |   await page.route("**/api/v1/seller/products", async (route) => {
  171 |     if (route.request().method() === "POST") {
  172 |       created = true
  173 |       await mockJson(route, envelope({ id: "prod-new", name: "Widget", slug: "widget" }), 201)
  174 |       return
  175 |     }
  176 |     await mockJson(route, envelope({ items: [] }))
  177 |   })
  178 | 
  179 |   await page.goto("/seller/products/new")
  180 |   await expect(page.getByRole("heading", { name: "New product" })).toBeVisible({ timeout: 15_000 })
  181 |   await page.getByPlaceholder("Name").fill("Widget")
  182 |   await page.getByPlaceholder("Slug").fill("widget")
  183 |   await page.getByRole("button", { name: "Create product" }).click()
  184 |   await expect.poll(() => created).toBe(true)
  185 | })
  186 | 
  187 | test("seller ships order from order detail (@critical)", async ({ page }) => {
  188 |   await stubSellerBootstrap(page, ["seller"])
  189 |   let shipped = false
  190 |   await page.route("**/api/v1/seller/orders/ord-s1**", async (route) => {
  191 |     const url = route.request().url()
  192 |     if (url.includes("/orders/ord-s1/ship") && route.request().method() === "POST") {
  193 |       shipped = true
  194 |       await mockJson(route, envelope({ ok: true }))
  195 |       return
  196 |     }
  197 |     if (url.includes("/shipments")) {
  198 |       await mockJson(route, envelope({ items: [] }))
  199 |       return
  200 |     }
  201 |     if (url.includes("/items")) {
  202 |       await mockJson(route, envelope({ items: [] }))
  203 |       return
  204 |     }
  205 |     await mockJson(route, envelope({ id: "ord-s1", status: "paid" }))
  206 |   })
  207 | 
  208 |   await page.goto("/seller/orders/ord-s1")
  209 |   await page.getByRole("button", { name: "Ship" }).first().click()
  210 |   await expect.poll(() => shipped).toBe(true)
  211 | })
  212 | 
  213 | test("admin approves seller application (@critical)", async ({ page }) => {
  214 |   await stubAdminSession(page)
  215 |   await page.route("**/api/v1/admin/seller-applications", async (route) => {
  216 |     await mockJson(route, envelope({ items: [{ id: "app-99", status: "pending" }] }))
  217 |   })
  218 |   let reviewed = false
  219 |   await page.route("**/api/v1/admin/seller-applications/app-99/review", async (route) => {
  220 |     reviewed = true
  221 |     await mockJson(route, envelope({ id: "app-99", decision: "approved" }))
  222 |   })
  223 | 
  224 |   await page.goto("/admin/sellers")
  225 |   await expect(page.getByRole("heading", { name: "Sellers" })).toBeVisible({ timeout: 15_000 })
  226 |   await page.getByPlaceholder("Application ID").first().fill("app-99")
  227 |   await page.getByRole("button", { name: "Approve application" }).click({ force: true })
  228 |   await expect.poll(() => reviewed).toBe(true)
  229 | })
  230 | 
```