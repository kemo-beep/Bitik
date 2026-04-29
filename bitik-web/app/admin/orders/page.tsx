"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return (
    <AdminResourceClient
      title="Orders"
      description="List/administer orders and refunds."
      queryKey={queryKeys.admin.orders()}
      queryFn={() => adminApi.listOrders()}
      actions={[
        { label: "Get order", placeholder: "Order ID", action: (id) => adminApi.getOrder(id) },
        { label: "Cancel order", placeholder: "Order ID", action: (id) => adminApi.cancelOrder(id) },
        { label: "Refund order", placeholder: "Order ID", action: (id) => adminApi.refundOrder(id) },
      ]}
    />
  )
}
