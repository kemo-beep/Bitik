"use client"

import { useQuery } from "@tanstack/react-query"
import dynamic from "next/dynamic"
import { sellerApi } from "@/lib/api/seller"
import { queryKeys } from "@/lib/queryKeys"
import { SellerJsonView, SellerSectionCard } from "@/components/seller/section-card"
import { asArray, asNumber, asRecord } from "@/lib/safe"

const SellerSalesChart = dynamic(
  () => import("@/components/seller/seller-sales-chart").then((m) => m.SellerSalesChart),
  { ssr: false }
)

export function SellerAnalyticsClient() {
  const dashboard = useQuery({ queryKey: queryKeys.seller.dashboard(), queryFn: sellerApi.getDashboard })
  const rec = asRecord(dashboard.data) ?? {}
  const chartData = (asArray(rec.sales_chart) ?? []).map((item, idx) => {
    const row = asRecord(item) ?? {}
    return {
      label: String(row.day ?? row.label ?? idx + 1),
      value: asNumber(row.total_sales_cents ?? row.value ?? 0) ?? 0,
    }
  })

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">Analytics</h1>
      <SellerSectionCard
        title="Sales / orders / products / conversion snapshot"
        description="Uses seller dashboard analytics aggregates currently exposed by backend."
      >
        <SellerJsonView value={dashboard.data} />
      </SellerSectionCard>
      <SellerSectionCard title="Sales chart (lazy loaded)">
        <SellerSalesChart data={chartData} />
      </SellerSectionCard>
    </div>
  )
}

