"use client"

import Link from "next/link"
import { useRouter, useSearchParams } from "next/navigation"
import { useForm } from "react-hook-form"
import { z } from "zod"
import { zodResolver } from "@hookform/resolvers/zod"
import { useMutation } from "@tanstack/react-query"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Field,
  FieldContent,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { FormError } from "@/components/shared/form-error"
import { routes } from "@/lib/routes"
import { useAuth } from "@/lib/auth/auth-context"
import { env } from "@/lib/env"
const Schema = z.object({
  email: z.string().email(),
  password: z.string().min(1, "Password is required."),
})
type Values = z.infer<typeof Schema>

export default function Page() {
  const router = useRouter()
  const search = useSearchParams()
  const redirectTo = search.get("redirect") || routes.storefront.home
  const auth = useAuth()

  const form = useForm<Values>({
    resolver: zodResolver(Schema),
    defaultValues: { email: "", password: "" },
  })

  const mutation = useMutation({
    mutationFn: async (values: Values) => {
      await auth.signIn(values)
    },
    onSuccess: () => {
      router.replace(redirectTo)
    },
  })

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-lg font-semibold tracking-tight">Welcome back</h1>
        <p className="text-sm text-muted-foreground">
          Sign in to manage your account, orders, and seller tools.
        </p>
      </div>

      <FormError error={mutation.error} title="Sign in failed" />

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

          <Field data-invalid={!!form.formState.errors.password}>
            <FieldLabel htmlFor="password">Password</FieldLabel>
            <FieldContent>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                placeholder="Your password"
                {...form.register("password")}
              />
              <FieldError errors={[form.formState.errors.password]} />
            </FieldContent>
          </Field>
        </FieldGroup>

        <Button
          type="submit"
          className="w-full"
          disabled={mutation.isPending}
        >
          {mutation.isPending ? "Signing in…" : "Sign in"}
        </Button>

        <div className="flex items-center justify-between text-sm">
          <Link className="underline underline-offset-4" href={routes.auth.forgotPassword}>
            Forgot password?
          </Link>
          <Link className="underline underline-offset-4" href={routes.auth.register}>
            Create account
          </Link>
        </div>
      </form>

      <div className="space-y-2">
        <p className="text-xs text-muted-foreground">Or continue with</p>
        <div className="grid grid-cols-1 gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              window.location.href = `${env.apiBaseUrl}/auth/oauth/google`
            }}
          >
            Google
          </Button>
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              window.location.href = `${env.apiBaseUrl}/auth/oauth/facebook`
            }}
          >
            Facebook
          </Button>
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              window.location.href = `${env.apiBaseUrl}/auth/oauth/apple`
            }}
          >
            Apple
          </Button>
        </div>
      </div>
    </div>
  )
}
