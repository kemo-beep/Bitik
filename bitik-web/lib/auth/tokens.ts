const DEVICE_KEY = "bitik.device_id"

let memoryAccessToken: string | null = null
let memoryDeviceId: string | null = null

function safeRandomId() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

export function getDeviceId(): string {
  if (typeof window === "undefined") {
    if (!memoryDeviceId) memoryDeviceId = safeRandomId()
    return memoryDeviceId
  }

  const existing = window.localStorage.getItem(DEVICE_KEY)
  if (existing) return existing
  const created = safeRandomId()
  window.localStorage.setItem(DEVICE_KEY, created)
  return created
}

export function getAccessToken(): string | null {
  return memoryAccessToken
}

export function setAccessToken(token: string | null) {
  memoryAccessToken = token
}

export function clearAccessToken() {
  memoryAccessToken = null
}

