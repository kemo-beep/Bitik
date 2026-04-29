"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return (
    <AdminResourceClient
      title="CMS announcements"
      description="Announcements management."
      queryKey={queryKeys.admin.cmsAnnouncements()}
      queryFn={adminApi.listCmsAnnouncements}
    />
  )
}

