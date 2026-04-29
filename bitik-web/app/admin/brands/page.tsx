"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"


export default function Page() {
  return (
    <AdminResourceClient
      title="Brands"
      description="Manage brand CRUD."
      queryKey={queryKeys.admin.brands()}
      queryFn={adminApi.listBrands}
      actions={[
        { label: "Create", placeholder: '{"name":"Acme","slug":"acme"}', action: (v) => adminApi.createBrand(JSON.parse(v || "{}")) },
        { label: "Update", placeholder: '{"id":"...","name":"Acme+"}', action: (v) => {
          const input = JSON.parse(v || "{}") as { id?: string } & Record<string, unknown>
          const id = String(input.id ?? "")
          const { id: _id, ...payload } = input
          return adminApi.patchBrand(id, payload)
        } },
        { label: "Delete", placeholder: "Brand ID", action: (id) => adminApi.deleteBrand(id) },
      ]}
    />
  )
}

