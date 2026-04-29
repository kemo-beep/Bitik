"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return (
    <AdminResourceClient
      title="Sellers"
      description="Review seller applications and update seller status."
      queryKey={queryKeys.admin.sellerApplications()}
      queryFn={() => adminApi.listSellerApplications()}
      actions={[
        { label: "Approve application", placeholder: "Application ID", action: (id) => adminApi.reviewSellerApplication(id, { decision: "approved" }) },
        { label: "Suspend seller", placeholder: "Seller ID", action: (id) => adminApi.patchSellerStatus(id, { status: "suspended" }) },
      ]}
    />
  )
}
