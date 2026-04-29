import { describe, expect, it } from "vitest"
import { decodeJWT } from "./jwt"

describe("decodeJWT", () => {
  it("returns null on invalid token", () => {
    expect(decodeJWT("nope")).toBeNull()
  })
})

