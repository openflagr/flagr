# feat: Migrate docs site Docsify → VitePress + SEO / LLM SEO

**Date:** 2026-07-14  
**Status:** **Implemented** (branch `feat/marketing-seo`, [PR #745](https://github.com/openflagr/flagr/pull/745))  
**Canonical context:** Docs site architecture, crawlable URLs, AI-readable surface, local/CI build commands.

---

## Problem frame

The public docs at [openflagr.github.io/flagr](https://openflagr.github.io/flagr) were a **Docsify SPA** with hash routes (`#/flagr_overview`). Search engines and LLM crawlers mostly saw a shell (`Loading …`), not page content. There was no `robots.txt`, `sitemap.xml`, `llms.txt`, per-page meta, or JSON-LD.

**Success criteria (met):**

1. Real paths (clean `/flagr/flagr_overview`) with static HTML content.
2. `robots.txt`, `sitemap` (VitePress + `api_docs` post-build), `llms.txt` at site root.
3. Site-wide JSON-LD (`Organization`, `SoftwareApplication`, `WebSite`); OG/Twitter/canonical.
4. `api_docs/` still at `/flagr/api_docs` (ReDoc + `bundle.yaml`).
5. `make serve-docs` / `make build-docs`; GitHub Pages deploys VitePress `dist`.
6. Hash-route redirects from old Docsify URLs (`#/page?id=section` → `/flagr/page#section`).
7. Stock VitePress default **light** theme (custom Docsify CSS dropped).

---

## As-built

| Piece | Location / command |
|-------|--------------------|
| VitePress app | `docs/` + `docs/.vitepress/` |
| Content | `docs/*.md` (`index.md` = home) |
| Theme | VitePress default (light); `docs/.vitepress/theme/index.ts` re-exports only |
| Static public | `docs/public/` (`robots.txt`, `llms.txt`, `images/`, favicon, `.nojekyll`) |
| OpenAPI artifacts | `docs/api_docs/` (source of truth for `make swagger`); **copied into dist** on `build-docs` |
| Env source embed | `make docs-sync-snippets` copies `pkg/config/env.go` → `docs/snippets/env.go` (gitignored); `flagr_env.md` uses `<<< @/snippets/env.go{go}` |
| Plans (internal) | `docs/plans/` — **`srcExclude`**; link from site via GitHub blob URLs |
| Deploy | `.github/workflows/pages.yml` → Node 22 → `make build-docs` → upload `docs/.vitepress/dist` |
| Local dev | **`make serve-docs`** (see below) |
| Prod build | **`make build-docs`** (see below) |

**Base URL:** `/flagr/` (GitHub project pages).

### Docs dev / build instructions (source of truth)

Run from **repo root**. Targets live in the root `Makefile`.

| Command | What it does |
|---------|----------------|
| **`make serve-docs`** | `docs-sync-snippets` → `cd docs && npm ci && npm run docs:dev -- --port 8080 --host 127.0.0.1` → open **http://127.0.0.1:8080/flagr/** |
| **`make build-docs`** | `docs-sync-snippets` → `cd docs && npm ci && npm run docs:build` → copy `api_docs/` into dist → append `api_docs/` to `sitemap.xml` |
| **`make docs-sync-snippets`** | `cp pkg/config/env.go docs/snippets/env.go` (also a dependency of serve/build) |

**Details:**

- Dependencies: `docs/package.json` + **`docs/package-lock.json`** (lockfile-strict via **`npm ci`**).
- Output: `docs/.vitepress/dist/` (gitignored).
- Preview production build: `cd docs && npx vitepress preview` after `make build-docs` (optional).
- CI: Pages workflow only — `make build-docs` on push to `main` / `workflow_dispatch`; not part of unit-test `ci.yml`.
- Dead links: `ignoreDeadLinks: false` — broken internal links fail the build.
- Anchors: VitePress `{#slug}` on headings (not Docsify `:id=`).

---

## Implementation units (shipped)

### U1. Plan + scaffold

- This plan file.
- `docs/package.json`, lockfile, `.vitepress/config.mts`, theme entry.
- `base: '/flagr/'`, `title`, `description`, `cleanUrls`, search, sidebar, nav, mermaid.

### U2. Content migration

- `home.md` → `index.md`.
- Page filenames kept (`flagr_overview.md`, …).
- `srcExclude`: `plans/**`, `api_docs/**`, `theme/**`, `snippets/**`.
- Images under `docs/public/images/`.
- Mermaid via `vitepress-plugin-mermaid`.
- Docsify `:id=` → `{#id}`; redundant markdown `---` HRs removed (VitePress h2 borders).

### U3. Theme

- Stock VitePress default theme, light only (`appearance: false`).
- Custom Docsify CSS / Geist CDN not used.

### U4. SEO / LLM assets

- `public/robots.txt` — allow crawlers + AI search bots; block CCBot; sitemap URL.
- `public/llms.txt` — product summary + key doc links.
- Config `head`: OG/Twitter, theme-color, JSON-LD, canonical via `transformPageData`.
- Hash redirect: `#/page?id=section` → `/flagr/page#section`.

### U5. Build, Makefile, Pages, docs pointers

- `make serve-docs` / `make build-docs` / `make docs-sync-snippets` as above.
- `pages.yml`: Node 22 + npm cache on `docs/package-lock.json` → `make build-docs` → upload dist.
- CONTRIBUTING, README docs table, `docs/theme/README.md` updated.
- Docsify `index.html`, `_sidebar.md`, `flagr-docs.css` removed.

### U6. Verify

- `make build-docs` succeeds with dead-link checking.
- Dist has crawlable HTML, `robots.txt`, `llms.txt`, sitemap (+ api_docs), embedded env.go, `api_docs/`.

---

## Out of scope (follow-ups)

- Marketing pages (`/alternatives/*`, `/vs/*`) — separate plan.
- Custom domain / non-`/flagr/` base.
- Full OKF bundle.
- Per-page author/date schema for every guide.

---

## Risks

| Risk | Mitigation |
|------|------------|
| Broken external `#/` links | Hash redirect + stable `{#slug}` anchors |
| `api_docs` missing after build | Explicit copy in `build-docs` + sitemap patch |
| Stale env embed | `docs-sync-snippets` before every serve/build |
| Plans 404 on site | `srcExclude` + GitHub blob links from published pages |
| Theme / double dividers | Stock theme; no bare `---` HRs before h2 |

---

## Verification checklist

- [x] `make build-docs` (from repo root; uses `npm ci`)
- [x] Dist has crawlable HTML for each sidebar page
- [x] `/api_docs/` present with ReDoc
- [x] `robots.txt` + `llms.txt` + sitemap (incl. api_docs)
- [x] JSON-LD present in homepage head
- [x] Old `#/flagr_overview` / `?id=` redirects to clean URL + fragment
- [x] README / CONTRIBUTING point at VitePress workflow
- [x] `make serve-docs` → http://127.0.0.1:8080/flagr/
