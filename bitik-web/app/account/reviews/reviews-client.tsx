"use client"

import { useState } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { deleteLocalBuyerReview, listLocalBuyerReviews, saveLocalBuyerReview } from "@/lib/api/buyer"
import { queryKeys } from "@/lib/queryKeys"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { asString } from "@/lib/safe"

export function ReviewsClient() {
  const qc = useQueryClient()
  const [title, setTitle] = useState("")
  const [body, setBody] = useState("")
  const [rating, setRating] = useState("5")

  const reviews = useQuery({
    queryKey: queryKeys.buyer.localReviews(),
    queryFn: async () => listLocalBuyerReviews(),
  })
  const refresh = () => qc.invalidateQueries({ queryKey: queryKeys.buyer.localReviews() })

  return (
    <div className="mx-auto max-w-screen-lg px-4 py-8">
      <h1 className="font-heading text-2xl font-semibold">My reviews</h1>
      <p className="mt-1 text-sm text-muted-foreground">Local stub until buyer review APIs are exposed in OpenAPI.</p>

      <div className="mt-4 rounded-xl border p-4">
        <h2 className="font-medium">Create / edit review</h2>
        <div className="mt-3 grid gap-2 md:grid-cols-3">
          <Input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Title" />
          <Input value={rating} onChange={(e) => setRating(e.target.value)} placeholder="Rating 1-5" />
          <Input value={body} onChange={(e) => setBody(e.target.value)} placeholder="Body" />
        </div>
        <Button
          className="mt-3"
          onClick={() => {
            saveLocalBuyerReview({
              id: crypto.randomUUID(),
              title: title.trim(),
              body: body.trim(),
              rating: Number.parseInt(rating, 10) || 5,
              created_at: new Date().toISOString(),
            })
            setTitle("")
            setBody("")
            setRating("5")
            refresh()
          }}
          disabled={!title.trim()}
        >
          Save review
        </Button>
      </div>

      <div className="mt-4 space-y-3">
        {(reviews.data ?? []).map((r) => {
          const rec = r as Record<string, unknown>
          const id = asString(rec.id) ?? ""
          return (
            <div key={id} className="rounded-xl border p-4">
              <div className="font-medium">{asString(rec.title) ?? "Review"}</div>
              <div className="text-sm text-muted-foreground">{asString(rec.body) ?? ""}</div>
              <div className="mt-2 text-xs">Rating: {String(rec.rating ?? "—")}</div>
              <div className="mt-2">
                <Button
                  variant="outline"
                  onClick={() => {
                    deleteLocalBuyerReview(id)
                    refresh()
                  }}
                >
                  Delete
                </Button>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

