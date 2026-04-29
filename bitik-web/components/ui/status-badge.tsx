import * as React from "react"
import { Badge } from "@/components/ui/badge"
import { statusMeta, type AnyStatusKind, type StatusTone } from "@/lib/status"
import { cn } from "@/lib/utils"

const TONE_TO_VARIANT: Record<StatusTone, React.ComponentProps<typeof Badge>["variant"]> = {
  neutral: "muted",
  info: "info",
  success: "success",
  warning: "warning",
  danger: "destructive",
  pending: "pending",
}

export function StatusBadge({
  kind,
  value,
  className,
}: {
  kind: AnyStatusKind
  value: string
  className?: string
}) {
  const meta = statusMeta(kind, value)
  return (
    <Badge variant={TONE_TO_VARIANT[meta.tone]} className={cn(className)}>
      {meta.label}
    </Badge>
  )
}
