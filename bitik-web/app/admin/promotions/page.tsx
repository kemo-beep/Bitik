"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return (
    <AdminResourceClient
      title="Promotions"
      description="Voucher lifecycle and campaign controls."
      queryKey={queryKeys.admin.vouchers()}
      queryFn={adminApi.listVouchers}
      actions={[
        { label: "Create voucher", placeholder: "Code", action: (code) => adminApi.createVoucher({ code, discount_percent: 10 }) },
        { label: "Delete voucher", placeholder: "Voucher ID", action: (id) => adminApi.deleteVoucher(id) },
      ]}
    />
  )
}
