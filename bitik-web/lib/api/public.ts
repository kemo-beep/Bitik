import { env } from "@/lib/env"
import { bitikFetch } from "@/lib/api/bitik-fetch"
import { parseEnvelope } from "@/lib/api/envelope"

export type PageParams = {
  page?: number
  per_page?: number
}

export type ProductSort =
  | "newest"
  | "price_asc"
  | "price_desc"
  | "popular"
  | "rating"

export type ProductListParams = PageParams & {
  q?: string
  category_id?: string
  brand_id?: string
  seller_id?: string
  min_price_cents?: number
  max_price_cents?: number
  sort?: ProductSort
}

export type PaginationMeta = {
  page?: number
  per_page?: number
  total?: number
}

export type PublicListResponse<T> = {
  items?: T[]
  pagination?: PaginationMeta
}

function withQuery(path: string, params?: Record<string, string | number | undefined>) {
  const url = new URL(`${env.apiBaseUrl}${path}`)
  for (const [k, v] of Object.entries(params ?? {})) {
    if (v === undefined || v === null || `${v}`.trim() === "") continue
    url.searchParams.set(k, String(v))
  }
  return url.toString()
}

export async function publicHome(): Promise<Record<string, unknown>> {
  const res = await bitikFetch(`${env.apiBaseUrl}/public/home`, { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function publicHomeSections(): Promise<Record<string, unknown>> {
  const res = await bitikFetch(`${env.apiBaseUrl}/public/home/sections`, { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function publicBanners(args: { placement?: string; page?: number; per_page?: number } = {}) {
  const res = await bitikFetch(
    withQuery("/public/banners", args),
    { method: "GET" },
    { skipAuth: true }
  )
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function listCategories(): Promise<Array<Record<string, unknown>>> {
  const res = await bitikFetch(`${env.apiBaseUrl}/public/categories`, { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Array<Record<string, unknown>>>(res)
  return data ?? []
}

export async function getCategory(categoryId: string): Promise<Record<string, unknown>> {
  const res = await bitikFetch(`${env.apiBaseUrl}/public/categories/${categoryId}`, { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function listCategoryProducts(categoryId: string, params: ProductListParams) {
  const res = await bitikFetch(
    withQuery(`/public/categories/${categoryId}/products`, params),
    { method: "GET" },
    { skipAuth: true }
  )
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function listBrands(params: PageParams = {}): Promise<Record<string, unknown>> {
  const res = await bitikFetch(withQuery("/public/brands", params), { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function getBrand(brandId: string): Promise<Record<string, unknown>> {
  const res = await bitikFetch(`${env.apiBaseUrl}/public/brands/${brandId}`, { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function listBrandProducts(brandId: string, params: ProductListParams) {
  const res = await bitikFetch(
    withQuery(`/public/brands/${brandId}/products`, params),
    { method: "GET" },
    { skipAuth: true }
  )
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function listProducts(params: ProductListParams) {
  const res = await bitikFetch(withQuery("/public/products", params), { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function getProduct(productId: string): Promise<Record<string, unknown>> {
  const res = await bitikFetch(`${env.apiBaseUrl}/public/products/${productId}`, { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function getProductBySlug(slug: string): Promise<Record<string, unknown>> {
  const res = await bitikFetch(`${env.apiBaseUrl}/public/products/slug/${encodeURIComponent(slug)}`, { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function listProductVariants(productId: string): Promise<Array<Record<string, unknown>>> {
  const res = await bitikFetch(`${env.apiBaseUrl}/public/products/${productId}/variants`, { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Array<Record<string, unknown>>>(res)
  return data ?? []
}

export async function listProductReviews(productId: string): Promise<Array<Record<string, unknown>>> {
  const res = await bitikFetch(`${env.apiBaseUrl}/public/products/${productId}/reviews`, { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Array<Record<string, unknown>>>(res)
  return data ?? []
}

export async function listRelatedProducts(productId: string): Promise<Array<Record<string, unknown>>> {
  const res = await bitikFetch(`${env.apiBaseUrl}/public/products/${productId}/related`, { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Array<Record<string, unknown>>>(res)
  return data ?? []
}

export async function getSeller(sellerId: string): Promise<Record<string, unknown>> {
  const res = await bitikFetch(`${env.apiBaseUrl}/public/sellers/${sellerId}`, { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function getSellerBySlug(slug: string): Promise<Record<string, unknown>> {
  const res = await bitikFetch(`${env.apiBaseUrl}/public/sellers/slug/${encodeURIComponent(slug)}`, { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function listSellerProducts(sellerId: string, params: ProductListParams) {
  const res = await bitikFetch(
    withQuery(`/public/sellers/${sellerId}/products`, params),
    { method: "GET" },
    { skipAuth: true }
  )
  const { data } = await parseEnvelope<Record<string, unknown>>(res)
  return data ?? {}
}

export async function listSellerReviews(sellerId: string): Promise<Array<Record<string, unknown>>> {
  const res = await bitikFetch(`${env.apiBaseUrl}/public/sellers/${sellerId}/reviews`, { method: "GET" }, { skipAuth: true })
  const { data } = await parseEnvelope<Array<Record<string, unknown>>>(res)
  return data ?? []
}

