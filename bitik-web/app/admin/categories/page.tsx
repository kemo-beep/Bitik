"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"


export default function Page() {
  return (
    <AdminResourceClient
      title="Categories"
      description="Manage category CRUD and sort ordering."
      queryKey={queryKeys.admin.categories()}
      queryFn={adminApi.listCategories}
      actions={[
        { label: "Create", placeholder: '{"name":"Shoes","slug":"shoes"}', action: (v) => adminApi.createCategory(JSON.parse(v || "{}")) },
        { label: "Update", placeholder: '{"id":"...","name":"Sneakers"}', action: (v) => {
          const input = JSON.parse(v || "{}") as { id?: string } & Record<string, unknown>
          const id = String(input.id ?? "")
          const { id: _id, ...payload } = input
          return adminApi.patchCategory(id, payload)
        } },
        { label: "Delete", placeholder: "Category ID", action: (id) => adminApi.deleteCategory(id) },
        { label: "Reorder", placeholder: '[{"id":"...","sort_order":1}]', action: (v) => adminApi.reorderCategories(JSON.parse(v || "[]")) },
      ]}
    />
  )
}

