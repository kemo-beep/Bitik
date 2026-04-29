"use client"

import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { useI18n } from "@/lib/i18n/context"

const SESSION_EXPIRED_KEY = "bitik.session.expired.v1"

export function NetworkAndSessionBanner() {
  const { t } = useI18n()
  const [online, setOnline] = useState(true)
  const [staleSession, setStaleSession] = useState(false)

  useEffect(() => {
    const sync = () => setOnline(typeof navigator === "undefined" ? true : navigator.onLine)
    sync()
    window.addEventListener("online", sync)
    window.addEventListener("offline", sync)
    setStaleSession(window.localStorage.getItem(SESSION_EXPIRED_KEY) === "1")
    return () => {
      window.removeEventListener("online", sync)
      window.removeEventListener("offline", sync)
    }
  }, [])

  if (online && !staleSession) return null

  return (
    <div className="fixed right-3 bottom-3 z-[120] max-w-sm space-y-2">
      {!online ? <div className="rounded border bg-amber-100 px-3 py-2 text-xs text-amber-900">{t("offline.message")}</div> : null}
      {staleSession ? (
        <div className="rounded border bg-destructive/10 px-3 py-2 text-xs">
          <p>{t("session.stale")}</p>
          <Button
            size="sm"
            className="mt-2"
            onClick={() => {
              window.localStorage.removeItem(SESSION_EXPIRED_KEY)
              setStaleSession(false)
              window.location.href = "/login"
            }}
          >
            Sign in
          </Button>
        </div>
      ) : null}
    </div>
  )
}

