"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return (
    <AdminResourceClient
      title="Users"
      description="List users, update status, and manage roles."
      queryKey={queryKeys.admin.users()}
      queryFn={() => adminApi.listUsers()}
      actions={[
        { label: "Get user", placeholder: "User ID", action: (id) => adminApi.getUser(id) },
        { label: "Deactivate user", placeholder: "User ID", action: (id) => adminApi.patchUserStatus(id, { status: "suspended" }) },
      ]}
    />
  )
}
