"use client"

import { useQuery } from "@tanstack/react-query"
import { Badge } from "@/components/ui/badge"
import { getBuyerNotificationsUnreadCount } from "@/lib/api/buyer"
import { sellerApi } from "@/lib/api/seller"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"
import { asNumber } from "@/lib/safe"

export function UnreadIndicator({ scope }: { scope: "buyer" | "seller" | "admin" }) {
  const query = useQuery({
    queryKey:
      scope === "buyer"
        ? queryKeys.buyer.notificationsUnread()
        : scope === "seller"
          ? queryKeys.seller.notificationsUnread()
          : queryKeys.admin.notificationsUnread(),
    queryFn: async () => {
      if (scope === "buyer") return getBuyerNotificationsUnreadCount()
      if (scope === "seller") return sellerApi.getNotificationsUnread()
      return adminApi.getNotificationsUnread()
    },
    refetchInterval: 15_000,
  })

  const unread = asNumber((query.data as Record<string, unknown> | undefined)?.unread ?? 0) ?? 0
  if (unread <= 0) return null
  return <Badge variant="destructive">{unread} pending</Badge>
}

