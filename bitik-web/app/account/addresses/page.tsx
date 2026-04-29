"use client"

import { useEffect, useMemo, useState } from "react"
import { useForm } from "react-hook-form"
import { z } from "zod"
import { zodResolver } from "@hookform/resolvers/zod"
import { useMutation, useQuery } from "@tanstack/react-query"

import { Card } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { FormError } from "@/components/shared/form-error"
import {
  createAddress,
  deleteAddress,
  listAddresses,
  setDefaultAddress,
  updateAddress,
  type Address,
  type CreateAddressInput,
} from "@/lib/api/account"

const AddressSchema = z.object({
  full_name: z.string().min(1, "Full name is required."),
  phone: z.string().min(3, "Phone is required."),
  country: z.string().min(1, "Country is required."),
  state: z.string().optional(),
  city: z.string().optional(),
  district: z.string().optional(),
  postal_code: z.string().optional(),
  address_line1: z.string().min(1, "Address line 1 is required."),
  address_line2: z.string().optional(),
  is_default: z.boolean().optional(),
})
type AddressValues = z.infer<typeof AddressSchema>

function toCreateInput(values: AddressValues) {
  const input: CreateAddressInput = {
    ...values,
    is_default: values.is_default ?? false,
  }
  return input
}

export default function Page() {
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Address | null>(null)

  const addresses = useQuery({ queryKey: ["addresses"], queryFn: listAddresses })

  const form = useForm<AddressValues>({
    resolver: zodResolver(AddressSchema),
    defaultValues: {
      full_name: "",
      phone: "",
      country: "",
      state: "",
      city: "",
      district: "",
      postal_code: "",
      address_line1: "",
      address_line2: "",
      is_default: false,
    },
  })

  useEffect(() => {
    if (!open) {
      setEditing(null)
      form.reset()
    }
  }, [form, open])

  const upsertMutation = useMutation({
    mutationFn: async (values: AddressValues) => {
      if (editing) {
        return await updateAddress(editing.id, values)
      }
      return await createAddress(toCreateInput(values))
    },
    onSuccess: () => {
      void addresses.refetch()
      setOpen(false)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (addressId: string) => {
      await deleteAddress(addressId)
    },
    onSuccess: () => void addresses.refetch(),
  })

  const defaultMutation = useMutation({
    mutationFn: setDefaultAddress,
    onSuccess: () => void addresses.refetch(),
  })

  const sorted = useMemo(() => {
    const items = addresses.data ?? []
    return [...items].sort((a, b) => Number(b.is_default) - Number(a.is_default))
  }, [addresses.data])

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div className="space-y-1">
          <h1 className="text-lg font-semibold tracking-tight">Addresses</h1>
          <p className="text-sm text-muted-foreground">Manage your shipping addresses.</p>
        </div>

        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger render={<Button />}>Add address</DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{editing ? "Edit address" : "Add address"}</DialogTitle>
            </DialogHeader>
            <FormError error={upsertMutation.error} title="Save failed" />
            <form
              className="space-y-4"
              onSubmit={form.handleSubmit((v) => upsertMutation.mutate(v))}
            >
              <FieldGroup>
                <Field data-invalid={!!form.formState.errors.full_name}>
                  <FieldLabel htmlFor="full_name">Full name</FieldLabel>
                  <FieldContent>
                    <Input id="full_name" {...form.register("full_name")} />
                    <FieldError errors={[form.formState.errors.full_name]} />
                  </FieldContent>
                </Field>
                <Field data-invalid={!!form.formState.errors.phone}>
                  <FieldLabel htmlFor="phone">Phone</FieldLabel>
                  <FieldContent>
                    <Input id="phone" {...form.register("phone")} />
                    <FieldError errors={[form.formState.errors.phone]} />
                  </FieldContent>
                </Field>
                <Field data-invalid={!!form.formState.errors.country}>
                  <FieldLabel htmlFor="country">Country</FieldLabel>
                  <FieldContent>
                    <Input id="country" {...form.register("country")} />
                    <FieldError errors={[form.formState.errors.country]} />
                  </FieldContent>
                </Field>
                <Field data-invalid={!!form.formState.errors.address_line1}>
                  <FieldLabel htmlFor="address_line1">Address line 1</FieldLabel>
                  <FieldContent>
                    <Input id="address_line1" {...form.register("address_line1")} />
                    <FieldError errors={[form.formState.errors.address_line1]} />
                  </FieldContent>
                </Field>
                <Field>
                  <FieldLabel htmlFor="address_line2">Address line 2</FieldLabel>
                  <FieldContent>
                    <Input id="address_line2" {...form.register("address_line2")} />
                  </FieldContent>
                </Field>
                <Field>
                  <FieldLabel htmlFor="city">City</FieldLabel>
                  <FieldContent>
                    <Input id="city" {...form.register("city")} />
                  </FieldContent>
                </Field>
                <Field>
                  <FieldLabel htmlFor="state">State</FieldLabel>
                  <FieldContent>
                    <Input id="state" {...form.register("state")} />
                  </FieldContent>
                </Field>
                <Field>
                  <FieldLabel htmlFor="district">District</FieldLabel>
                  <FieldContent>
                    <Input id="district" {...form.register("district")} />
                  </FieldContent>
                </Field>
                <Field>
                  <FieldLabel htmlFor="postal_code">Postal code</FieldLabel>
                  <FieldContent>
                    <Input id="postal_code" {...form.register("postal_code")} />
                  </FieldContent>
                </Field>
              </FieldGroup>

              <Button type="submit" className="w-full" disabled={upsertMutation.isPending}>
                {upsertMutation.isPending ? "Saving…" : "Save"}
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <FormError error={addresses.error} title="Could not load addresses" />
      <FormError error={deleteMutation.error} title="Delete failed" />
      <FormError error={defaultMutation.error} title="Update failed" />

      <div className="space-y-3">
        {addresses.isLoading ? (
          <p className="text-sm text-muted-foreground">Loading…</p>
        ) : sorted.length === 0 ? (
          <Card className="p-4">
            <p className="text-sm text-muted-foreground">No addresses yet.</p>
          </Card>
        ) : (
          sorted.map((a) => (
            <Card key={a.id} className="p-4 space-y-2">
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="font-medium truncate">{a.full_name}</p>
                    {a.is_default && <Badge variant="success">Default</Badge>}
                  </div>
                  <p className="text-sm text-muted-foreground truncate">{a.phone}</p>
                  <p className="text-sm">
                    {a.address_line1}
                    {a.address_line2 ? `, ${a.address_line2}` : ""}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    {[a.district, a.city, a.state, a.postal_code, a.country]
                      .filter(Boolean)
                      .join(", ")}
                  </p>
                </div>

                <div className="flex flex-col gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setEditing(a)
                      form.reset({
                        full_name: a.full_name,
                        phone: a.phone,
                        country: a.country,
                        state: a.state ?? "",
                        city: a.city ?? "",
                        district: a.district ?? "",
                        postal_code: a.postal_code ?? "",
                        address_line1: a.address_line1,
                        address_line2: a.address_line2 ?? "",
                        is_default: a.is_default,
                      })
                      setOpen(true)
                    }}
                  >
                    Edit
                  </Button>
                  {!a.is_default && (
                    <Button
                      variant="secondary"
                      size="sm"
                      disabled={defaultMutation.isPending}
                      onClick={() => defaultMutation.mutate(a.id)}
                    >
                      Set default
                    </Button>
                  )}
                  <Button
                    variant="destructive"
                    size="sm"
                    disabled={deleteMutation.isPending}
                    onClick={() => deleteMutation.mutate(a.id)}
                  >
                    Delete
                  </Button>
                </div>
              </div>
            </Card>
          ))
        )}
      </div>
    </div>
  )
}
