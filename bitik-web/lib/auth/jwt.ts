type JWTPayload = {
  sub?: string
  exp?: number
  iat?: number
  iss?: string
  roles?: string[]
}

function base64UrlDecode(input: string) {
  const pad = input.length % 4 ? "=".repeat(4 - (input.length % 4)) : ""
  const base64 = (input + pad).replace(/-/g, "+").replace(/_/g, "/")
  const bytes = atob(base64)
  const out = new Uint8Array(bytes.length)
  for (let i = 0; i < bytes.length; i++) out[i] = bytes.charCodeAt(i)
  return new TextDecoder().decode(out)
}

export function decodeJWT(token: string): JWTPayload | null {
  const parts = token.split(".")
  if (parts.length < 2) return null
  try {
    const json = base64UrlDecode(parts[1]!)
    const payload = JSON.parse(json) as JWTPayload
    return payload && typeof payload === "object" ? payload : null
  } catch {
    return null
  }
}

