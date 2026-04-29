"use client"

import { useQuery } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { queryKeys } from "@/lib/queryKeys"
import { SellerJsonView, SellerSectionCard } from "@/components/seller/section-card"

export function SellerDashboardClient() {
  const dashboard = useQuery({ queryKey: queryKeys.seller.dashboard(), queryFn: sellerApi.getDashboard })

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">Seller dashboard</h1>
      <SellerSectionCard title="Overview" description="Stats, chart, top products, recent orders, low stock.">
        {dashboard.isLoading ? <div className="text-sm">Loading dashboard…</div> : <SellerJsonView value={dashboard.data} />}
      </SellerSectionCard>
    </div>
  )
}

