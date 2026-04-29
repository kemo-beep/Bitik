"use client"

import { Bar, BarChart, CartesianGrid, XAxis } from "recharts"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"

export function SellerSalesChart({ data }: { data: Array<{ label: string; value: number }> }) {
  return (
    <ChartContainer
      className="h-56 w-full"
      config={{ value: { label: "Value", color: "var(--primary)" } }}
    >
      <BarChart data={data}>
        <CartesianGrid vertical={false} />
        <XAxis dataKey="label" tickLine={false} axisLine={false} />
        <ChartTooltip cursor={false} content={<ChartTooltipContent />} />
        <Bar dataKey="value" radius={4} fill="var(--color-value)" />
      </BarChart>
    </ChartContainer>
  )
}

