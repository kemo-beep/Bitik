"use client"

import { Card } from "@/components/ui/card"
import { Switch } from "@/components/ui/switch"

export default function Page() {
  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-lg font-semibold tracking-tight">Preferences</h1>
        <p className="text-sm text-muted-foreground">
          Notification preferences are UI-only for now (backend endpoint can be wired when available).
        </p>
      </div>

      <Card className="p-4 space-y-3">
        <div className="flex items-center justify-between gap-4">
          <div>
            <p className="text-sm font-medium">Order updates</p>
            <p className="text-xs text-muted-foreground">Emails about shipping and delivery.</p>
          </div>
          <Switch defaultChecked />
        </div>
        <div className="flex items-center justify-between gap-4">
          <div>
            <p className="text-sm font-medium">Promotions</p>
            <p className="text-xs text-muted-foreground">Deals and product updates.</p>
          </div>
          <Switch />
        </div>
      </Card>
    </div>
  )
}
