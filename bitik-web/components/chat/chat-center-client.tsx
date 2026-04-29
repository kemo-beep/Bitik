"use client"

import { useMemo, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  createChatConversation,
  createMediaPresignedUrl,
  deleteChatConversation,
  deleteChatMessage,
  listChatConversations,
  listChatMessages,
  markChatConversationRead,
  sendChatMessage,
  completeMediaPresignedUpload,
} from "@/lib/api/buyer"
import { sellerApi } from "@/lib/api/seller"
import { queryKeys } from "@/lib/queryKeys"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { asArray, asRecord, asString } from "@/lib/safe"
import { useRealtimeChannel } from "@/lib/realtime/use-realtime-channel"

export function ChatCenterClient({ scope }: { scope: "buyer" | "seller" }) {
  const qc = useQueryClient()
  const [selectedConversationId, setSelectedConversationId] = useState("")
  const [message, setMessage] = useState("")
  const [targetId, setTargetId] = useState("")

  const conversations = useQuery({
    queryKey: scope === "buyer" ? queryKeys.buyer.chatConversations() : queryKeys.seller.chatConversations(),
    queryFn: () => (scope === "buyer" ? listChatConversations() : sellerApi.listChatConversations()),
    refetchInterval: 20_000,
  })

  const messages = useQuery({
    queryKey: scope === "buyer" ? queryKeys.buyer.chatMessages(selectedConversationId) : queryKeys.seller.chatMessages(selectedConversationId),
    queryFn: () =>
      scope === "buyer"
        ? listChatMessages(selectedConversationId)
        : sellerApi.listChatMessages(selectedConversationId),
    enabled: !!selectedConversationId,
    refetchInterval: 10_000,
  })

  useRealtimeChannel(
    "chat",
    () => {
      if (scope === "buyer") {
        qc.invalidateQueries({ queryKey: queryKeys.buyer.chatConversations() })
        if (selectedConversationId) qc.invalidateQueries({ queryKey: queryKeys.buyer.chatMessages(selectedConversationId) })
      } else {
        qc.invalidateQueries({ queryKey: queryKeys.seller.chatConversations() })
        if (selectedConversationId) qc.invalidateQueries({ queryKey: queryKeys.seller.chatMessages(selectedConversationId) })
      }
    },
    true
  )

  const createConversation = useMutation({
    mutationFn: async () =>
      scope === "buyer"
        ? createChatConversation({ seller_id: targetId })
        : sellerApi.createChatConversation({ buyer_id: targetId }),
    onSuccess: (data) => {
      setSelectedConversationId(asString((data as Record<string, unknown>).id) ?? "")
      if (scope === "buyer") qc.invalidateQueries({ queryKey: queryKeys.buyer.chatConversations() })
      else qc.invalidateQueries({ queryKey: queryKeys.seller.chatConversations() })
    },
  })

  const send = useMutation({
    mutationFn: async (payload: { message?: string; attachment_url?: string }) =>
      scope === "buyer"
        ? sendChatMessage(selectedConversationId, payload)
        : sellerApi.sendChatMessage(selectedConversationId, payload),
    onSuccess: () => {
      setMessage("")
      if (scope === "buyer") qc.invalidateQueries({ queryKey: queryKeys.buyer.chatMessages(selectedConversationId) })
      else qc.invalidateQueries({ queryKey: queryKeys.seller.chatMessages(selectedConversationId) })
    },
  })

  const markRead = useMutation({
    mutationFn: async () =>
      scope === "buyer" ? markChatConversationRead(selectedConversationId) : sellerApi.markChatRead(selectedConversationId),
    onSuccess: () => {
      if (scope === "buyer") qc.invalidateQueries({ queryKey: queryKeys.buyer.chatConversations() })
      else qc.invalidateQueries({ queryKey: queryKeys.seller.chatConversations() })
    },
  })

  const removeConversation = useMutation({
    mutationFn: async () =>
      scope === "buyer" ? deleteChatConversation(selectedConversationId) : sellerApi.deleteChatConversation(selectedConversationId),
    onSuccess: () => {
      setSelectedConversationId("")
      if (scope === "buyer") qc.invalidateQueries({ queryKey: queryKeys.buyer.chatConversations() })
      else qc.invalidateQueries({ queryKey: queryKeys.seller.chatConversations() })
    },
  })

  const list = asArray((conversations.data as Record<string, unknown> | undefined)?.items ?? conversations.data) ?? []
  const messageList = useMemo(
    () => asArray((messages.data as Record<string, unknown> | undefined)?.items ?? messages.data) ?? [],
    [messages.data]
  )

  return (
    <div className="mx-auto grid max-w-screen-xl gap-4 px-4 py-8 md:grid-cols-[320px_1fr]">
      <section className="rounded-xl border p-4">
        <h1 className="font-heading text-xl font-semibold">Conversations</h1>
        <div className="mt-3 flex gap-2">
          <Input
            placeholder={scope === "buyer" ? "Seller ID" : "Buyer ID"}
            value={targetId}
            onChange={(e) => setTargetId(e.target.value)}
          />
          <Button onClick={() => createConversation.mutate()} disabled={!targetId.trim() || createConversation.isPending}>
            Start
          </Button>
        </div>
        <div className="mt-4 space-y-2">
          {list.map((c, idx) => {
            const rec = asRecord(c) ?? {}
            const id = asString(rec.id) ?? `${idx}`
            return (
              <button
                key={id}
                className="w-full rounded border px-3 py-2 text-left text-sm hover:bg-muted"
                onClick={() => setSelectedConversationId(id)}
                type="button"
              >
                <div className="font-medium">{id}</div>
                <div className="text-xs text-muted-foreground">Unread: {asString(rec.unread_count) ?? "0"}</div>
              </button>
            )
          })}
        </div>
      </section>

      <section className="rounded-xl border p-4">
        <h2 className="font-medium">Conversation detail</h2>
        {!selectedConversationId ? (
          <p className="mt-2 text-sm text-muted-foreground">Select or create a conversation.</p>
        ) : (
          <>
            <div className="mt-3 flex flex-wrap gap-2">
              <Button variant="outline" onClick={() => markRead.mutate()} disabled={markRead.isPending}>
                Mark read
              </Button>
              <Button variant="destructive" onClick={() => removeConversation.mutate()} disabled={removeConversation.isPending}>
                Delete conversation
              </Button>
            </div>
            <div className="mt-4 max-h-80 space-y-2 overflow-auto rounded border p-3">
              {messageList.map((msg, idx) => {
                const rec = asRecord(msg) ?? {}
                const id = asString(rec.id) ?? `${idx}`
                return (
                  <div key={id} className="rounded bg-muted p-2 text-sm">
                    <div>{asString(rec.message) ?? "(attachment only)"}</div>
                    {asString(rec.attachment_url) ? (
                      <a className="text-xs text-primary underline" href={asString(rec.attachment_url) ?? "#"} target="_blank" rel="noreferrer">
                        Attachment
                      </a>
                    ) : null}
                    <div className="mt-1">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={async () => {
                          if (scope === "buyer") await deleteChatMessage(selectedConversationId, id)
                          else await sellerApi.deleteChatMessage(selectedConversationId, id)
                          if (scope === "buyer") qc.invalidateQueries({ queryKey: queryKeys.buyer.chatMessages(selectedConversationId) })
                          else qc.invalidateQueries({ queryKey: queryKeys.seller.chatMessages(selectedConversationId) })
                        }}
                      >
                        Delete message
                      </Button>
                    </div>
                  </div>
                )
              })}
            </div>
            <div className="mt-4 flex gap-2">
              <Input value={message} onChange={(e) => setMessage(e.target.value)} placeholder="Type a message" />
              <Button onClick={() => send.mutate({ message })} disabled={!message.trim() || send.isPending}>
                Send
              </Button>
              <Button
                variant="outline"
                onClick={async () => {
                  const prepared = await createMediaPresignedUrl({
                    filename: "chat-image.png",
                    content_type: "image/png",
                    size_bytes: 1000,
                  })
                  const file = asRecord(prepared.file) ?? {}
                  const uploadUrl = asString(prepared.upload_url)
                  if (!uploadUrl) return
                  await fetch(uploadUrl, {
                    method: "PUT",
                    headers: { "Content-Type": "image/png" },
                    body: new Blob(["chat-attachment"], { type: "image/png" }),
                  })
                  const fileId = asString(file.id)
                  if (!fileId) return
                  const ready = await completeMediaPresignedUpload(fileId)
                  const attachmentUrl = asString((ready as Record<string, unknown>).url)
                  if (!attachmentUrl) return
                  send.mutate({ attachment_url: attachmentUrl })
                }}
              >
                Attachment
              </Button>
            </div>
          </>
        )}
      </section>
    </div>
  )
}

