import type { Metadata, Viewport } from "next"
import { Geist_Mono, Inter } from "next/font/google"

import "./globals.css"
import { ThemeProvider } from "@/components/theme-provider"
import { Toaster } from "@/components/ui/sonner"
import { cn } from "@/lib/utils"
import { A11Y } from "@/lib/a11y"
import { AppProviders } from "@/app/providers"
import { WebVitals } from "@/components/observability/web-vitals"
import { NetworkAndSessionBanner } from "@/components/system/network-and-session-banner"

const inter = Inter({ subsets: ["latin"], variable: "--font-sans" })

const fontMono = Geist_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
})

export const metadata: Metadata = {
  title: {
    default: "Bitik",
    template: "%s · Bitik",
  },
  description: "Bitik marketplace — discover, shop, sell.",
  applicationName: "Bitik",
}

export const viewport: Viewport = {
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "#ffffff" },
    { media: "(prefers-color-scheme: dark)", color: "#0a0a0a" },
  ],
  width: "device-width",
  initialScale: 1,
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html
      lang="en"
      suppressHydrationWarning
      className={cn(
        "antialiased",
        fontMono.variable,
        "font-sans",
        inter.variable
      )}
    >
      <body>
        <a
          href={`#${A11Y.skipLinkId}`}
          className="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-[100] focus:rounded-md focus:bg-background focus:px-3 focus:py-2 focus:text-sm focus:shadow-md focus:outline focus:outline-2 focus:outline-ring"
        >
          Skip to main content
        </a>
        <ThemeProvider>
          <AppProviders>{children}</AppProviders>
          <WebVitals />
          <Toaster />
          <NetworkAndSessionBanner />
          <div
            id={A11Y.liveRegionId}
            aria-live="polite"
            aria-atomic="true"
            className="sr-only"
          />
        </ThemeProvider>
      </body>
    </html>
  )
}
