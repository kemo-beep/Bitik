"use client"

import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { asArray, asString } from "@/lib/safe"
import { SellerJsonView, SellerSectionCard } from "@/components/seller/section-card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

export function SellerReviewsClient() {
  const qc = useQueryClient()
  const [reviewId, setReviewId] = useState("")
  const [reply, setReply] = useState("")
  const reviews = useQuery({
    queryKey: ["seller", "reviews"],
    queryFn: sellerApi.listReviews,
  })
  const replyMutation = useMutation({
    mutationFn: () => sellerApi.replyReview(reviewId.trim(), reply.trim()),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["seller", "reviews"] }),
  })
  const reportMutation = useMutation({
    mutationFn: () => sellerApi.reportReview(reviewId.trim(), { reason: "inappropriate" }),
  })
  const items = asArray((reviews.data as Record<string, unknown> | undefined)?.items ?? reviews.data) ?? []
  const selected = items.find((it) => asString((it as Record<string, unknown>).id) === reviewId.trim()) ?? null

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">Reviews</h1>
      <SellerSectionCard title="Review list">
        <div className="space-y-2">
          {items.map((it, idx) => {
            const rec = it as Record<string, unknown>
            const id = asString(rec.id) ?? `${idx}`
            return (
              <button
                type="button"
                key={id}
                className="w-full rounded border p-3 text-left"
                onClick={() => setReviewId(id)}
              >
                <div className="font-medium">{asString(rec.title) ?? asString(rec.comment) ?? "Review"}</div>
                <div className="text-xs text-muted-foreground">{id}</div>
              </button>
            )
          })}
        </div>
      </SellerSectionCard>
      <SellerSectionCard title="Reply / report review">
        <div className="grid gap-2 md:grid-cols-3">
          <Input placeholder="Review ID" value={reviewId} onChange={(e) => setReviewId(e.target.value)} />
          <Input placeholder="Reply text" value={reply} onChange={(e) => setReply(e.target.value)} />
          <div className="flex gap-2">
            <Button onClick={() => replyMutation.mutate()} disabled={!reviewId.trim() || !reply.trim() || replyMutation.isPending}>
              Reply
            </Button>
            <Button variant="outline" onClick={() => reportMutation.mutate()} disabled={!reviewId.trim() || reportMutation.isPending}>
              Report
            </Button>
          </div>
        </div>
      </SellerSectionCard>
      {selected ? <SellerSectionCard title="Review detail"><SellerJsonView value={selected} /></SellerSectionCard> : null}
      {replyMutation.data ? <SellerJsonView value={replyMutation.data} /> : null}
      {reportMutation.data ? <SellerJsonView value={reportMutation.data} /> : null}
    </div>
  )
}

