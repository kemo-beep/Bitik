"use client"

import Link from "next/link"
import { useRouter } from "next/navigation"
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

const Schema = z.object({
  email: z.string().email(),
  password: z.string().min(8, "Password must be at least 8 characters."),
  displayName: z.string().optional(),
})
type Values = z.infer<typeof Schema>

export default function Page() {
  const router = useRouter()
  const auth = useAuth()

  const form = useForm<Values>({
    resolver: zodResolver(Schema),
    defaultValues: { email: "", password: "", displayName: "" },
  })

  const mutation = useMutation({
    mutationFn: async (values: Values) => {
      await auth.register({ email: values.email, password: values.password })
    },
    onSuccess: () => {
      router.replace(routes.storefront.home)
    },
  })

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-lg font-semibold tracking-tight">Create account</h1>
        <p className="text-sm text-muted-foreground">
          Register with your email and a password.
        </p>
      </div>

      <FormError error={mutation.error} title="Registration failed" />

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
                autoComplete="new-password"
                placeholder="At least 8 characters"
                {...form.register("password")}
              />
              <FieldError errors={[form.formState.errors.password]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!form.formState.errors.displayName}>
            <FieldLabel htmlFor="displayName">Display name (optional)</FieldLabel>
            <FieldContent>
              <Input
                id="displayName"
                autoComplete="name"
                placeholder="How you want to appear"
                {...form.register("displayName")}
              />
              <FieldError errors={[form.formState.errors.displayName]} />
            </FieldContent>
          </Field>
        </FieldGroup>

        <Button type="submit" className="w-full" disabled={mutation.isPending}>
          {mutation.isPending ? "Creating…" : "Create account"}
        </Button>

        <div className="text-center text-sm">
          Already have an account?{" "}
          <Link className="underline underline-offset-4" href={routes.auth.login}>
            Sign in
          </Link>
        </div>
      </form>
    </div>
  )
}
