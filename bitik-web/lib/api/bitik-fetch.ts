import { env } from "@/lib/env"
import { getDeviceId, getAccessToken, setAccessToken, clearAccessToken } from "@/lib/auth/tokens"
import { parseEnvelope } from "./envelope"

type RefreshResponse = { access_token: string; refresh_token: string }
const SESSION_EXPIRED_KEY = "bitik.session.expired.v1"

let refreshPromise: Promise<string | null> | null = null

function newRequestId() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

async function refreshAccessToken(): Promise<string | null> {
  if (refreshPromise) return refreshPromise

  refreshPromise = (async () => {
    const res = await fetch(`${env.apiBaseUrl}/auth/refresh-token`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Request-ID": newRequestId(),
        "X-Device-Id": getDeviceId(),
        "X-Platform": "web",
      },
      credentials: "include",
      body: "{}",
    })

    try {
      const { data } = await parseEnvelope<RefreshResponse>(res)
      if (!data?.access_token) return null
      setAccessToken(data.access_token)
      return data.access_token
    } catch {
      clearAccessToken()
      if (typeof window !== "undefined") window.localStorage.setItem(SESSION_EXPIRED_KEY, "1")
      return null
    } finally {
      refreshPromise = null
    }
  })()

  return refreshPromise
}

export type BitikFetchOptions = {
  idempotencyKey?: string
  skipAuth?: boolean
  skipRefreshRetry?: boolean
}

export async function bitikFetch(
  input: RequestInfo | URL,
  init?: RequestInit,
  opts: BitikFetchOptions = {}
): Promise<Response> {
  const headers = new Headers(init?.headers || {})

  if (!headers.has("X-Request-ID")) headers.set("X-Request-ID", newRequestId())
  if (!headers.has("X-Device-Id")) headers.set("X-Device-Id", getDeviceId())
  if (!headers.has("X-Platform")) headers.set("X-Platform", "web")
  if (opts.idempotencyKey && !headers.has("Idempotency-Key")) {
    headers.set("Idempotency-Key", opts.idempotencyKey)
  }

  if (!opts.skipAuth) {
    const token = getAccessToken()
    if (token && !headers.has("Authorization")) {
      headers.set("Authorization", `Bearer ${token}`)
    }
  }

  const res = await fetch(input, {
    ...init,
    headers,
    credentials: init?.credentials ?? "include",
  })

  if (res.status !== 401 || opts.skipRefreshRetry || opts.skipAuth) {
    return res
  }

  const refreshed = await refreshAccessToken()
  if (!refreshed) return res

  const retryHeaders = new Headers(headers)
  retryHeaders.set("Authorization", `Bearer ${refreshed}`)
  retryHeaders.set("X-Request-ID", newRequestId())

  return fetch(input, {
    ...init,
    headers: retryHeaders,
    credentials: init?.credentials ?? "include",
  })
}

