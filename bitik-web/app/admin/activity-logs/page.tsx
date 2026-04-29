"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return <AdminResourceClient title="Activity logs" description="Admin activity stream." queryKey={queryKeys.admin.activityLogs()} queryFn={() => adminApi.listActivityLogs()} />
}

