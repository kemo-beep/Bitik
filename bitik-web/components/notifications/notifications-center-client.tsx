"use client"

import { useMemo } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  deleteBuyerNotification,
  getBuyerNotificationsUnreadCount,
  listBuyerNotifications,
  markAllBuyerNotificationsRead,
  markBuyerNotificationRead,
} from "@/lib/api/buyer"
import { sellerApi } from "@/lib/api/seller"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"
import { asArray, asNumber, asRecord, asString } from "@/lib/safe"
import { Button } from "@/components/ui/button"
import { useRealtimeChannel } from "@/lib/realtime/use-realtime-channel"

export function NotificationsCenterClient({ scope }: { scope: "buyer" | "seller" | "admin" }) {
  const qc = useQueryClient()
  const qk =
    scope === "buyer" ? queryKeys.buyer.notifications() : scope === "seller" ? queryKeys.seller.notifications() : queryKeys.admin.notifications()
  const unreadQk =
    scope === "buyer"
      ? queryKeys.buyer.notificationsUnread()
      : scope === "seller"
        ? queryKeys.seller.notificationsUnread()
        : queryKeys.admin.notificationsUnread()

  const list = useQuery({
    queryKey: qk,
    queryFn: () => {
      if (scope === "buyer") return listBuyerNotifications()
      if (scope === "seller") return sellerApi.listNotifications()
      return adminApi.listNotifications()
    },
    refetchInterval: 15_000,
  })

  const unread = useQuery({
    queryKey: unreadQk,
    queryFn: () => {
      if (scope === "buyer") return getBuyerNotificationsUnreadCount()
      if (scope === "seller") return sellerApi.getNotificationsUnread()
      return adminApi.getNotificationsUnread()
    },
    refetchInterval: 15_000,
  })

  useRealtimeChannel(
    "notifications",
    () => {
      qc.invalidateQueries({ queryKey: qk })
      qc.invalidateQueries({ queryKey: unreadQk })
    },
    true
  )

  const markAll = useMutation({
    mutationFn: async () => {
      if (scope === "buyer") return markAllBuyerNotificationsRead()
      if (scope === "seller") return sellerApi.markAllNotificationsRead()
      return adminApi.markAllNotificationsRead()
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk })
      qc.invalidateQueries({ queryKey: unreadQk })
    },
  })

  const items = useMemo(() => asArray((list.data as Record<string, unknown> | undefined)?.items ?? list.data) ?? [], [list.data])
  const unreadCount = asNumber((unread.data as Record<string, unknown> | undefined)?.unread ?? 0) ?? 0

  return (
    <div className="mx-auto max-w-screen-lg px-4 py-8">
      <h1 className="font-heading text-2xl font-semibold">Notifications</h1>
      <p className="mt-1 text-sm text-muted-foreground">Unread: {unreadCount}</p>
      <div className="mt-3">
        <Button variant="outline" onClick={() => markAll.mutate()} disabled={markAll.isPending}>
          Mark all read
        </Button>
      </div>
      <div className="mt-4 space-y-3">
        {items.map((item, idx) => {
          const rec = asRecord(item) ?? {}
          const id = asString(rec.id) ?? `${idx}`
          return (
            <div key={id} className="rounded-xl border p-4">
              <div className="font-medium">{asString(rec.title) ?? "Notification"}</div>
              <div className="text-sm text-muted-foreground">{asString(rec.body) ?? ""}</div>
              <div className="mt-2 flex gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={async () => {
                    if (scope === "buyer") await markBuyerNotificationRead(id)
                    else if (scope === "seller") await sellerApi.markNotificationRead(id)
                    else await adminApi.markNotificationRead(id)
                    qc.invalidateQueries({ queryKey: qk })
                    qc.invalidateQueries({ queryKey: unreadQk })
                  }}
                >
                  Mark read
                </Button>
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={async () => {
                    if (scope === "buyer") await deleteBuyerNotification(id)
                    else if (scope === "seller") await sellerApi.deleteNotification(id)
                    else await adminApi.deleteNotification(id)
                    qc.invalidateQueries({ queryKey: qk })
                    qc.invalidateQueries({ queryKey: unreadQk })
                  }}
                >
                  Delete
                </Button>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

