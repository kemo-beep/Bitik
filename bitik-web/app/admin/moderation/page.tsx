"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return (
    <AdminResourceClient
      title="Moderation"
      description="Moderation reports and cases."
      queryKey={queryKeys.admin.moderationReports()}
      queryFn={() => adminApi.listModerationReports()}
      actions={[
        { label: "Get report", placeholder: "Report ID", action: (id) => adminApi.getModerationReport(id) },
        { label: "Resolve report", placeholder: "Report ID", action: (id) => adminApi.patchModerationReportStatus(id, { status: "resolved" }) },
      ]}
    />
  )
}
