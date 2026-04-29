"use client"

import { useReportWebVitals } from "next/web-vitals"
import type { Metric } from "web-vitals"
import { env } from "@/lib/env"
import { useAnalytics } from "@/lib/analytics"

type NextWebVitalsMetric = Metric & {
  rating?: "good" | "needs-improvement" | "poor"
  navigationType?: string
}

export function WebVitals() {
  const analytics = useAnalytics()

  useReportWebVitals((metric) => {
    const m = metric as NextWebVitalsMetric
    analytics.track({
      name: "web_vitals",
      properties: {
        name: m.name,
        value: m.value,
        rating: m.rating,
        id: m.id,
        delta: m.delta,
        navigationType: m.navigationType,
      },
    })

    if (env.sentryDsn) {
      import("@sentry/nextjs").then((Sentry) => {
        Sentry.addBreadcrumb({
          category: "web-vitals",
          message: m.name,
          data: { ...m },
          level: "info",
        })
      })
    }
  })

  return null
}

