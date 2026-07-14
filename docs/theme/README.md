# Docs site theme

- **Stack:** VitePress (`docs/.vitepress/`)
- **Theme:** stock [VitePress default theme](https://vitepress.dev/guide/default-theme-config), **light only** (`appearance: false`)
- **Config / sidebar / nav / SEO head:** [`../.vitepress/config.mts`](../.vitepress/config.mts)
- **Theme entry:** [`../.vitepress/theme/index.ts`](../.vitepress/theme/index.ts) (re-exports default; no custom CSS)
- **Public assets:** `../public/` (`robots.txt`, `llms.txt`, `images/`, favicon)
- **Local / CI commands** (from repo root):
  - `make serve-docs` — `docs-sync-snippets` + `npm ci` + VitePress dev on **http://127.0.0.1:8080/flagr/**
  - `make build-docs` — same sync + `npm ci` + build → `docs/.vitepress/dist` (copies `api_docs/`, patches sitemap)
  - `make docs-sync-snippets` — `pkg/config/env.go` → `docs/snippets/env.go`
- **As-built plan:** [`../plans/2026-07-14-001-vitepress-docs-seo-plan.md`](../plans/2026-07-14-001-vitepress-docs-seo-plan.md)
