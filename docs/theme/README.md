# Docs site theme

- **Stack:** VitePress (`docs/.vitepress/`)
- **Theme:** stock [VitePress default theme](https://vitepress.dev/guide/default-theme-config), **light only** (`appearance: false`)
- **Config / sidebar / nav / SEO head:** [`../.vitepress/config.mts`](../.vitepress/config.mts)
- **Theme entry:** [`../.vitepress/theme/index.ts`](../.vitepress/theme/index.ts) (re-exports default; no custom CSS)
- **Public assets:** `../public/` (`robots.txt`, `llms.txt`, `images/`, favicon)
- **Preview / build:** `make serve-docs` / `make build-docs` from repo root
