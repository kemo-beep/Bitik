"use client"

import { useEffect, useMemo, useState } from "react"
import { env } from "@/lib/env"
import { getAccessToken, setAccessToken } from "@/lib/auth/tokens"
import { refreshFromCookie } from "@/lib/api/auth"
import { RealtimeWsClient, type RealtimeMessage } from "@/lib/realtime/ws-client"

async function tokenProvider() {
  const existing = getAccessToken()
  if (existing) return existing
  try {
    const pair = await refreshFromCookie()
    setAccessToken(pair.access_token)
    return pair.access_token
  } catch {
    return null
  }
}

export function useRealtimeChannel(
  channel: "chat" | "notifications" | "order-status",
  onMessage: (event: RealtimeMessage) => void,
  enabled = true
) {
  const [status, setStatus] = useState<"connecting" | "open" | "closed">("closed")

  const client = useMemo(() => {
    const path = channel === "order-status" ? "/ws/order-status" : `/ws/${channel}`
    return new RealtimeWsClient(`${env.wsBaseUrl}${path}`, tokenProvider, onMessage, setStatus)
  }, [channel, onMessage])

  useEffect(() => {
    if (!enabled) return
    void client.connect()
    return () => client.disconnect()
  }, [client, enabled])

  return { status, connected: status === "open" }
}

