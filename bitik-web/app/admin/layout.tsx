import { DashboardShell } from "@/components/shared/dashboard-shell"
import { AdminSidebar } from "@/components/admin/sidebar"
import { routes } from "@/lib/routes"
import { RequireAnyRole } from "@/components/auth/guard"
import Link from "next/link"
import { UnreadIndicator } from "@/components/notifications/unread-indicator"

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <RequireAnyRole anyOf={["admin", "staff"]}>
      <DashboardShell
        brand="Bitik · Admin"
        brandHref={routes.admin.dashboard}
        sidebar={<AdminSidebar />}
        topbar={
          <Link href="/admin/notifications" className="text-xs text-muted-foreground hover:text-foreground">
            <UnreadIndicator scope="admin" />
          </Link>
        }
      >
        {children}
      </DashboardShell>
    </RequireAnyRole>
  )
}
