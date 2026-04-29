"use client"

import { ErrorState } from "@/components/ui/error-state"

export default function StorefrontError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <div className="mx-auto max-w-screen-md px-4 py-10">
      <ErrorState
        title="Storefront unavailable"
        description={error.message || "Could not render storefront content."}
        onRetry={reset}
      />
    </div>
  )
}

