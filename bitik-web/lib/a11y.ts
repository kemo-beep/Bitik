export const A11Y = {
  skipLinkId: "main-content",
  liveRegionId: "app-live-region",
} as const

export function srOnly(text: string) {
  return { className: "sr-only", children: text }
}

export function visuallyHiddenLabelProps(label: string) {
  return {
    "aria-label": label,
  }
}

export function pressKeyHandler<E extends { key: string; preventDefault: () => void }>(
  keys: string[],
  handler: () => void
) {
  return (event: E) => {
    if (keys.includes(event.key)) {
      event.preventDefault()
      handler()
    }
  }
}
