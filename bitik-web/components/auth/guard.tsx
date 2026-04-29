"use client"

import * as React from "react"
import { usePathname, useRouter } from "next/navigation"
import { useAuth } from "@/lib/auth/auth-context"
import { routes } from "@/lib/routes"
import type { PermissionKey, UserRole } from "@/lib/roles"
import { AREA_ACCESS, canAccess, hasPermission } from "@/lib/roles"
import { PagePlaceholder } from "@/components/shared/page-placeholder"

function GuardFallback({ title }: { title: string }) {
  return (
    <PagePlaceholder
      title={title}
      description="Checking your session…"
      phase="Phase 1"
      notes={[]}
    />
  )
}

export function GuestOnly({ children }: { children: React.ReactNode }) {
  const { status } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  React.useEffect(() => {
    if (status === "authenticated") {
      router.replace(routes.storefront.home)
    }
  }, [pathname, router, status])

  if (status === "loading") return <GuardFallback title="Loading" />
  if (status === "authenticated") return <GuardFallback title="Redirecting" />
  return <>{children}</>
}

export function RequireAreaAccess({
  area,
  children,
}: {
  area: keyof typeof AREA_ACCESS
  children: React.ReactNode
}) {
  const { status, user } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  React.useEffect(() => {
    if (status === "guest") {
      router.replace(routes.auth.login)
    }
  }, [pathname, router, status])

  if (status === "loading") return <GuardFallback title="Loading" />
  if (status === "guest") return <GuardFallback title="Redirecting" />

  const role = user?.role ?? "guest"
  if (!canAccess(area, role)) {
    router.replace(routes.storefront.home)
    return <GuardFallback title="Redirecting" />
  }

  return <>{children}</>
}

export function RequireRole({
  role,
  children,
}: {
  role: UserRole
  children: React.ReactNode
}) {
  const { status, user } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  React.useEffect(() => {
    if (status === "guest") {
      router.replace(routes.auth.login)
    }
  }, [pathname, router, status])

  if (status === "loading") return <GuardFallback title="Loading" />
  if (status === "guest") return <GuardFallback title="Redirecting" />

  if ((user?.role ?? "guest") !== role) {
    router.replace(routes.storefront.home)
    return <GuardFallback title="Redirecting" />
  }

  return <>{children}</>
}

export function RequireAnyRole({
  anyOf,
  children,
}: {
  anyOf: UserRole[]
  children: React.ReactNode
}) {
  const { status, user } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  React.useEffect(() => {
    if (status === "guest") {
      router.replace(routes.auth.login)
    }
  }, [pathname, router, status])

  if (status === "loading") return <GuardFallback title="Loading" />
  if (status === "guest") return <GuardFallback title="Redirecting" />

  if (!anyOf.includes(user?.role ?? "guest")) {
    router.replace(routes.storefront.home)
    return <GuardFallback title="Redirecting" />
  }

  return <>{children}</>
}

export function RequirePermissions({
  anyOf,
  children,
}: {
  anyOf: PermissionKey[]
  children: React.ReactNode
}) {
  const { status, user } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  React.useEffect(() => {
    if (status === "guest") {
      router.replace(routes.auth.login)
    }
  }, [pathname, router, status])

  if (status === "loading") return <GuardFallback title="Loading" />
  if (status === "guest") return <GuardFallback title="Redirecting" />

  const ok = anyOf.some((k) => hasPermission(user, k))
  if (!ok) {
    router.replace(routes.storefront.home)
    return <GuardFallback title="Redirecting" />
  }

  return <>{children}</>
}

