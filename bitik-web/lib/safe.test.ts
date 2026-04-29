import { describe, expect, it } from "vitest"
import { asArray, asNumber, asRecord, asString } from "./safe"

describe("asRecord", () => {
  it("returns object for plain objects", () => {
    expect(asRecord({ a: 1 })).toEqual({ a: 1 })
  })
  it("returns null for arrays, null, primitives", () => {
    expect(asRecord([])).toBeNull()
    expect(asRecord(null)).toBeNull()
    expect(asRecord("x")).toBeNull()
  })
})

describe("asString", () => {
  it("returns string only for strings", () => {
    expect(asString("hi")).toBe("hi")
    expect(asString(1)).toBeNull()
  })
})

describe("asNumber", () => {
  it("returns finite numbers only", () => {
    expect(asNumber(3)).toBe(3)
    expect(asNumber(NaN)).toBeNull()
    expect(asNumber(Infinity)).toBeNull()
    expect(asNumber("1")).toBeNull()
  })
})

describe("asArray", () => {
  it("returns arrays only", () => {
    expect(asArray([1])).toEqual([1])
    expect(asArray({})).toBeNull()
  })
})
