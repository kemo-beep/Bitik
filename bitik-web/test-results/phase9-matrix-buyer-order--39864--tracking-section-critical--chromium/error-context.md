# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: phase9-matrix.spec.ts >> buyer order detail shows tracking section (@critical)
- Location: tests/phase9-matrix.spec.ts:144:1

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: getByRole('heading', { name: 'Order detail' })
Expected: visible
Timeout: 15000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 15000ms
  - waiting for getByRole('heading', { name: 'Order detail' })

```

# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - link "Skip to main content" [ref=e2] [cursor=pointer]:
    - /url: "#main-content"
  - generic [ref=e3]:
    - banner [ref=e4]:
      - link "Bitik home" [ref=e6] [cursor=pointer]:
        - /url: /
        - text: Bitik
    - main [ref=e7]:
      - generic [ref=e9]:
        - generic [ref=e10]:
          - heading "Welcome back" [level=1] [ref=e11]
          - paragraph [ref=e12]: Sign in to manage your account, orders, and seller tools.
        - generic [ref=e13]:
          - generic [ref=e14]:
            - group [ref=e15]:
              - generic [ref=e16]: Email
              - textbox "Email" [ref=e18]:
                - /placeholder: you@example.com
            - group [ref=e19]:
              - generic [ref=e20]: Password
              - textbox "Password" [ref=e22]:
                - /placeholder: Your password
          - button "Sign in" [ref=e23]
          - generic [ref=e24]:
            - link "Forgot password?" [ref=e25] [cursor=pointer]:
              - /url: /forgot-password
            - link "Create account" [ref=e26] [cursor=pointer]:
              - /url: /register
        - generic [ref=e27]:
          - paragraph [ref=e28]: Or continue with
          - generic [ref=e29]:
            - button "Google" [ref=e30]
            - button "Facebook" [ref=e31]
            - button "Apple" [ref=e32]
  - region "Notifications alt+T"
  - alert [ref=e34]: Bitik
```

# Test source

```ts
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
  141 |   await expect(page.getByText("ord-pending-1")).toBeVisible()
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
> 163 |   await expect(page.getByRole("heading", { name: "Order detail" })).toBeVisible({ timeout: 15_000 })
      |                                                                     ^ Error: expect(locator).toBeVisible() failed
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