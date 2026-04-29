import { describe, expect, it } from "vitest"
import type { paths } from "./generated"

/** Stable paths we rely on across the web app; breaks if OpenAPI contract drifts. */
const REQUIRED_PATHS: (keyof paths)[] = [
  "/health",
  "/api/v1/auth/register",
  "/api/v1/auth/login",
  "/api/v1/buyer/orders",
  "/api/v1/analytics/events",
]

describe("generated OpenAPI paths contract", () => {
  it("includes required path keys", () => {
    type Keys = keyof paths
    for (const p of REQUIRED_PATHS) {
      expect(p as Keys).toBeDefined()
    }
  })
})
