"use client"

import { useMemo, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

type AnyRecord = Record<string, unknown>

export function AdminResourceClient({
  title,
  description,
  queryKey,
  queryFn,
  actions = [],
}: {
  title: string
  description: string
  queryKey: readonly unknown[]
  queryFn: () => Promise<AnyRecord>
  actions?: Array<{
    label: string
    placeholder?: string
    action: (value: string) => Promise<AnyRecord>
  }>
}) {
  const qc = useQueryClient()
  const [values, setValues] = useState<string[]>(() => actions.map(() => ""))
  const data = useQuery({ queryKey, queryFn, staleTime: 30_000, gcTime: 5 * 60_000 })
  const [page, setPage] = useState(1)
  const pageSize = 50
  const dataItems = useMemo(() => {
    const raw = (data.data as AnyRecord | undefined)?.items
    if (!Array.isArray(raw)) return null
    return raw as AnyRecord[]
  }, [data.data])
  const pagedItems = useMemo(() => {
    if (!dataItems) return null
    const start = (page - 1) * pageSize
    return dataItems.slice(start, start + pageSize)
  }, [dataItems, page])
  const runAction = useMutation({
    mutationFn: async ({ index }: { index: number }) => {
      const item = actions[index]
      if (!item) return {}
      return item.action(values[index] ?? "")
    },
    onSuccess: () => qc.invalidateQueries({ queryKey }),
  })

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">{title}</h1>
      <p className="text-sm text-muted-foreground">{description}</p>

      {actions.length > 0 ? (
        <section className="rounded-xl border p-4">
          <h2 className="font-medium">Actions</h2>
          <div className="mt-3 space-y-2">
            {actions.map((item, idx) => (
              <div className="flex gap-2" key={`${item.label}-${idx}`}>
                <Input
                  placeholder={item.placeholder ?? "Input"}
                  value={values[idx] ?? ""}
                  onChange={(e) =>
                    setValues((prev) => {
                      const next = [...prev]
                      next[idx] = e.target.value
                      return next
                    })
                  }
                />
                <Button
                  onClick={() => runAction.mutate({ index: idx })}
                  disabled={runAction.isPending}
                >
                  {item.label}
                </Button>
              </div>
            ))}
          </div>
        </section>
      ) : null}

      <section className="rounded-xl border p-4">
        <h2 className="font-medium">Data</h2>
        {dataItems ? (
          <div className="mt-2 flex items-center gap-2 text-xs">
            <span>
              Showing {(page - 1) * pageSize + 1}-{Math.min(page * pageSize, dataItems.length)} of {dataItems.length}
            </span>
            <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage((p) => Math.max(1, p - 1))}>
              Prev
            </Button>
            <Button
              size="sm"
              variant="outline"
              disabled={page * pageSize >= dataItems.length}
              onClick={() => setPage((p) => p + 1)}
            >
              Next
            </Button>
          </div>
        ) : null}
        <pre className="mt-3 max-h-96 overflow-auto rounded-md bg-muted p-3 text-xs">
          {JSON.stringify(dataItems ? { ...(data.data ?? {}), items: pagedItems } : data.data ?? {}, null, 2)}
        </pre>
      </section>
    </div>
  )
}

