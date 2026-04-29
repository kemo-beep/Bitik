import { z } from "zod"

const truthy = new Set(["1", "true", "yes", "on"])

function parseBool(value: string | undefined, defaultValue: boolean) {
  if (value == null || value.trim() === "") return defaultValue
  return truthy.has(value.trim().toLowerCase())
}

function parseJson<T>(raw: string | undefined, fallback: T): T {
  if (!raw || raw.trim() === "") return fallback
  try {
    return JSON.parse(raw) as T
  } catch {
    return fallback
  }
}

const FeatureFlagsSchema = z.record(z.string(), z.boolean())

const PublicEnvSchema = z.object({
  NEXT_PUBLIC_API_BASE_URL: z
    .string()
    .url()
    .default("http://localhost:8080/api/v1"),
  NEXT_PUBLIC_ASSET_BASE_URL: z
    .string()
    .url()
    .optional()
    .default("http://localhost:8080"),
  NEXT_PUBLIC_SENTRY_DSN: z.string().optional().default(""),
  NEXT_PUBLIC_WS_BASE_URL: z.string().url().optional().default("ws://localhost:8081"),
  NEXT_PUBLIC_FEATURE_FLAGS: z.string().optional().default(""),
  NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL: z
    .string()
    .url()
    .optional()
    .default("http://localhost:3000"),
  NEXT_PUBLIC_ANALYTICS_ENABLED: z.string().optional().default("0"),
})

export type FeatureFlags = z.infer<typeof FeatureFlagsSchema>

export const env = (() => {
  const parsed = PublicEnvSchema.parse({
    NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL,
    NEXT_PUBLIC_ASSET_BASE_URL: process.env.NEXT_PUBLIC_ASSET_BASE_URL,
    NEXT_PUBLIC_SENTRY_DSN: process.env.NEXT_PUBLIC_SENTRY_DSN,
    NEXT_PUBLIC_WS_BASE_URL: process.env.NEXT_PUBLIC_WS_BASE_URL,
    NEXT_PUBLIC_FEATURE_FLAGS: process.env.NEXT_PUBLIC_FEATURE_FLAGS,
    NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL:
      process.env.NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL,
    NEXT_PUBLIC_ANALYTICS_ENABLED: process.env.NEXT_PUBLIC_ANALYTICS_ENABLED,
  })

  const featureFlags = FeatureFlagsSchema.catch({}).parse(
    parseJson(parsed.NEXT_PUBLIC_FEATURE_FLAGS, {})
  )

  return {
    apiBaseUrl: parsed.NEXT_PUBLIC_API_BASE_URL.replace(/\/+$/, ""),
    assetBaseUrl: parsed.NEXT_PUBLIC_ASSET_BASE_URL.replace(/\/+$/, ""),
    wsBaseUrl: parsed.NEXT_PUBLIC_WS_BASE_URL.replace(/\/+$/, ""),
    sentryDsn: parsed.NEXT_PUBLIC_SENTRY_DSN,
    oauthRedirectBaseUrl: parsed.NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL.replace(
      /\/+$/,
      ""
    ),
    analyticsEnabled: parseBool(parsed.NEXT_PUBLIC_ANALYTICS_ENABLED, false),
    featureFlags,
  }
})()

