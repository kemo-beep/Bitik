"use client"

import * as React from "react"
import { usePathname } from "next/navigation"
import { RequireAnyRole } from "./guard"
import type { UserRole } from "@/lib/roles"

export function SellerRouteGuard({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()

  const applyPaths = ["/seller/apply", "/seller/application"]
  const isApplyFlow = applyPaths.some((p) => pathname === p || pathname.startsWith(`${p}/`))

  const anyOf: UserRole[] = isApplyFlow
    ? ["buyer", "seller_pending", "admin", "staff"]
    : ["seller_active", "seller_suspended", "admin", "staff"]

  return <RequireAnyRole anyOf={anyOf}>{children}</RequireAnyRole>
}

