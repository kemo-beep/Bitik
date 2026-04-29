#!/usr/bin/env node
/**
 * After `next build` with `output: "standalone"`, copy static assets next to the
 * standalone server so `node server.js` from `.next/standalone` resolves chunks.
 * @see https://nextjs.org/docs/app/api-reference/config/next-config-js/output
 */
import fs from "node:fs"
import path from "node:path"
import { fileURLToPath } from "node:url"

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..")
const standalone = path.join(root, ".next", "standalone")
const serverJs = path.join(standalone, "server.js")

if (!fs.existsSync(serverJs)) {
  console.error("prepare-standalone: missing .next/standalone/server.js — run npm run build first.")
  process.exit(1)
}

const staticSrc = path.join(root, ".next", "static")
const staticDest = path.join(standalone, ".next", "static")
const publicSrc = path.join(root, "public")
const publicDest = path.join(standalone, "public")

if (fs.existsSync(staticSrc)) {
  fs.mkdirSync(path.dirname(staticDest), { recursive: true })
  fs.cpSync(staticSrc, staticDest, { recursive: true })
}

if (fs.existsSync(publicSrc)) {
  fs.cpSync(publicSrc, publicDest, { recursive: true })
}
