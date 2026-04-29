import { describe, expect, it } from "vitest"
import { parseEnvelope } from "./envelope"
import { BitikAPIError } from "./errors"

function jsonResponse(body: unknown, init: ResponseInit & { contentType?: string } = {}) {
  const headers = new Headers(init.headers)
  headers.set("content-type", init.contentType ?? "application/json")
  return new Response(JSON.stringify(body), { ...init, headers })
}

describe("parseEnvelope", () => {
  it("returns data on success envelope", async () => {
    const res = jsonResponse({ success: true, data: { id: "1" } })
    const { data } = await parseEnvelope<{ id: string }>(res)
    expect(data).toEqual({ id: "1" })
  })

  it("throws BitikAPIError on success false", async () => {
    const res = jsonResponse(
      { success: false, error: { code: "bad", message: "nope" } },
      { status: 400 }
    )
    await expect(parseEnvelope(res)).rejects.toSatisfy((e: unknown) => {
      return e instanceof BitikAPIError && e.message === "nope" && e.code === "bad"
    })
  })

  it("throws on invalid shape", async () => {
    const res = jsonResponse({ foo: 1 })
    await expect(parseEnvelope(res)).rejects.toThrow(BitikAPIError)
  })

  it("throws on non-JSON content type", async () => {
    const res = new Response("plain", { status: 500, headers: { "content-type": "text/plain" } })
    await expect(parseEnvelope(res)).rejects.toThrow(BitikAPIError)
  })
})
