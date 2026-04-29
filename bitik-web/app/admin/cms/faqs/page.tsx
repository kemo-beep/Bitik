"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return <AdminResourceClient title="CMS FAQs" description="FAQ management." queryKey={queryKeys.admin.cmsFaqs()} queryFn={adminApi.listCmsFaqs} />
}

