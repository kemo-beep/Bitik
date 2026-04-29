"use client"

import { useQuery } from "@tanstack/react-query"
import { getMe } from "@/lib/api/auth"
import { Card } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"

export default function Page() {
  const me = useQuery({ queryKey: ["me"], queryFn: getMe })

  return (
    <div className="space-y-4">
      <div className="space-y-1">
        <h1 className="text-lg font-semibold tracking-tight">Account</h1>
        <p className="text-sm text-muted-foreground">Your account overview.</p>
      </div>

      <Card className="p-4">
        {me.isLoading ? (
          <p className="text-sm text-muted-foreground">Loading…</p>
        ) : me.isError ? (
          <p className="text-sm text-destructive">Failed to load account.</p>
        ) : (
          <div className="space-y-2">
            <div className="flex items-center justify-between gap-4">
              <div className="min-w-0">
                <p className="text-sm font-medium truncate">{me.data?.email ?? "—"}</p>
                <p className="text-xs text-muted-foreground truncate">{me.data?.id ?? "—"}</p>
              </div>
              <Badge variant="muted">{me.data?.status ?? "unknown"}</Badge>
            </div>
            <div className="grid gap-2 md:grid-cols-2">
              <div className="text-sm">
                <span className="text-muted-foreground">Email verified:</span>{" "}
                {me.data?.email_verified ? "Yes" : "No"}
              </div>
              <div className="text-sm">
                <span className="text-muted-foreground">Phone verified:</span>{" "}
                {me.data?.phone_verified ? "Yes" : "No"}
              </div>
            </div>
          </div>
        )}
      </Card>
    </div>
  )
}
