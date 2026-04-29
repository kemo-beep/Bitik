import { env } from "@/lib/env"
import { bitikFetch } from "./bitik-fetch"
import { parseEnvelope } from "./envelope"

export type AuthPair = {
  access_token: string
  refresh_token: string
}

export type PublicUser = {
  id: string
  status: string
  email_verified: boolean
  phone_verified: boolean
  created_at: string
  email?: string
  phone?: string
}

export async function login(args: { email: string; password: string }) {
  const res = await bitikFetch(`${env.apiBaseUrl}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(args),
  }, { skipAuth: true, skipRefreshRetry: true })
  const { data } = await parseEnvelope<AuthPair>(res)
  return data
}

export async function register(args: {
  email: string
  password: string
  display_name?: string
}) {
  const res = await bitikFetch(`${env.apiBaseUrl}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(args),
  }, { skipAuth: true, skipRefreshRetry: true })
  const { data } = await parseEnvelope<AuthPair>(res)
  return data
}

export async function logout(args?: { refreshToken?: string }) {
  const res = await bitikFetch(
    `${env.apiBaseUrl}/auth/logout`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(
        args?.refreshToken ? { refresh_token: args.refreshToken } : {}
      ),
    },
    { skipRefreshRetry: true }
  )
  await parseEnvelope<unknown>(res)
}

export async function refreshFromCookie() {
  const res = await bitikFetch(
    `${env.apiBaseUrl}/auth/refresh-token`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: "{}",
    },
    { skipAuth: true, skipRefreshRetry: true }
  )
  const { data } = await parseEnvelope<AuthPair>(res)
  return data
}

export async function getMe() {
  const res = await bitikFetch(`${env.apiBaseUrl}/users/me`, {
    method: "GET",
  })
  const { data } = await parseEnvelope<PublicUser>(res)
  return data
}

export async function forgotPassword(args: { email: string }) {
  const res = await bitikFetch(
    `${env.apiBaseUrl}/auth/forgot-password`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(args),
    },
    { skipAuth: true, skipRefreshRetry: true }
  )
  await parseEnvelope<unknown>(res)
}

export async function resetPassword(args: { token: string; new_password: string }) {
  const res = await bitikFetch(
    `${env.apiBaseUrl}/auth/reset-password`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(args),
    },
    { skipAuth: true, skipRefreshRetry: true }
  )
  await parseEnvelope<unknown>(res)
}

export async function verifyEmail(args: { token: string }) {
  const res = await bitikFetch(
    `${env.apiBaseUrl}/auth/verify-email`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(args),
    },
    { skipAuth: true, skipRefreshRetry: true }
  )
  await parseEnvelope<unknown>(res)
}

export async function resendEmailVerification() {
  const res = await bitikFetch(`${env.apiBaseUrl}/auth/resend-email-verification`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: "{}",
  })
  await parseEnvelope<unknown>(res)
}

export async function sendPhoneOtp(args: { phone: string }) {
  const res = await bitikFetch(`${env.apiBaseUrl}/auth/send-phone-otp`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(args),
  })
  await parseEnvelope<unknown>(res)
}

export async function verifyPhoneOtp(args: { phone: string; otp: string }) {
  const res = await bitikFetch(`${env.apiBaseUrl}/auth/verify-phone-otp`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(args),
  })
  await parseEnvelope<unknown>(res)
}

