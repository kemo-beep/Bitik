import * as React from "react"
import { CircleAlertIcon, RefreshCwIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export function ErrorState({
  title = "Something went wrong",
  description = "Please try again. If this keeps happening, contact support.",
  onRetry,
  retryLabel = "Try again",
  className,
  children,
}: {
  title?: string
  description?: React.ReactNode
  onRetry?: () => void
  retryLabel?: string
  className?: string
  children?: React.ReactNode
}) {
  return (
    <div
      role="alert"
      aria-live="polite"
      className={cn(
        "flex w-full min-w-0 flex-col items-center justify-center gap-3 rounded-xl border border-dashed border-destructive/30 bg-destructive/5 p-6 text-center",
        className
      )}
    >
      <div className="flex size-9 items-center justify-center rounded-lg bg-destructive/10 text-destructive">
        <CircleAlertIcon className="size-5" aria-hidden="true" />
      </div>
      <div className="flex max-w-sm flex-col gap-1">
        <p className="text-sm font-medium">{title}</p>
        {description ? (
          <p className="text-sm text-muted-foreground">{description}</p>
        ) : null}
      </div>
      {children}
      {onRetry ? (
        <Button variant="outline" size="sm" onClick={onRetry}>
          <RefreshCwIcon data-icon="inline-start" />
          {retryLabel}
        </Button>
      ) : null}
    </div>
  )
}
