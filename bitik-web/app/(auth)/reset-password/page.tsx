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
import { resetPassword } from "@/lib/api/auth"
import { cn } from "@/lib/utils"

const Schema = z
  .object({
    token: z.string().min(1, "Token is required."),
    new_password: z.string().min(8, "Password must be at least 8 characters."),
    confirm: z.string().min(1),
  })
  .refine((v) => v.new_password === v.confirm, {
    message: "Passwords do not match.",
    path: ["confirm"],
  })
type Values = z.infer<typeof Schema>

export default function Page() {
  const search = useSearchParams()
  const tokenFromQuery = search.get("token") ?? ""

  const form = useForm<Values>({
    resolver: zodResolver(Schema),
    defaultValues: { token: tokenFromQuery, new_password: "", confirm: "" },
  })

  const mutation = useMutation({
    mutationFn: async (values: Values) => {
      await resetPassword({ token: values.token, new_password: values.new_password })
    },
  })

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-lg font-semibold tracking-tight">Reset password</h1>
        <p className="text-sm text-muted-foreground">
          Choose a new password for your account.
        </p>
      </div>

      <FormError error={mutation.error} title="Reset failed" />

      {mutation.isSuccess ? (
        <div className="space-y-3">
          <p className="text-sm">Your password has been updated. You can sign in now.</p>
          <Link
            href={routes.auth.login}
            className={cn(buttonVariants({ className: "w-full" }))}
          >
            Sign in
          </Link>
        </div>
      ) : (
        <form
          className="space-y-4"
          onSubmit={form.handleSubmit((v) => mutation.mutate(v))}
        >
          <FieldGroup>
            <Field data-invalid={!!form.formState.errors.token}>
              <FieldLabel htmlFor="token">Reset token</FieldLabel>
              <FieldContent>
                <Input id="token" placeholder="Paste token" {...form.register("token")} />
                <FieldError errors={[form.formState.errors.token]} />
              </FieldContent>
            </Field>

            <Field data-invalid={!!form.formState.errors.new_password}>
              <FieldLabel htmlFor="new_password">New password</FieldLabel>
              <FieldContent>
                <Input
                  id="new_password"
                  type="password"
                  autoComplete="new-password"
                  placeholder="At least 8 characters"
                  {...form.register("new_password")}
                />
                <FieldError errors={[form.formState.errors.new_password]} />
              </FieldContent>
            </Field>

            <Field data-invalid={!!form.formState.errors.confirm}>
              <FieldLabel htmlFor="confirm">Confirm password</FieldLabel>
              <FieldContent>
                <Input
                  id="confirm"
                  type="password"
                  autoComplete="new-password"
                  placeholder="Repeat password"
                  {...form.register("confirm")}
                />
                <FieldError errors={[form.formState.errors.confirm]} />
              </FieldContent>
            </Field>
          </FieldGroup>

          <button
            type="submit"
            className={cn(buttonVariants({ className: "w-full" }))}
            disabled={mutation.isPending}
          >
            {mutation.isPending ? "Updating…" : "Update password"}
          </button>
        </form>
      )}
    </div>
  )
}
