"use client"

import { ErrorState } from "@/components/ui/error-state"

export default function AccountError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <div className="mx-auto max-w-screen-md px-4 py-10">
      <ErrorState title="Account page error" description={error.message} onRetry={reset} />
    </div>
  )
}

