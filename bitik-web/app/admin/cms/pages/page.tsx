"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return (
    <AdminResourceClient
      title="CMS pages"
      description="Manage static pages."
      queryKey={queryKeys.admin.cmsPages()}
      queryFn={adminApi.listCmsPages}
      actions={[
        { label: "Create page", placeholder: "Slug", action: (slug) => adminApi.createCmsPage({ slug, title: slug, body: "..." }) },
        { label: "Delete page", placeholder: "Page ID", action: (id) => adminApi.deleteCmsPage(id) },
      ]}
    />
  )
}
