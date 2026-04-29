import createClient from "openapi-fetch"
import type { paths } from "./generated"
import { env } from "@/lib/env"
import { bitikFetch } from "./bitik-fetch"

export const api = createClient<paths>({
  baseUrl: env.apiBaseUrl,
  fetch: (input: RequestInfo | URL, init?: RequestInit) => bitikFetch(input, init),
})

