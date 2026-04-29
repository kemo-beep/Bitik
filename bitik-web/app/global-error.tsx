"use client"

import { ErrorState } from "@/components/ui/error-state"

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <html>
      <body>
        <main className="mx-auto max-w-screen-md px-4 py-16">
          <ErrorState
            title="Application error"
            description={error.message || "Unexpected application failure."}
            onRetry={reset}
          />
        </main>
      </body>
    </html>
  )
}

