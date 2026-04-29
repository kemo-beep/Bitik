"use client"

import * as React from "react"
import type { SessionUser, UserRole } from "@/lib/roles"
import { roleFromBackendRoles } from "@/lib/roles"
import { decodeJWT } from "@/lib/auth/jwt"
import {
  clearAccessToken,
  getAccessToken,
  setAccessToken as setMemoryAccessToken,
} from "@/lib/auth/tokens"
import * as authApi from "@/lib/api/auth"

type AuthStatus = "loading" | "authenticated" | "guest"

type AuthContextValue = {
  status: AuthStatus
  user: SessionUser | null
  roles: string[]
  accessToken: string | null
  signIn: (args: { email: string; password: string }) => Promise<void>
  register: (args: { email: string; password: string }) => Promise<void>
  signOut: () => Promise<void>
  hasRole: (role: UserRole) => boolean
}

const AuthContext = React.createContext<AuthContextValue | null>(null)

function rolesFromAccessToken(token: string | null): string[] {
  if (!token) return []
  const payload = decodeJWT(token)
  return Array.isArray(payload?.roles) ? payload!.roles! : []
}

function toSessionUser(me: authApi.PublicUser, roles: string[]): SessionUser {
  return {
    id: me.id,
    email: me.email,
    role: roleFromBackendRoles(roles),
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const initial = React.useMemo(() => {
    const accessToken = getAccessToken()
    const roles = rolesFromAccessToken(accessToken)
    const status: AuthStatus = "loading"
    return { accessToken, roles, status }
  }, [])

  const [status, setStatus] = React.useState<AuthStatus>(initial.status)
  const [accessToken, setAccessToken] = React.useState<string | null>(initial.accessToken)
  const [roles, setRoles] = React.useState<string[]>(initial.roles)
  const [user, setUser] = React.useState<SessionUser | null>(null)

  React.useEffect(() => {
    let cancelled = false
    setStatus("loading")

    ;(async () => {
      try {
        let token = accessToken
        if (!token) {
          const pair = await authApi.refreshFromCookie()
          token = pair.access_token
          setMemoryAccessToken(token)
          setAccessToken(token)
        }

        const r = rolesFromAccessToken(token)
        setRoles(r)

        const me = await authApi.getMe()
        if (cancelled) return
        setUser(toSessionUser(me, r))
        setStatus("authenticated")
      } catch {
        if (cancelled) return
        clearAccessToken()
        setAccessToken(null)
        setRoles([])
        setUser(null)
        setStatus("guest")
      }
    })()

    return () => {
      cancelled = true
    }
  }, [accessToken])

  const signIn = React.useCallback(async (args: { email: string; password: string }) => {
    const pair = await authApi.login(args)
    setMemoryAccessToken(pair.access_token)
    setAccessToken(pair.access_token)
  }, [])

  const register = React.useCallback(async (args: { email: string; password: string }) => {
    const pair = await authApi.register({ email: args.email, password: args.password })
    setMemoryAccessToken(pair.access_token)
    setAccessToken(pair.access_token)
  }, [])

  const signOut = React.useCallback(async () => {
    try {
      await authApi.logout()
    } finally {
      clearAccessToken()
      setMemoryAccessToken(null)
      setAccessToken(null)
      setRoles([])
      setUser(null)
      setStatus("guest")
    }
  }, [])

  const hasRole = React.useCallback(
    (role: UserRole) => (user?.role ?? "guest") === role,
    [user]
  )

  const value = React.useMemo<AuthContextValue>(
    () => ({
      status,
      user,
      roles,
      accessToken,
      signIn,
      register,
      signOut,
      hasRole,
    }),
    [accessToken, hasRole, register, roles, signIn, signOut, status, user]
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = React.useContext(AuthContext)
  if (!ctx) throw new Error("useAuth must be used within AuthProvider")
  return ctx
}

