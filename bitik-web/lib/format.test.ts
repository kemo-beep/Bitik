import { describe, expect, it, beforeEach } from "vitest"
import {
  formatMoney,
  formatNumber,
  formatDate,
  truncate,
  initials,
  setDefaultFormatLocale,
  getDefaultFormatLocale,
} from "./format"

beforeEach(() => {
  setDefaultFormatLocale("en-US")
})

describe("formatMoney", () => {
  it("formats MMK amounts", () => {
    const s = formatMoney(10, { locale: "en-US", currency: "MMK" })
    expect(s).toMatch(/10/)
  })
  it("returns em dash for non-finite", () => {
    expect(formatMoney(Number.NaN, { currency: "MMK" })).toBe("—")
  })
})

describe("formatNumber", () => {
  it("formats integers", () => {
    expect(formatNumber(1234, { locale: "en-US" })).toBe("1,234")
  })
})

describe("formatDate", () => {
  it("formats valid ISO date", () => {
    const s = formatDate("2024-06-01T12:00:00.000Z", { locale: "en-US" })
    expect(s).not.toBe("—")
    expect(s.length).toBeGreaterThan(4)
  })
  it("returns em dash for invalid", () => {
    expect(formatDate("not-a-date", { locale: "en-US" })).toBe("—")
  })
})

describe("truncate", () => {
  it("truncates long strings", () => {
    expect(truncate("hello world", 6).endsWith("…")).toBe(true)
  })
})

describe("initials", () => {
  it("takes up to two word initials", () => {
    expect(initials("Ada Lovelace")).toBe("AL")
  })
})

describe("setDefaultFormatLocale", () => {
  it("updates default locale for formatters", () => {
    setDefaultFormatLocale("de-DE")
    expect(getDefaultFormatLocale()).toBe("de-DE")
  })
})
