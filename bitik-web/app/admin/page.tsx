"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return <AdminResourceClient title="Admin dashboard" description="Overview totals and event charts." queryKey={queryKeys.admin.dashboardOverview()} queryFn={adminApi.dashboardOverview} />
}
