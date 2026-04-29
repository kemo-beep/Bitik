"use client"

import * as React from "react"
import { env, type FeatureFlags } from "@/lib/env"

const FeatureFlagsContext = React.createContext<FeatureFlags>(env.featureFlags)

export function FeatureFlagsProvider({ children }: { children: React.ReactNode }) {
  return (
    <FeatureFlagsContext.Provider value={env.featureFlags}>
      {children}
    </FeatureFlagsContext.Provider>
  )
}

export function useFeatureFlags() {
  return React.useContext(FeatureFlagsContext)
}

