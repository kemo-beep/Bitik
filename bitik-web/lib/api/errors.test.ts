import { describe, expect, it } from "vitest"
import { BitikAPIError } from "./errors"

describe("BitikAPIError", () => {
  it("carries status, code, fields", () => {
    const err = new BitikAPIError({
      status: 422,
      message: "Invalid",
      code: "validation_error",
      fields: [{ field: "email", message: "bad" }],
      traceId: "t1",
      requestId: "r1",
    })
    expect(err.status).toBe(422)
    expect(err.code).toBe("validation_error")
    expect(err.fields).toEqual([{ field: "email", message: "bad" }])
    expect(err.traceId).toBe("t1")
    expect(err.requestId).toBe("r1")
    expect(err.message).toBe("Invalid")
  })
})
