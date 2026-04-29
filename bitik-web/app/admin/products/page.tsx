"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"
import { useAnalytics, analyticsEvents } from "@/lib/analytics"


export default function Page() {
  const analytics = useAnalytics()
  return (
    <AdminResourceClient
      title="Products moderation"
      description="Moderate products from the admin product feed."
      queryKey={queryKeys.admin.products()}
      queryFn={() => adminApi.listProducts()}
      actions={[
        {
          label: "Approve product",
          placeholder: "Product ID",
          action: (id) => {
            analytics.track({ name: analyticsEvents.adminModerationAction, properties: { action: "approve", product_id: id } })
            return adminApi.patchProductModeration(id, { action: "approve" })
          },
        },
        {
          label: "Reject product",
          placeholder: "Product ID",
          action: (id) => {
            analytics.track({ name: analyticsEvents.adminModerationAction, properties: { action: "reject", product_id: id } })
            return adminApi.patchProductModeration(id, { action: "reject" })
          },
        },
        { label: "View detail", placeholder: "Product ID", action: (id) => adminApi.getProduct(id) },
      ]}
    />
  )
}
