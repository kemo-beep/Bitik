## Web performance baseline (Phase 3)

### What we have in place

- **Web Vitals reporting hook**: `components/observability/web-vitals.tsx` (wired to the analytics stub and Sentry if enabled).
- **Storefront skeleton states**: list/detail pages render `Skeleton` components while loading.
- **`next/image` for key visuals**: homepage hero banner and seller/category/brand images.
- **Client caching**: public catalog queries use TanStack Query with short `staleTime` and moderate `gcTime`.

See also: `docs/storefront-caching.md`.

### Quick Lighthouse run (local)

1. Start the web app:

```bash
cd bitik-web
npm run dev -- -p 3100
```

2. Open Chrome DevTools → Lighthouse:
   - Mode: **Navigation**
   - Categories: **Performance**, **Accessibility**, **Best Practices**, **SEO**
   - Device: **Mobile**

3. Test key routes:
   - `/`
   - `/products`
   - `/search?q=test`
   - `/products/<id>` or `/p/<slug>`

### Recommended baseline targets (not enforced in CI yet)

- **LCP**: < 2.5s (mobile)
- **INP**: < 200ms
- **CLS**: < 0.1

### CI notes

Phase 3 does not gate merges on performance budgets. We only keep a baseline document and basic E2E coverage.

