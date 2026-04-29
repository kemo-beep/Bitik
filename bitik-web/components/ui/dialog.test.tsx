import { describe, expect, it } from "vitest"
import { render, screen } from "@testing-library/react"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "./dialog"
import { Button } from "./button"

describe("Dialog", () => {
  it("renders title when dialog is open", () => {
    render(
      <Dialog defaultOpen>
        <DialogTrigger render={<Button />}>Open</DialogTrigger>
        <DialogContent showCloseButton={false}>
          <DialogHeader>
            <DialogTitle>Dialog smoke title</DialogTitle>
          </DialogHeader>
        </DialogContent>
      </Dialog>
    )
    expect(screen.getByText("Dialog smoke title")).toBeInTheDocument()
  })
})
