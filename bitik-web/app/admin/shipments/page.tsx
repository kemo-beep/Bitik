"use client"

import { AdminResourceClient } from "@/components/admin/admin-resource-client"
import { adminApi } from "@/lib/api/admin"
import { queryKeys } from "@/lib/queryKeys"


export default function Page() {
  return (
    <AdminResourceClient
      title="Shipping and shipments"
      description="Manage shipment list/detail/tracking and status."
      queryKey={queryKeys.admin.shipments()}
      queryFn={() => adminApi.listShipments()}
      actions={[
        { label: "Create provider", placeholder: "Provider name", action: (name) => adminApi.createShippingProvider({ name, code: name.toLowerCase().replaceAll(" ", "_") }) },
        { label: "Shipment detail", placeholder: "Shipment ID", action: (id) => adminApi.getShipment(id) },
        { label: "Shipment tracking", placeholder: "Shipment ID", action: (id) => adminApi.shipmentTracking(id) },
        { label: "Set shipment status", placeholder: "Shipment ID", action: (id) => adminApi.patchShipmentStatus(id, { status: "in_transit" }) },
      ]}
    />
  )
}
