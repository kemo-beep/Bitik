"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { z } from "zod"
import { zodResolver } from "@hookform/resolvers/zod"
import { useMutation } from "@tanstack/react-query"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { FormError } from "@/components/shared/form-error"
import { sendPhoneOtp, verifyPhoneOtp } from "@/lib/api/auth"
import { useAuth } from "@/lib/auth/auth-context"
import { BitikAPIError } from "@/lib/api/errors"

const SendSchema = z.object({
  phone: z.string().min(5, "Enter a valid phone number."),
})
type SendValues = z.infer<typeof SendSchema>

const VerifySchema = z.object({
  phone: z.string().min(5),
  otp: z.string().min(4, "Enter the code."),
})
type VerifyValues = z.infer<typeof VerifySchema>

export default function Page() {
  const auth = useAuth()
  const [cooldownUntil, setCooldownUntil] = React.useState<number | null>(null)
  const [now, setNow] = React.useState(() => Date.now())

  const sendForm = useForm<SendValues>({
    resolver: zodResolver(SendSchema),
    defaultValues: { phone: "" },
  })

  const verifyForm = useForm<VerifyValues>({
    resolver: zodResolver(VerifySchema),
    defaultValues: { phone: "", otp: "" },
  })

  const sendMutation = useMutation({
    mutationFn: async (values: SendValues) => {
      await sendPhoneOtp(values)
    },
    onSuccess: (_, values) => {
      verifyForm.setValue("phone", values.phone)
      setCooldownUntil(Date.now() + 60_000)
    },
    onError: (err) => {
      if (err instanceof BitikAPIError && err.status === 429) {
        setCooldownUntil(Date.now() + 60_000)
      }
    },
  })

  const verifyMutation = useMutation({
    mutationFn: async (values: VerifyValues) => {
      await verifyPhoneOtp(values)
    },
  })

  const remaining = cooldownUntil ? Math.max(0, cooldownUntil - now) : 0
  const canSend = auth.status === "authenticated" && remaining === 0

  React.useEffect(() => {
    if (!cooldownUntil) return
    if (remaining === 0) return
    const t = window.setTimeout(() => {
      setNow(Date.now())
    }, 1000)
    return () => window.clearTimeout(t)
  }, [cooldownUntil, remaining])

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-lg font-semibold tracking-tight">Verify phone</h1>
        <p className="text-sm text-muted-foreground">
          Send an OTP to your phone and confirm it. You must be signed in.
        </p>
      </div>

      <FormError error={sendMutation.error} title="OTP send failed" />

      <form
        className="space-y-4"
        onSubmit={sendForm.handleSubmit((v) => sendMutation.mutate(v))}
      >
        <FieldGroup>
          <Field data-invalid={!!sendForm.formState.errors.phone}>
            <FieldLabel htmlFor="phone">Phone</FieldLabel>
            <FieldContent>
              <Input
                id="phone"
                autoComplete="tel"
                inputMode="tel"
                placeholder="+1 555 123 4567"
                {...sendForm.register("phone")}
              />
              <FieldError errors={[sendForm.formState.errors.phone]} />
            </FieldContent>
          </Field>
        </FieldGroup>

        <Button type="submit" className="w-full" disabled={!canSend || sendMutation.isPending}>
          {sendMutation.isPending
            ? "Sending…"
            : remaining > 0
              ? `Try again in ${Math.ceil(remaining / 1000)}s`
              : "Send OTP"}
        </Button>

        {auth.status !== "authenticated" && (
          <p className="text-xs text-muted-foreground">
            Please sign in first to verify your phone.
          </p>
        )}
      </form>

      <div className="h-px bg-border" />

      <FormError error={verifyMutation.error} title="Verification failed" />

      {verifyMutation.isSuccess ? (
        <p className="text-sm">Phone verified.</p>
      ) : (
        <form
          className="space-y-4"
          onSubmit={verifyForm.handleSubmit((v) => verifyMutation.mutate(v))}
        >
          <FieldGroup>
            <Field data-invalid={!!verifyForm.formState.errors.phone}>
              <FieldLabel htmlFor="phone2">Phone</FieldLabel>
              <FieldContent>
                <Input id="phone2" placeholder="+1 555…" {...verifyForm.register("phone")} />
                <FieldError errors={[verifyForm.formState.errors.phone]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!verifyForm.formState.errors.otp}>
              <FieldLabel htmlFor="otp">OTP code</FieldLabel>
              <FieldContent>
                <Input id="otp" placeholder="123456" {...verifyForm.register("otp")} />
                <FieldError errors={[verifyForm.formState.errors.otp]} />
              </FieldContent>
            </Field>
          </FieldGroup>

          <Button type="submit" className="w-full" disabled={verifyMutation.isPending}>
            {verifyMutation.isPending ? "Verifying…" : "Verify"}
          </Button>
        </form>
      )}
    </div>
  )
}
