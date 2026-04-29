#!/usr/bin/env node
/**
 * Fails Docker/CI builds when required NEXT_PUBLIC_* vars are missing or invalid.
 * Mirrors bitik-web/lib/env.ts expectations (URLs must be absolute).
 */
import { z } from "zod"

const Schema = z.object({
  NEXT_PUBLIC_API_BASE_URL: z.string().url(),
  NEXT_PUBLIC_ASSET_BASE_URL: z.string().url(),
  NEXT_PUBLIC_WS_BASE_URL: z.string().url(),
  NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL: z.string().url(),
})

const parsed = Schema.safeParse({
  NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL,
  NEXT_PUBLIC_ASSET_BASE_URL: process.env.NEXT_PUBLIC_ASSET_BASE_URL,
  NEXT_PUBLIC_WS_BASE_URL: process.env.NEXT_PUBLIC_WS_BASE_URL,
  NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL: process.env.NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL,
})

if (!parsed.success) {
  console.error("Invalid or missing NEXT_PUBLIC_* build environment:", parsed.error.flatten().fieldErrors)
  process.exit(1)
}

console.log("web build env OK")
