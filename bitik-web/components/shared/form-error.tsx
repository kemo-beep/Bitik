"use client"

import * as React from "react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { BitikAPIError } from "@/lib/api/errors"

export function FormError({
  error,
  title = "Something went wrong",
}: {
  error: unknown
  title?: string
}) {
  if (!error) return null

  if (error instanceof BitikAPIError) {
    return (
      <Alert variant="destructive">
        <AlertTitle>{title}</AlertTitle>
        <AlertDescription>
          <div className="space-y-2">
            <p>{error.message}</p>
            {(error.fields?.length ?? 0) > 0 && (
              <ul className="ml-4 list-disc space-y-1">
                {error.fields?.map((f, idx) => (
                  <li key={idx}>
                    {f.field ? <span className="font-medium">{f.field}: </span> : null}
                    {f.message ?? "Invalid value"}
                  </li>
                ))}
              </ul>
            )}
            {(error.requestId || error.traceId) && (
              <p className="text-xs text-muted-foreground">
                {error.requestId ? `request_id=${error.requestId}` : null}
                {error.requestId && error.traceId ? " · " : null}
                {error.traceId ? `trace_id=${error.traceId}` : null}
              </p>
            )}
          </div>
        </AlertDescription>
      </Alert>
    )
  }

  const message =
    error instanceof Error ? error.message : "Unexpected error. Please try again."
  return (
    <Alert variant="destructive">
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription>{message}</AlertDescription>
    </Alert>
  )
}

