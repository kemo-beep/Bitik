"use client"

import * as React from "react"
import { UploadCloudIcon, XIcon, FileIcon } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

export type FileUploadProps = {
  accept?: string
  multiple?: boolean
  maxFiles?: number
  maxSize?: number
  value?: File[]
  onChange?: (files: File[]) => void
  disabled?: boolean
  className?: string
  label?: string
  hint?: string
  id?: string
  name?: string
}

export function FileUpload({
  accept,
  multiple = false,
  maxFiles,
  maxSize,
  value,
  onChange,
  disabled,
  className,
  label = "Drop files here or click to browse",
  hint,
  id,
  name,
}: FileUploadProps) {
  const inputRef = React.useRef<HTMLInputElement>(null)
  const [internal, setInternal] = React.useState<File[]>([])
  const [dragOver, setDragOver] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const files = value ?? internal
  const generatedId = React.useId()
  const inputId = id ?? generatedId

  function commit(next: File[]) {
    if (value === undefined) setInternal(next)
    onChange?.(next)
  }

  function validate(list: File[]): { ok: File[]; err: string | null } {
    let err: string | null = null
    const ok: File[] = []
    for (const f of list) {
      if (maxSize && f.size > maxSize) {
        err = `${f.name} exceeds ${(maxSize / 1024 / 1024).toFixed(1)}MB`
        continue
      }
      ok.push(f)
    }
    if (maxFiles && ok.length + files.length > maxFiles) {
      err = `Only ${maxFiles} file${maxFiles === 1 ? "" : "s"} allowed`
      return { ok: ok.slice(0, Math.max(0, maxFiles - files.length)), err }
    }
    return { ok, err }
  }

  function handle(list: FileList | null) {
    if (!list) return
    const next = Array.from(list)
    const { ok, err } = validate(next)
    setError(err)
    commit(multiple ? [...files, ...ok] : ok.slice(0, 1))
  }

  function remove(idx: number) {
    commit(files.filter((_, i) => i !== idx))
    setError(null)
  }

  return (
    <div className={cn("flex flex-col gap-2", className)}>
      <button
        type="button"
        disabled={disabled}
        onClick={() => inputRef.current?.click()}
        onDragOver={(e) => {
          e.preventDefault()
          if (!disabled) setDragOver(true)
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={(e) => {
          e.preventDefault()
          setDragOver(false)
          if (!disabled) handle(e.dataTransfer.files)
        }}
        className={cn(
          "group/file-upload relative flex min-h-28 w-full flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-input bg-background px-4 py-6 text-sm text-muted-foreground transition-colors outline-none",
          "hover:bg-muted/40 focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/40",
          "disabled:cursor-not-allowed disabled:opacity-50",
          dragOver && "border-primary bg-primary/5 text-foreground"
        )}
        aria-describedby={hint ? `${inputId}-hint` : undefined}
      >
        <UploadCloudIcon className="size-5" aria-hidden="true" />
        <span className="font-medium text-foreground">{label}</span>
        {hint ? (
          <span id={`${inputId}-hint`} className="text-xs">{hint}</span>
        ) : null}
        <input
          ref={inputRef}
          id={inputId}
          name={name}
          type="file"
          accept={accept}
          multiple={multiple}
          disabled={disabled}
          className="sr-only"
          onChange={(e) => handle(e.target.files)}
        />
      </button>

      {error ? (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      ) : null}

      {files.length > 0 ? (
        <ul className="flex flex-col gap-1">
          {files.map((f, i) => (
            <li
              key={`${f.name}-${i}`}
              className="flex items-center gap-2 rounded-md border bg-background px-2 py-1.5 text-sm"
            >
              <FileIcon className="size-4 text-muted-foreground" aria-hidden="true" />
              <span className="flex-1 truncate">{f.name}</span>
              <span className="shrink-0 text-xs text-muted-foreground">
                {(f.size / 1024).toFixed(0)} KB
              </span>
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                onClick={() => remove(i)}
                aria-label={`Remove ${f.name}`}
              >
                <XIcon />
              </Button>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  )
}
