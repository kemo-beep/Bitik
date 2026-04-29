import { BitikAPIError, type APIErrorPayload } from "./errors"

export type Envelope<T> = {
  success: boolean
  data?: T | null
  error?: APIErrorPayload | null
  meta?: Record<string, unknown> | null
  trace_id?: string
}

export async function parseEnvelope<T>(
  res: Response
): Promise<{ data: T; requestId?: string; traceId?: string }> {
  const requestId = res.headers.get("X-Request-ID") ?? undefined

  const contentType = res.headers.get("content-type") || ""
  if (!contentType.includes("application/json")) {
    if (!res.ok) {
      throw new BitikAPIError({
        status: res.status,
        message: res.statusText || "Request failed.",
        requestId,
      })
    }
    throw new BitikAPIError({
      status: res.status,
      message: "Unexpected non-JSON response.",
      requestId,
    })
  }

  const json: unknown = await res.json()
  if (
    typeof json !== "object" ||
    json === null ||
    !("success" in json) ||
    typeof (json as { success?: unknown }).success !== "boolean"
  ) {
    throw new BitikAPIError({
      status: res.status,
      message: "Invalid API response shape.",
      requestId,
    })
  }

  const env = json as Envelope<T>
  const traceId = env.trace_id

  if (!res.ok || !env.success) {
    const message =
      env.error?.message || res.statusText || "Request failed."
    throw new BitikAPIError({
      status: res.status,
      message,
      code: env.error?.code,
      fields: env.error?.fields,
      traceId,
      requestId,
    })
  }

  return { data: (env.data ?? null) as T, requestId, traceId }
}

