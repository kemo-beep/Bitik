"use client"

export type RealtimeMessage = {
  type: string
  data?: Record<string, unknown>
}

type TokenProvider = () => Promise<string | null>

export class RealtimeWsClient {
  private ws: WebSocket | null = null
  private retryAttempt = 0
  private retryTimer: ReturnType<typeof setTimeout> | null = null
  private closedByUser = false

  constructor(
    private readonly url: string,
    private readonly getToken: TokenProvider,
    private readonly onMessage: (event: RealtimeMessage) => void,
    private readonly onStatus?: (status: "connecting" | "open" | "closed") => void
  ) {}

  async connect() {
    this.closedByUser = false
    this.onStatus?.("connecting")
    const token = await this.getToken()
    if (!token) {
      this.scheduleReconnect()
      return
    }
    const wsUrl = new URL(this.url)
    wsUrl.searchParams.set("token", token)
    this.ws = new WebSocket(wsUrl.toString())
    this.ws.onopen = () => {
      this.retryAttempt = 0
      this.onStatus?.("open")
    }
    this.ws.onmessage = (event) => {
      try {
        const parsed = JSON.parse(String(event.data)) as RealtimeMessage
        this.onMessage(parsed)
      } catch {
        // ignore malformed frames
      }
    }
    this.ws.onclose = () => {
      this.onStatus?.("closed")
      if (!this.closedByUser) this.scheduleReconnect()
    }
    this.ws.onerror = () => {
      this.ws?.close()
    }
  }

  disconnect() {
    this.closedByUser = true
    if (this.retryTimer) {
      clearTimeout(this.retryTimer)
      this.retryTimer = null
    }
    this.ws?.close()
    this.ws = null
    this.onStatus?.("closed")
  }

  private scheduleReconnect() {
    const delay = Math.min(30_000, 1_000 * 2 ** this.retryAttempt)
    this.retryAttempt += 1
    this.retryTimer = setTimeout(() => {
      void this.connect()
    }, delay)
  }
}

