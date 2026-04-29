import type { Page, Request as PlaywrightRequest, Route } from "@playwright/test"

/** Playwright host globs miss some cross-origin fetches (e.g. web on :3000 to API on :8080). */
export function routeApiMatch(pathContains: string) {
  return (url: URL) =>
    url.pathname.includes(pathContains) || url.href.includes(pathContains)
}

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

function headerFrom(req: PlaywrightRequest, name: string) {
  const want = name.toLowerCase()
  for (const [k, v] of Object.entries(req.headers())) {
    if (k.toLowerCase() === want) return v
  }
  return undefined
}

function corsHeadersForPlaywrightRequest(req: PlaywrightRequest) {
  const origin = headerFrom(req, "origin") ?? "http://127.0.0.1:3000"
  const allowHeaders =
    headerFrom(req, "access-control-request-headers") ??
    "Content-Type, X-Request-ID, X-Device-Id, X-Platform, Authorization, Idempotency-Key"
  return {
    "Access-Control-Allow-Origin": origin,
    "Access-Control-Allow-Credentials": "true",
    "Access-Control-Allow-Methods": "GET, POST, PUT, PATCH, DELETE, OPTIONS",
    "Access-Control-Allow-Headers": allowHeaders,
  } as Record<string, string>
}

/** Fulfill JSON for API mocks; handles CORS preflight for cross-origin browser fetches. */
export async function mockJson(route: Route, body: unknown, status = 200) {
  const req = route.request()
  if (req.method() === "OPTIONS") {
    await route.fulfill({
      status: 204,
      headers: corsHeadersForPlaywrightRequest(req),
    })
    return
  }
  await route.fulfill({
    status,
    contentType: "application/json",
    body: JSON.stringify(body),
    headers: {
      "X-Request-ID": "test-request-id",
      ...corsHeadersForPlaywrightRequest(req),
    },
  })
}

/**
 * Mocks auth refresh (401) for guest bootstrap. Uses `page.route` with OPTIONS CORS so
 * cross-origin browser fetches to the API host complete (preflight must not get a JSON 401).
 */
export async function stubAuthBootstrap(page: Page) {
  await page.route(
    (url) => url.pathname.includes("auth/refresh-token"),
    async (route) => {
      const req = route.request()
      if (req.method() === "OPTIONS") {
        await route.fulfill({
          status: 204,
          headers: corsHeadersForPlaywrightRequest(req),
        })
        return
      }
      await route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({
          success: false,
          error: { code: "unauthorized", message: "no session" },
        }),
        headers: { "X-Request-ID": "test", ...corsHeadersForPlaywrightRequest(req) },
      })
    }
  )
  await page.goto("about:blank")
}

export async function stubLoggedInSession(
  page: Page,
  roles: string[] = ["buyer"],
  me: { id?: string; email?: string; sub?: string } = {}
) {
  const sub = me.sub ?? "user-1"
  const accessToken = makeFakeJWT({ roles, sub })
  const user = {
    id: me.id ?? "00000000-0000-0000-0000-000000000001",
    status: "active",
    email_verified: true,
    phone_verified: true,
    created_at: new Date().toISOString(),
    email: me.email ?? "user@example.com",
  }
  await page.route(
    (url) => url.pathname.includes("auth/refresh-token"),
    async (route) => {
      const req = route.request()
      if (req.method() === "OPTIONS") {
        await route.fulfill({
          status: 204,
          headers: corsHeadersForPlaywrightRequest(req),
        })
        return
      }
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          success: true,
          data: { access_token: accessToken, refresh_token: "r1" },
        }),
        headers: { "X-Request-ID": "test", ...corsHeadersForPlaywrightRequest(req) },
      })
    }
  )
  await page.route(
    (url) => url.pathname.endsWith("/users/me"),
    async (route) => {
      const req = route.request()
      if (req.method() === "OPTIONS") {
        await route.fulfill({
          status: 204,
          headers: corsHeadersForPlaywrightRequest(req),
        })
        return
      }
      if (req.method() !== "GET") {
        await route.continue()
        return
      }
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ success: true, data: user }),
        headers: { "X-Request-ID": "test", ...corsHeadersForPlaywrightRequest(req) },
      })
    }
  )

  await page.goto("about:blank")
}
