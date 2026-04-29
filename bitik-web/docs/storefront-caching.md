## Storefront caching (Phase 3)

### What we cache

In Phase 3, **public storefront data is cached in TanStack Query** (client-side cache), not in Next “server fetch” caching.

- **Why**: most Phase 3 storefront pages are interactive client components (filters, URL state, pagination).
- **Trade-off**: the first request is still network-bound, but subsequent navigations within the session reuse cached data.

### Default query settings

Most public catalog queries use:

- **`staleTime`**: ~60s (fresh enough for browsing, reduces refetch spam)
- **`gcTime`**: 10–30 minutes (keeps recently visited lists/details around)
- **`retry`**: 1 (avoid long retry loops for end-users)

You’ll find these settings directly in the page/client components using `useQuery`.

### URL-driven caching keys

Paged lists use query keys that include the full filter/sort/page state (example: products list uses `queryKeys.public.listProducts(params)`), so navigating between pages/filters keeps results isolated and cacheable.

### When to switch to server rendering

If we want SEO-first rendering for top pages (home/products/search), we can later:

- implement server components that fetch public endpoints with `fetch()` and Next cache hints, and
- hydrate TanStack Query for client interactivity.

That change is intentionally deferred until Phase 4+ to keep Phase 3 implementation straightforward.

