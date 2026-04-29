"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"


export default function Page() {
  return (
    <AdminResourceClient
      title="Payments"
      description="Payment operations and webhook event management."
      queryKey={queryKeys.admin.webhookEvents()}
      queryFn={() => adminApi.listWebhookEvents()}
      actions={[
        { label: "Approve Wave", placeholder: "Payment ID", action: (id) => adminApi.approveWavePayment(id) },
        { label: "Reject Wave", placeholder: "Payment ID", action: (id) => adminApi.rejectWavePayment(id) },
        { label: "Webhook detail", placeholder: "Webhook Event ID", action: (id) => adminApi.getWebhookEvent(id) },
        { label: "Webhook reprocess", placeholder: "Webhook Event ID", action: (id) => adminApi.reprocessWebhookEvent(id) },
      ]}
    />
  )
}
