"use client"

import { useMutation, useQuery } from "@tanstack/react-query"
import { Card } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { FormError } from "@/components/shared/form-error"
import { listDevices, listSessions, revokeDevice, revokeSession } from "@/lib/api/account"

export default function Page() {
  const sessions = useQuery({ queryKey: ["sessions"], queryFn: listSessions })
  const devices = useQuery({ queryKey: ["devices"], queryFn: listDevices })

  const revokeSessionMutation = useMutation({
    mutationFn: revokeSession,
    onSuccess: () => void sessions.refetch(),
  })
  const revokeDeviceMutation = useMutation({
    mutationFn: revokeDevice,
    onSuccess: () => void devices.refetch(),
  })

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-lg font-semibold tracking-tight">Sessions and devices</h1>
        <p className="text-sm text-muted-foreground">
          Revoke access from sessions or devices you don’t recognize.
        </p>
      </div>

      <Card className="p-4 space-y-3">
        <h2 className="text-sm font-medium">Sessions</h2>
        <FormError error={revokeSessionMutation.error} title="Revoke failed" />
        {sessions.isLoading ? (
          <p className="text-sm text-muted-foreground">Loading…</p>
        ) : sessions.isError ? (
          <p className="text-sm text-destructive">Failed to load sessions.</p>
        ) : (sessions.data ?? []).length === 0 ? (
          <p className="text-sm text-muted-foreground">No sessions.</p>
        ) : (
          <div className="space-y-2">
            {(sessions.data ?? []).map((s) => (
              <div key={s.id} className="flex items-center justify-between gap-4">
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate">{s.platform ?? "unknown"}</p>
                  <p className="text-xs text-muted-foreground truncate">
                    {s.id} · device {s.device_id ?? "—"}
                  </p>
                </div>
                <Button
                  variant="destructive"
                  size="sm"
                  disabled={revokeSessionMutation.isPending}
                  onClick={() => revokeSessionMutation.mutate(s.id)}
                >
                  Revoke
                </Button>
              </div>
            ))}
          </div>
        )}
      </Card>

      <Card className="p-4 space-y-3">
        <h2 className="text-sm font-medium">Devices</h2>
        <FormError error={revokeDeviceMutation.error} title="Revoke failed" />
        {devices.isLoading ? (
          <p className="text-sm text-muted-foreground">Loading…</p>
        ) : devices.isError ? (
          <p className="text-sm text-destructive">Failed to load devices.</p>
        ) : (devices.data ?? []).length === 0 ? (
          <p className="text-sm text-muted-foreground">No devices.</p>
        ) : (
          <div className="space-y-2">
            {(devices.data ?? []).map((d) => (
              <div key={d.id} className="flex items-center justify-between gap-4">
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate">{d.platform}</p>
                  <p className="text-xs text-muted-foreground truncate">{d.device_id}</p>
                </div>
                <Button
                  variant="destructive"
                  size="sm"
                  disabled={revokeDeviceMutation.isPending}
                  onClick={() => revokeDeviceMutation.mutate(d.device_id)}
                >
                  Revoke
                </Button>
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  )
}
