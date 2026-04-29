"use client"

import * as React from "react"
import { useParams, useRouter, useSearchParams } from "next/navigation"
import { useMutation } from "@tanstack/react-query"

import { PagePlaceholder } from "@/components/shared/page-placeholder"
import { FormError } from "@/components/shared/form-error"
import { env } from "@/lib/env"
import { parseEnvelope } from "@/lib/api/envelope"
import { bitikFetch } from "@/lib/api/bitik-fetch"
import { setAccessToken } from "@/lib/auth/tokens"
import { routes } from "@/lib/routes"

type AuthPair = { access_token: string; refresh_token: string }

export default function Page() {
  const router = useRouter()
  const params = useParams<{ provider: string }>()
  const search = useSearchParams()

  const provider = (params?.provider ?? "").toLowerCase()
  const code = search.get("code") ?? ""
  const state = search.get("state") ?? ""
  const oauthError = search.get("error") ?? ""

  const mutation = useMutation({
    mutationFn: async () => {
      if (!provider || !code || !state) {
        throw new Error("Missing OAuth parameters.")
      }

      const res = await bitikFetch(
        `${env.apiBaseUrl}/auth/oauth/${provider}/callback?code=${encodeURIComponent(
          code
        )}&state=${encodeURIComponent(state)}`,
        { method: "GET" },
        { skipAuth: true, skipRefreshRetry: true }
      )
      const { data } = await parseEnvelope<AuthPair>(res)

      // Backend sets HttpOnly refresh cookie; keep access token in memory only.
      setAccessToken(data.access_token)
    },
    onSuccess: () => {
      router.replace(routes.storefront.home)
    },
  })

  React.useEffect(() => {
    if (oauthError) return
    mutation.mutate()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [oauthError])

  if (oauthError) {
    return (
      <div className="mx-auto max-w-md p-6">
        <FormError error={new Error(oauthError)} title="OAuth failed" />
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-md p-6 space-y-4">
      <PagePlaceholder
        title="Signing you in…"
        description="Completing OAuth sign-in."
        phase="Phase 2"
        notes={[`provider=${provider}`]}
      />
      <FormError error={mutation.error} title="OAuth sign-in failed" />
    </div>
  )
}

