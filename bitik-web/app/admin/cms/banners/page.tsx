"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return <AdminResourceClient title="CMS banners" description="Banner management." queryKey={queryKeys.admin.cmsBanners()} queryFn={adminApi.listCmsBanners} />
}

