import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"

export function PagePlaceholder({
  title,
  description,
  phase,
  notes,
}: {
  title: string
  description?: string
  phase?: string
  notes?: string[]
}) {
  return (
    <div className="mx-auto flex w-full max-w-screen-lg flex-col gap-4 px-4 py-8 lg:px-6">
      <header className="flex flex-col gap-2">
        <div className="flex items-center gap-2">
          <h1 className="font-heading text-xl font-semibold tracking-tight">{title}</h1>
          {phase ? <Badge variant="muted">{phase}</Badge> : null}
        </div>
        {description ? (
          <p className="text-sm text-muted-foreground">{description}</p>
        ) : null}
      </header>

      <Card>
        <CardHeader>
          <CardTitle>Phase 0 stub</CardTitle>
          <CardDescription>
            This page is part of the route map. Implementation arrives in later phases.
          </CardDescription>
        </CardHeader>
        {notes && notes.length > 0 ? (
          <CardContent className="text-sm text-muted-foreground">
            <ul className="list-disc pl-5">
              {notes.map((n) => (
                <li key={n}>{n}</li>
              ))}
            </ul>
          </CardContent>
        ) : null}
      </Card>
    </div>
  )
}
