import type { MetadataRoute } from "next"
import { env } from "@/lib/env"

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: "/",
        disallow: ["/account", "/seller", "/admin", "/auth"],
      },
    ],
    sitemap: `${env.oauthRedirectBaseUrl}/sitemap.xml`,
  }
}

