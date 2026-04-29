export async function uploadToPresignedUrl(args: {
  url: string
  file: File | Blob
  contentType?: string
}): Promise<void> {
  const res = await fetch(args.url, {
    method: "PUT",
    headers: args.contentType ? { "Content-Type": args.contentType } : undefined,
    body: args.file,
  })
  if (!res.ok) {
    throw new Error(`Upload failed (${res.status})`)
  }
}

