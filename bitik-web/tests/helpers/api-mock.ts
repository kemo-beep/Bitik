import type { Page, Route } from "@playwright/test"

export function makeFakeJWT(payload: Record<string, unknown>) {
  const header = { alg: "none", typ: "JWT" }
  const enc = (obj: unknown) =>
    Buffer.from(JSON.stringify(obj))
      .toString("base64")
      .replace(/=/g, "")
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
  return `${enc(header)}.${enc(payload)}.`
}

export function envelope<T>(data: T) {
  return { success: true, data }
}

export function errorEnvelope(code: string, message: string, status = 400) {
  return { status, body: { success: false, error: { code, message } } }
}

export async function mockJson(route: Route, body: unknown, status = 200) {
  await route.fulfill({
    status,
    contentType: "application/json",
    body: JSON.stringify(body),
    headers: { "X-Request-ID": "test-request-id" },
  })
}

export async function stubAuthBootstrap(page: Page) {
  await page.route("**/api/v1/auth/refresh-token", async (route) => {
    // Default: treat as logged-out session.
    await mockJson(route, { success: false, error: { code: "unauthorized", message: "no session" } }, 401)
  })
}

