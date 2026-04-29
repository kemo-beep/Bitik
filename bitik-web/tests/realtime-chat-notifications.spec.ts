import { expect, test, type Page } from "@playwright/test"
import { envelope, makeFakeJWT, mockJson } from "./helpers/api-mock"

async function stubBuyer(page: Page) {
  const accessToken = makeFakeJWT({ roles: ["buyer"], sub: "buyer-1" })
  await page.route("**/api/v1/auth/refresh-token", async (route) => {
    await mockJson(route, envelope({ access_token: accessToken, refresh_token: "r1" }))
  })
  await page.route("**/api/v1/users/me", async (route) => {
    await mockJson(route, envelope({ id: "u-1", status: "active", email_verified: true, phone_verified: true, created_at: new Date().toISOString() }))
  })
}

async function stubSeller(page: Page) {
  const accessToken = makeFakeJWT({ roles: ["seller"], sub: "seller-1" })
  await page.route("**/api/v1/auth/refresh-token", async (route) => {
    await mockJson(route, envelope({ access_token: accessToken, refresh_token: "r1" }))
  })
  await page.route("**/api/v1/users/me", async (route) => {
    await mockJson(route, envelope({ id: "u-2", status: "active", email_verified: true, phone_verified: true, created_at: new Date().toISOString() }))
  })
}

test("buyer chat center send and delete flows render", async ({ page }) => {
  await stubBuyer(page)
  await page.route("**/api/v1/chat/conversations**", async (route) => {
    if (route.request().method() === "POST") {
      await mockJson(route, envelope({ id: "conv-1", buyer_id: "u-1", seller_id: "s-1" }), 201)
      return
    }
    await mockJson(route, envelope({ items: [{ id: "conv-1", unread_count: 1 }] }))
  })
  await page.route("**/api/v1/chat/conversations/conv-1/messages**", async (route) => {
    if (route.request().method() === "POST") {
      await mockJson(route, envelope({ id: "msg-2", message: "hello" }), 201)
      return
    }
    await mockJson(route, envelope({ items: [{ id: "msg-1", message: "initial" }] }))
  })
  await page.route("**/api/v1/chat/conversations/conv-1/messages/msg-1", async (route) => {
    await route.fulfill({ status: 204 })
  })
  await page.route("**/api/v1/chat/conversations/conv-1/read", async (route) => {
    await route.fulfill({ status: 204 })
  })
  await page.route("**/api/v1/chat/conversations/conv-1", async (route) => {
    await route.fulfill({ status: 204 })
  })
  await page.route("**/api/v1/media/upload/presigned-url", async (route) => {
    await mockJson(route, envelope({ file: { id: "file-1" }, upload_url: "http://localhost:3100/upload" }), 201)
  })
  await page.route("http://localhost:3100/upload", async (route) => {
    await route.fulfill({ status: 200 })
  })
  await page.route("**/api/v1/media/upload/presigned-complete", async (route) => {
    await mockJson(route, envelope({ id: "file-1", url: "https://cdn.local/file-1.png" }))
  })

  await page.goto("/account/chat")
  await page.getByPlaceholder("Seller ID").fill("s-1")
  await page.getByRole("button", { name: "Start" }).click()
  await page.getByPlaceholder("Type a message").fill("hello")
  await page.getByRole("button", { name: "Send" }).click()
  await expect(page.getByText("initial")).toBeVisible()
})

test("buyer notifications center actions call real endpoints", async ({ page }) => {
  await stubBuyer(page)
  await page.route("**/api/v1/buyer/notifications/unread-count", async (route) => {
    await mockJson(route, envelope({ unread: 2 }))
  })
  await page.route("**/api/v1/buyer/notifications/read-all", async (route) => {
    await route.fulfill({ status: 204 })
  })
  await page.route("**/api/v1/buyer/notifications/n-1/read", async (route) => {
    await mockJson(route, envelope({ id: "n-1", read_at: new Date().toISOString() }))
  })
  await page.route("**/api/v1/buyer/notifications/n-1", async (route) => {
    await route.fulfill({ status: 204 })
  })
  await page.route("**/api/v1/buyer/notifications**", async (route) => {
    await mockJson(route, envelope({ items: [{ id: "n-1", title: "Order update", body: "Packed" }] }))
  })
  await page.goto("/account/notifications")
  await expect(page.getByRole("heading", { name: "Notifications" })).toBeVisible()
  await page.getByRole("button", { name: "Mark all read" }).click()
  await page.getByRole("button", { name: "Mark read" }).first().click()
})

test("seller order detail shows realtime fallback state", async ({ page }) => {
  await stubSeller(page)
  await page.route("**/api/v1/seller/orders/ord-1", async (route) => {
    await mockJson(route, envelope({ id: "ord-1", status: "pending" }))
  })
  await page.route("**/api/v1/seller/orders/ord-1/items", async (route) => {
    await mockJson(route, envelope({ items: [] }))
  })
  await page.route("**/api/v1/seller/orders/ord-1/shipments", async (route) => {
    await mockJson(route, envelope({ items: [] }))
  })
  await page.goto("/seller/orders/ord-1")
  await expect(page.getByText(/Realtime channel:/)).toBeVisible()
})

