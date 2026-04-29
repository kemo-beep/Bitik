import { describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { ErrorState } from "@/components/ui/error-state"

describe("ErrorState", () => {
  it("invokes onRetry when Try again is clicked", async () => {
    const user = userEvent.setup()
    const onRetry = vi.fn()
    render(<ErrorState title="Failed" description="Try again later." onRetry={onRetry} />)
    await user.click(screen.getByRole("button", { name: "Try again" }))
    expect(onRetry).toHaveBeenCalledTimes(1)
  })
})
