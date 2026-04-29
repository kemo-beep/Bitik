# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: phase9-matrix.spec.ts >> admin approves seller application (@critical)
- Location: tests/phase9-matrix.spec.ts:213:1

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: getByRole('heading', { name: 'Sellers' })
Expected: visible
Timeout: 15000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 15000ms
  - waiting for getByRole('heading', { name: 'Sellers' })

```

# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - link "Skip to main content" [ref=e2] [cursor=pointer]:
    - /url: "#main-content"
  - generic [ref=e3]:
    - banner [ref=e4]:
      - generic [ref=e5]:
        - heading "Loading" [level=1] [ref=e6]
        - text: Phase 1
      - paragraph [ref=e7]: Checking your session…
    - generic [ref=e9]:
      - generic [ref=e10]: Phase 0 stub
      - generic [ref=e11]: This page is part of the route map. Implementation arrives in later phases.
  - region "Notifications alt+T"
```

# Test source

```ts
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
> 225 |   await expect(page.getByRole("heading", { name: "Sellers" })).toBeVisible({ timeout: 15_000 })
      |                                                                ^ Error: expect(locator).toBeVisible() failed
  226 |   await page.getByPlaceholder("Application ID").first().fill("app-99")
  227 |   await page.getByRole("button", { name: "Approve application" }).click({ force: true })
  228 |   await expect.poll(() => reviewed).toBe(true)
  229 | })
  230 | 
```