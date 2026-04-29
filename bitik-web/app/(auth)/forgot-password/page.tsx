"use client"

import Link from "next/link"
import { useForm } from "react-hook-form"
import { z } from "zod"
import { zodResolver } from "@hookform/resolvers/zod"
import { useMutation } from "@tanstack/react-query"

import { buttonVariants } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { FormError } from "@/components/shared/form-error"
import { routes } from "@/lib/routes"
import { forgotPassword } from "@/lib/api/auth"
import { cn } from "@/lib/utils"

const Schema = z.object({
  email: z.string().email(),
})
type Values = z.infer<typeof Schema>

export default function Page() {
  const form = useForm<Values>({
    resolver: zodResolver(Schema),
    defaultValues: { email: "" },
  })

  const mutation = useMutation({
    mutationFn: forgotPassword,
  })

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-lg font-semibold tracking-tight">Forgot password</h1>
        <p className="text-sm text-muted-foreground">
          We’ll email you a reset link if that address exists.
        </p>
      </div>

      <FormError error={mutation.error} title="Request failed" />

      {mutation.isSuccess ? (
        <div className="space-y-3">
          <p className="text-sm">
            If that email exists, you’ll receive instructions shortly.
          </p>
          <Link
            href={routes.auth.login}
            className={cn(buttonVariants({ className: "w-full" }))}
          >
            Back to sign in
          </Link>
        </div>
      ) : (
        <form
          className="space-y-4"
          onSubmit={form.handleSubmit((v) => mutation.mutate(v))}
        >
          <FieldGroup>
            <Field data-invalid={!!form.formState.errors.email}>
              <FieldLabel htmlFor="email">Email</FieldLabel>
              <FieldContent>
                <Input
                  id="email"
                  autoComplete="email"
                  inputMode="email"
                  placeholder="you@example.com"
                  {...form.register("email")}
                />
                <FieldError errors={[form.formState.errors.email]} />
              </FieldContent>
            </Field>
          </FieldGroup>

          <button
            type="submit"
            className={cn(buttonVariants({ className: "w-full" }))}
            disabled={mutation.isPending}
          >
            {mutation.isPending ? "Sending…" : "Send reset link"}
          </button>

          <div className="text-center text-sm">
            <Link className="underline underline-offset-4" href={routes.auth.login}>
              Back to sign in
            </Link>
          </div>
        </form>
      )}
    </div>
  )
}
