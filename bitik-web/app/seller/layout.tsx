import { DashboardShell } from "@/components/shared/dashboard-shell"
import { SellerSidebar } from "@/components/seller/sidebar"
import { routes } from "@/lib/routes"
import { SellerRouteGuard } from "@/components/auth/seller-route-guard"
import Link from "next/link"
import { UnreadIndicator } from "@/components/notifications/unread-indicator"

export default function SellerLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <SellerRouteGuard>
      <DashboardShell
        brand="Bitik · Seller"
        brandHref={routes.seller.dashboard}
        sidebar={<SellerSidebar />}
        topbar={
          <Link href="/seller/notifications" className="text-xs text-muted-foreground hover:text-foreground">
            <UnreadIndicator scope="seller" />
          </Link>
        }
      >
        {children}
      </DashboardShell>
    </SellerRouteGuard>
  )
}
