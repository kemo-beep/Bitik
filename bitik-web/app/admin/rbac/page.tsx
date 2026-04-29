"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"

export default function Page() {
  return (
    <AdminResourceClient
      title="RBAC"
      description="Manage roles and permissions."
      queryKey={queryKeys.admin.roles()}
      queryFn={adminApi.listRoles}
      actions={[
        { label: "Create role", placeholder: "Role key", action: (key) => adminApi.createRole({ key, name: key }) },
        { label: "Delete role", placeholder: "Role ID", action: (id) => adminApi.deleteRole(id) },
      ]}
    />
  )
}
