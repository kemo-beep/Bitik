"use client"

import * as React from "react"
import { useMutation } from "@tanstack/react-query"

import { Card } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { FormError } from "@/components/shared/form-error"
import { deleteMe } from "@/lib/api/account"
import { useAuth } from "@/lib/auth/auth-context"

export default function Page() {
  const auth = useAuth()
  const [phrase, setPhrase] = React.useState("")

  const mutation = useMutation({
    mutationFn: async () => {
      await deleteMe()
      await auth.signOut()
    },
  })

  const required = "delete my account"
  const ok = phrase.trim().toLowerCase() === required

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-lg font-semibold tracking-tight">Security</h1>
        <p className="text-sm text-muted-foreground">
          Sensitive actions like account deletion require confirmation.
        </p>
      </div>

      <Card className="p-4 space-y-3">
        <h2 className="text-sm font-medium">Sign out</h2>
        <Button variant="outline" onClick={() => auth.signOut()}>
          Sign out
        </Button>
      </Card>

      <Card className="p-4 space-y-3">
        <h2 className="text-sm font-medium text-destructive">Delete account</h2>
        <p className="text-sm text-muted-foreground">
          This will soft-delete your account and revoke sessions. This action is hard to undo.
        </p>
        <Dialog>
          <DialogTrigger render={<Button variant="destructive" />}>
            Delete my account
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm account deletion</DialogTitle>
            </DialogHeader>
            <FormError error={mutation.error} title="Deletion failed" />
            <div className="space-y-3">
              <p className="text-sm">
                Type <span className="font-medium">{required}</span> to confirm.
              </p>
              <Input value={phrase} onChange={(e) => setPhrase(e.target.value)} />
              <Button
                variant="destructive"
                disabled={!ok || mutation.isPending}
                onClick={() => mutation.mutate()}
              >
                {mutation.isPending ? "Deleting…" : "Permanently delete"}
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </Card>
    </div>
  )
}
