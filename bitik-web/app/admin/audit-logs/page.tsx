"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return <AdminResourceClient title="Audit logs" description="Sensitive action history." queryKey={queryKeys.admin.auditLogs()} queryFn={() => adminApi.listAuditLogs()} />
}
