"use client"

import Link from "next/link"
import { useSearchParams } from "next/navigation"
import { useForm } from "react-hook-form"
import { z } from "zod"
import { zodResolver } from "@hookform/resolvers/zod"
import { useMutation } from "@tanstack/react-query"

import { buttonVariants } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { FormError } from "@/components/shared/form-error"
import { routes } from "@/lib/routes"
import { resendEmailVerification, verifyEmail } from "@/lib/api/auth"
import { useAuth } from "@/lib/auth/auth-context"
import { cn } from "@/lib/utils"

const Schema = z.object({
  token: z.string().min(1, "Token is required."),
})
type Values = z.infer<typeof Schema>

export default function Page() {
  const search = useSearchParams()
  const tokenFromQuery = search.get("token") ?? ""
  const auth = useAuth()

  const form = useForm<Values>({
    resolver: zodResolver(Schema),
    defaultValues: { token: tokenFromQuery },
  })

  const verifyMutation = useMutation({
    mutationFn: verifyEmail,
  })

  const resendMutation = useMutation({
    mutationFn: resendEmailVerification,
  })

  const canResend = auth.status === "authenticated"

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-lg font-semibold tracking-tight">Verify email</h1>
        <p className="text-sm text-muted-foreground">
          Paste your verification token, or resend if you’re signed in.
        </p>
      </div>

      <FormError error={verifyMutation.error} title="Verification failed" />

      {verifyMutation.isSuccess ? (
        <div className="space-y-3">
          <p className="text-sm">Email verified. You can continue.</p>
          <Link
            href={routes.storefront.home}
            className={cn(buttonVariants({ className: "w-full" }))}
          >
            Go to home
          </Link>
        </div>
      ) : (
        <form
          className="space-y-4"
          onSubmit={form.handleSubmit((v) => verifyMutation.mutate(v))}
        >
          <FieldGroup>
            <Field data-invalid={!!form.formState.errors.token}>
              <FieldLabel htmlFor="token">Verification token</FieldLabel>
              <FieldContent>
                <Input id="token" placeholder="Paste token" {...form.register("token")} />
                <FieldError errors={[form.formState.errors.token]} />
              </FieldContent>
            </Field>
          </FieldGroup>

          <button
            type="submit"
            className={cn(buttonVariants({ className: "w-full" }))}
            disabled={verifyMutation.isPending}
          >
            {verifyMutation.isPending ? "Verifying…" : "Verify"}
          </button>
        </form>
      )}

      <div className="space-y-2">
        <FormError error={resendMutation.error} title="Resend failed" />
        <button
          type="button"
          className={cn(buttonVariants({ variant: "outline", className: "w-full" }))}
          disabled={!canResend || resendMutation.isPending}
          onClick={() => resendMutation.mutate()}
        >
          {resendMutation.isPending ? "Resending…" : "Resend verification email"}
        </button>
        {!canResend && (
          <p className="text-xs text-muted-foreground">
            Sign in to resend verification email.
          </p>
        )}
      </div>
    </div>
  )
}
