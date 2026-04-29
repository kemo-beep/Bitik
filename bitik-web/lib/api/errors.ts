export type APIFieldError = {
  field?: string
  message?: string
}

export type APIErrorPayload = {
  code?: string
  message?: string
  fields?: APIFieldError[]
}

export class BitikAPIError extends Error {
  readonly status: number
  readonly code?: string
  readonly fields?: APIFieldError[]
  readonly traceId?: string
  readonly requestId?: string

  constructor(args: {
    status: number
    message: string
    code?: string
    fields?: APIFieldError[]
    traceId?: string
    requestId?: string
  }) {
    super(args.message)
    this.name = "BitikAPIError"
    this.status = args.status
    this.code = args.code
    this.fields = args.fields
    this.traceId = args.traceId
    this.requestId = args.requestId
  }
}

