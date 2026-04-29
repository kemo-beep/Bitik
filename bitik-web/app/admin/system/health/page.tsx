"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return <AdminResourceClient title="System health" description="Service health and readiness checks." queryKey={queryKeys.admin.health()} queryFn={adminApi.health} />
}
