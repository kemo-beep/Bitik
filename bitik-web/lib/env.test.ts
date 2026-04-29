import { describe, expect, it, vi, beforeEach, afterEach } from "vitest"

describe("env module", () => {
  const original = { ...process.env }

  beforeEach(() => {
    vi.resetModules()
  })

  afterEach(() => {
    process.env = { ...original }
  })

  it("enables analytics when NEXT_PUBLIC_ANALYTICS_ENABLED is truthy", async () => {
    process.env = { ...original, NEXT_PUBLIC_ANALYTICS_ENABLED: "1" }
    const { env } = await import("./env")
    expect(env.analyticsEnabled).toBe(true)
  })

  it("disables analytics by default when unset", async () => {
    process.env = { ...original }
    delete process.env.NEXT_PUBLIC_ANALYTICS_ENABLED
    const { env } = await import("./env")
    expect(env.analyticsEnabled).toBe(false)
  })

  it("strips trailing slashes from api base URL", async () => {
    process.env = {
      ...original,
      NEXT_PUBLIC_API_BASE_URL: "https://api.example.com/api/v1/",
    }
    const { env } = await import("./env")
    expect(env.apiBaseUrl).toBe("https://api.example.com/api/v1")
  })
})
