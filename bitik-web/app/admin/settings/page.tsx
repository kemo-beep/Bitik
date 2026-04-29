"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return (
    <AdminResourceClient
      title="Settings"
      description="Platform settings and feature flags."
      queryKey={queryKeys.admin.platformSettings()}
      queryFn={adminApi.listPlatformSettings}
      actions={[
        { label: "Set maintenance", placeholder: "true|false", action: (value) => adminApi.putFeatureFlag("maintenance_mode", { enabled: value === "true" }) },
      ]}
    />
  )
}
