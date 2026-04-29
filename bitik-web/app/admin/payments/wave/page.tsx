"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return (
    <AdminResourceClient
      title="Wave approvals"
      description="Approve/reject pending manual Wave payments."
      queryKey={queryKeys.admin.wavePending()}
      queryFn={adminApi.listPendingWavePayments}
      actions={[
        { label: "Approve", placeholder: "Payment ID", action: (id) => adminApi.approveWavePayment(id) },
        { label: "Reject", placeholder: "Payment ID", action: (id) => adminApi.rejectWavePayment(id) },
      ]}
    />
  )
}
