"use client"

import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { z } from "zod"
import { zodResolver } from "@hookform/resolvers/zod"
import { useMutation, useQuery } from "@tanstack/react-query"

import { Card } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { FormError } from "@/components/shared/form-error"
import { getMe } from "@/lib/api/auth"
import { getProfile, patchMe, patchProfile } from "@/lib/api/account"

const ContactSchema = z.object({
  email: z.string().email().optional().or(z.literal("")),
  phone: z.string().optional().or(z.literal("")),
})
type ContactValues = z.infer<typeof ContactSchema>

const ProfileSchema = z.object({
  first_name: z.string().optional().or(z.literal("")),
  last_name: z.string().optional().or(z.literal("")),
  display_name: z.string().optional().or(z.literal("")),
  avatar_url: z.string().url().optional().or(z.literal("")),
  language: z.string().optional().or(z.literal("")),
  country_code: z.string().optional().or(z.literal("")),
  timezone: z.string().optional().or(z.literal("")),
})
type ProfileValues = z.infer<typeof ProfileSchema>

export default function Page() {
  const me = useQuery({ queryKey: ["me"], queryFn: getMe })
  const profile = useQuery({ queryKey: ["profile"], queryFn: getProfile })

  const contactForm = useForm<ContactValues>({
    resolver: zodResolver(ContactSchema),
    defaultValues: { email: "", phone: "" },
  })
  const profileForm = useForm<ProfileValues>({
    resolver: zodResolver(ProfileSchema),
    defaultValues: {
      first_name: "",
      last_name: "",
      display_name: "",
      avatar_url: "",
      language: "",
      country_code: "",
      timezone: "",
    },
  })

  useEffect(() => {
    if (me.data) {
      contactForm.reset({
        email: me.data.email ?? "",
        phone: me.data.phone ?? "",
      })
    }
  }, [contactForm, me.data])

  useEffect(() => {
    if (profile.data) {
      profileForm.reset({
        first_name: profile.data.first_name ?? "",
        last_name: profile.data.last_name ?? "",
        display_name: profile.data.display_name ?? "",
        avatar_url: profile.data.avatar_url ?? "",
        language: profile.data.language ?? "",
        country_code: profile.data.country_code ?? "",
        timezone: profile.data.timezone ?? "",
      })
    }
  }, [profile.data, profileForm])

  const contactMutation = useMutation({
    mutationFn: async (values: ContactValues) => {
      await patchMe({
        email: values.email?.trim() ? values.email.trim() : null,
        phone: values.phone?.trim() ? values.phone.trim() : null,
      })
    },
    onSuccess: () => {
      void me.refetch()
    },
  })

  const profileMutation = useMutation({
    mutationFn: async (values: ProfileValues) => {
      await patchProfile({
        first_name: values.first_name?.trim() || undefined,
        last_name: values.last_name?.trim() || undefined,
        display_name: values.display_name?.trim() || undefined,
        avatar_url: values.avatar_url?.trim() || undefined,
        language: values.language?.trim() || undefined,
        country_code: values.country_code?.trim() || undefined,
        timezone: values.timezone?.trim() || undefined,
      })
    },
    onSuccess: () => {
      void profile.refetch()
    },
  })

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-lg font-semibold tracking-tight">Profile</h1>
        <p className="text-sm text-muted-foreground">
          Update your contact info and profile details.
        </p>
      </div>

      <Card className="p-4 space-y-4">
        <h2 className="text-sm font-medium">Contact</h2>
        <FormError error={contactMutation.error} title="Update failed" />
        <form
          className="space-y-4"
          onSubmit={contactForm.handleSubmit((v) => contactMutation.mutate(v))}
        >
          <FieldGroup>
            <Field data-invalid={!!contactForm.formState.errors.email}>
              <FieldLabel htmlFor="email">Email</FieldLabel>
              <FieldContent>
                <Input id="email" {...contactForm.register("email")} />
                <FieldError errors={[contactForm.formState.errors.email]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!contactForm.formState.errors.phone}>
              <FieldLabel htmlFor="phone">Phone</FieldLabel>
              <FieldContent>
                <Input id="phone" {...contactForm.register("phone")} />
                <FieldError errors={[contactForm.formState.errors.phone]} />
              </FieldContent>
            </Field>
          </FieldGroup>
          <Button type="submit" disabled={contactMutation.isPending}>
            {contactMutation.isPending ? "Saving…" : "Save contact"}
          </Button>
        </form>
      </Card>

      <Card className="p-4 space-y-4">
        <h2 className="text-sm font-medium">Profile details</h2>
        <FormError error={profileMutation.error} title="Update failed" />
        <form
          className="space-y-4"
          onSubmit={profileForm.handleSubmit((v) => profileMutation.mutate(v))}
        >
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="display_name">Display name</FieldLabel>
              <FieldContent>
                <Input id="display_name" {...profileForm.register("display_name")} />
              </FieldContent>
            </Field>
            <Field>
              <FieldLabel htmlFor="first_name">First name</FieldLabel>
              <FieldContent>
                <Input id="first_name" {...profileForm.register("first_name")} />
              </FieldContent>
            </Field>
            <Field>
              <FieldLabel htmlFor="last_name">Last name</FieldLabel>
              <FieldContent>
                <Input id="last_name" {...profileForm.register("last_name")} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!profileForm.formState.errors.avatar_url}>
              <FieldLabel htmlFor="avatar_url">Avatar URL</FieldLabel>
              <FieldContent>
                <Input id="avatar_url" placeholder="https://…" {...profileForm.register("avatar_url")} />
                <FieldError errors={[profileForm.formState.errors.avatar_url]} />
              </FieldContent>
            </Field>
            <Field>
              <FieldLabel htmlFor="language">Language</FieldLabel>
              <FieldContent>
                <Input id="language" placeholder="en" {...profileForm.register("language")} />
              </FieldContent>
            </Field>
            <Field>
              <FieldLabel htmlFor="country_code">Country code</FieldLabel>
              <FieldContent>
                <Input id="country_code" placeholder="US" {...profileForm.register("country_code")} />
              </FieldContent>
            </Field>
            <Field>
              <FieldLabel htmlFor="timezone">Timezone</FieldLabel>
              <FieldContent>
                <Input id="timezone" placeholder="UTC" {...profileForm.register("timezone")} />
              </FieldContent>
            </Field>
          </FieldGroup>
          <Button type="submit" disabled={profileMutation.isPending}>
            {profileMutation.isPending ? "Saving…" : "Save profile"}
          </Button>
        </form>
      </Card>
    </div>
  )
}
