# Visual regression (Playwright) — deferred

Full visual regression (screenshot baselines in-repo) is **out of scope** for Phase 9 by default: baselines add storage cost, flaky diffs on fonts/OS, and CI maintenance.

## Adding `@visual` later

1. In Playwright config, enable `expect({ timeout }).toHaveScreenshot()` with a stable viewport and `mask` for dynamic timestamps.
2. Tag scenarios with `@visual`; run only on `main` or a scheduled workflow to reduce noise.
3. Store baselines in Git LFS or a dedicated artifact bucket; pin CI runner image for consistency.
4. Start with 1–2 critical pages (home, product detail) before expanding.

Until then, rely on functional E2E (`@smoke`, `@critical`) and manual design QA for visual changes.
