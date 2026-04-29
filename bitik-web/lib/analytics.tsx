"use client"

import * as React from "react"
import { env } from "@/lib/env"
import { bitikFetch } from "@/lib/api/bitik-fetch"
import { analyticsEvents, type AnalyticsEventName } from "@/lib/analytics-events"

export type AnalyticsEvent = {
  name: AnalyticsEventName | string
  properties?: Record<string, unknown>
}

type AnalyticsContextValue = {
  enabled: boolean
  track: (event: AnalyticsEvent) => void
}

const AnalyticsContext = React.createContext<AnalyticsContextValue>({
  enabled: false,
  track: () => {},
})

export function AnalyticsProvider({ children }: { children: React.ReactNode }) {
  const value = React.useMemo<AnalyticsContextValue>(() => {
    const active = env.analyticsEnabled
    return {
      enabled: active,
      track: (event) => {
        if (!active) return
        void bitikFetch(`${env.apiBaseUrl}/analytics/events`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            event_name: event.name,
            metadata: event.properties ?? {},
          }),
        }).catch(() => {
          // best-effort telemetry
        })
        if (process.env.NODE_ENV !== "production") {
          console.info("[analytics]", event.name, event.properties ?? {})
        }
      },
    }
  }, [])

  return <AnalyticsContext.Provider value={value}>{children}</AnalyticsContext.Provider>
}

export function useAnalytics() {
  return React.useContext(AnalyticsContext)
}

export { analyticsEvents }

