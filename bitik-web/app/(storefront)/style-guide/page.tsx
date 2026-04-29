import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Separator } from "@/components/ui/separator"
import { StatusBadge } from "@/components/ui/status-badge"
import { ErrorState } from "@/components/ui/error-state"
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty"
import { ImageGallery } from "@/components/ui/image-gallery"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { InboxIcon, InfoIcon } from "lucide-react"
import { formatMoney, formatDate, formatRelative } from "@/lib/format"
import {
  ORDER_STATUS,
  PAYMENT_STATUS,
  SHIPMENT_STATUS,
  SELLER_STATUS,
  PRODUCT_STATUS,
  REFUND_STATUS,
  RETURN_STATUS,
} from "@/lib/status"

export const metadata = { title: "Style guide" }

const SAMPLE_DATE = new Date("2026-04-28T12:00:00Z")
const SAMPLE_DATE_RECENT = new Date("2026-04-28T06:00:00Z")
const SAMPLE_DATE_30M = new Date("2026-04-28T11:30:00Z")

function Section({
  heading,
  children,
}: {
  heading: string
  children: React.ReactNode
}) {
  return (
    <section className="flex flex-col gap-3">
      <h2 className="font-heading text-base font-semibold tracking-tight">{heading}</h2>
      <div>{children}</div>
    </section>
  )
}

export default function StyleGuidePage() {
  return (
    <div className="mx-auto flex w-full max-w-screen-lg flex-col gap-10 px-4 py-8 lg:px-6">
      <header className="flex flex-col gap-1">
        <h1 className="font-heading text-2xl font-semibold tracking-tight">Bitik design system</h1>
        <p className="text-sm text-muted-foreground">
          Tokens, formatters, and shared components used across storefront, seller center, and admin console.
        </p>
      </header>

      <Section heading="Buttons">
        <div className="flex flex-wrap items-center gap-2">
          <Button>Primary</Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="outline">Outline</Button>
          <Button variant="ghost">Ghost</Button>
          <Button variant="destructive">Destructive</Button>
          <Button variant="link">Link</Button>
          <Button size="sm">Small</Button>
          <Button size="lg">Large</Button>
          <Button disabled>Disabled</Button>
        </div>
      </Section>

      <Section heading="Inputs">
        <div className="grid max-w-md gap-3">
          <div className="grid gap-1">
            <Label htmlFor="sg-input">Email</Label>
            <Input id="sg-input" type="email" placeholder="you@example.com" />
          </div>
          <div className="grid gap-1">
            <Label htmlFor="sg-textarea">Message</Label>
            <Textarea id="sg-textarea" placeholder="Tell us about your shop" />
          </div>
        </div>
      </Section>

      <Section heading="Status badges">
        <div className="flex flex-col gap-3">
          <Row label="Order">
            {ORDER_STATUS.map((s) => (
              <StatusBadge key={s} kind="order" value={s} />
            ))}
          </Row>
          <Row label="Payment">
            {PAYMENT_STATUS.map((s) => (
              <StatusBadge key={s} kind="payment" value={s} />
            ))}
          </Row>
          <Row label="Shipment">
            {SHIPMENT_STATUS.map((s) => (
              <StatusBadge key={s} kind="shipment" value={s} />
            ))}
          </Row>
          <Row label="Seller">
            {SELLER_STATUS.map((s) => (
              <StatusBadge key={s} kind="seller" value={s} />
            ))}
          </Row>
          <Row label="Product">
            {PRODUCT_STATUS.map((s) => (
              <StatusBadge key={s} kind="product" value={s} />
            ))}
          </Row>
          <Row label="Refund">
            {REFUND_STATUS.map((s) => (
              <StatusBadge key={s} kind="refund" value={s} />
            ))}
          </Row>
          <Row label="Return">
            {RETURN_STATUS.map((s) => (
              <StatusBadge key={s} kind="return" value={s} />
            ))}
          </Row>
        </div>
      </Section>

      <Section heading="Badges">
        <div className="flex flex-wrap gap-2">
          <Badge>Default</Badge>
          <Badge variant="secondary">Secondary</Badge>
          <Badge variant="outline">Outline</Badge>
          <Badge variant="muted">Muted</Badge>
          <Badge variant="success">Success</Badge>
          <Badge variant="info">Info</Badge>
          <Badge variant="warning">Warning</Badge>
          <Badge variant="destructive">Destructive</Badge>
          <Badge variant="pending">Pending</Badge>
        </div>
      </Section>

      <Section heading="Cards">
        <div className="grid gap-3 sm:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Order #B-1029</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-2 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Total</span>
                <span className="font-medium">{formatMoney(48900, { currency: "MMK" })}</span>
              </div>
              <Separator />
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Placed</span>
                <span>{formatRelative(SAMPLE_DATE_RECENT, { now: SAMPLE_DATE })}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Status</span>
                <StatusBadge kind="order" value="processing" />
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Skeleton</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-2">
              <Skeleton className="h-4 w-3/4" />
              <Skeleton className="h-4 w-1/2" />
              <Skeleton className="h-24 w-full" />
            </CardContent>
          </Card>
        </div>
      </Section>

      <Section heading="Empty and error states">
        <div className="grid gap-3 sm:grid-cols-2">
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <InboxIcon />
              </EmptyMedia>
              <EmptyTitle>No orders yet</EmptyTitle>
              <EmptyDescription>Place your first order to see it here.</EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Button size="sm">Browse products</Button>
            </EmptyContent>
          </Empty>
          <ErrorState description="We could not load your orders." />
        </div>
      </Section>

      <Section heading="Alerts">
        <Alert>
          <InfoIcon />
          <AlertTitle>Manual Wave payment pending</AlertTitle>
          <AlertDescription>
            Send the exact amount to the Wave business account, then wait for admin confirmation.
          </AlertDescription>
        </Alert>
      </Section>

      <Section heading="Image gallery">
        <ImageGallery
          className="max-w-md"
          images={[
            { src: "https://placehold.co/600x600/png?text=1", alt: "Sample 1" },
            { src: "https://placehold.co/600x600/png?text=2", alt: "Sample 2" },
            { src: "https://placehold.co/600x600/png?text=3", alt: "Sample 3" },
          ]}
        />
      </Section>

      <Section heading="Formatters">
        <ul className="grid gap-1 text-sm text-muted-foreground sm:grid-cols-2">
          <li>Money: <span className="text-foreground">{formatMoney(125000, { currency: "MMK" })}</span></li>
          <li>Money compact: <span className="text-foreground">{formatMoney(125000000, { currency: "MMK", compact: true })}</span></li>
          <li>Date: <span className="text-foreground">{formatDate(SAMPLE_DATE)}</span></li>
          <li>Relative: <span className="text-foreground">{formatRelative(SAMPLE_DATE_30M, { now: SAMPLE_DATE })}</span></li>
        </ul>
      </Section>
    </div>
  )
}

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="w-20 shrink-0 text-xs font-medium text-muted-foreground uppercase">
        {label}
      </span>
      <div className="flex flex-wrap gap-1.5">{children}</div>
    </div>
  )
}
