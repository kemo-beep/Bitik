import { env } from "@/lib/env"
import { bitikFetch } from "@/lib/api/bitik-fetch"
import { parseEnvelope } from "@/lib/api/envelope"

export type UserProfile = {
  user_id?: string
  first_name?: string
  last_name?: string
  display_name?: string
  avatar_url?: string
  language?: string
  country_code?: string
  timezone?: string
}

export type SessionItem = {
  id: string
  device_id?: string
  platform?: string
  created_at: string
  last_seen_at: string
  revoked: boolean
}

export type DeviceItem = {
  id: string
  device_id: string
  platform: string
  last_seen_at: string
  revoked: boolean
}

export type Address = {
  id: string
  full_name: string
  phone: string
  country: string
  state?: string | null
  city?: string | null
  district?: string | null
  postal_code?: string | null
  address_line1: string
  address_line2?: string | null
  is_default: boolean
  created_at?: string
  updated_at?: string
}

export type CreateAddressInput = {
  full_name: string
  phone: string
  country: string
  state?: string
  city?: string
  district?: string
  postal_code?: string
  address_line1: string
  address_line2?: string
  is_default: boolean
}

export async function getProfile(): Promise<UserProfile> {
  const res = await bitikFetch(`${env.apiBaseUrl}/users/me/profile`, { method: "GET" })
  const { data } = await parseEnvelope<UserProfile>(res)
  return data
}

export async function patchProfile(input: Partial<UserProfile>) {
  const res = await bitikFetch(`${env.apiBaseUrl}/users/me/profile`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  await parseEnvelope<unknown>(res)
}

export async function patchMe(input: { email?: string | null; phone?: string | null }) {
  const res = await bitikFetch(`${env.apiBaseUrl}/users/me`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<unknown>(res)
  return data
}

export async function listSessions(): Promise<SessionItem[]> {
  const res = await bitikFetch(`${env.apiBaseUrl}/users/me/sessions`, { method: "GET" })
  const { data } = await parseEnvelope<{ sessions: SessionItem[] }>(res)
  return data.sessions ?? []
}

export async function revokeSession(sessionId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/users/me/sessions/${sessionId}`, {
    method: "DELETE",
  })
  await parseEnvelope<unknown>(res)
}

export async function listDevices(): Promise<DeviceItem[]> {
  const res = await bitikFetch(`${env.apiBaseUrl}/users/me/devices`, { method: "GET" })
  const { data } = await parseEnvelope<{ devices: DeviceItem[] }>(res)
  return data.devices ?? []
}

export async function revokeDevice(deviceId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/users/me/devices/${deviceId}`, {
    method: "DELETE",
  })
  await parseEnvelope<unknown>(res)
}

export async function listAddresses(): Promise<Address[]> {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/addresses`, { method: "GET" })
  const { data } = await parseEnvelope<Address[]>(res)
  return data ?? []
}

export async function createAddress(input: CreateAddressInput) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/addresses`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<Address>(res)
  return data
}

export async function updateAddress(addressId: string, input: Partial<Address>) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/addresses/${addressId}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<Address>(res)
  return data
}

export async function deleteAddress(addressId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/addresses/${addressId}`, {
    method: "DELETE",
  })
  if (!res.ok && res.status !== 204) {
    await parseEnvelope<unknown>(res)
  }
}

export async function setDefaultAddress(addressId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/addresses/${addressId}/set-default`, {
    method: "POST",
  })
  const { data } = await parseEnvelope<Address>(res)
  return data
}

export async function deleteMe() {
  const res = await bitikFetch(`${env.apiBaseUrl}/users/me`, { method: "DELETE" })
  await parseEnvelope<unknown>(res)
}

