"use client"

import * as React from "react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { AuthProvider } from "@/lib/auth/auth-context"
import { FeatureFlagsProvider } from "@/lib/feature-flags"
import { AnalyticsProvider } from "@/lib/analytics"
import { I18nProvider } from "@/lib/i18n/context"

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: 1,
        refetchOnWindowFocus: false,
      },
    },
  })
}

export function AppProviders({ children }: { children: React.ReactNode }) {
  const [client] = React.useState(makeQueryClient)

  return (
    <QueryClientProvider client={client}>
      <I18nProvider>
        <FeatureFlagsProvider>
          <AnalyticsProvider>
            <AuthProvider>{children}</AuthProvider>
          </AnalyticsProvider>
        </FeatureFlagsProvider>
      </I18nProvider>
    </QueryClientProvider>
  )
}

