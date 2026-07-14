# feat: Migrate docs site Docsify → VitePress + SEO / LLM SEO

**Date:** 2026-07-14  
**Status:** **Implemented** (branch `feat/marketing-seo`)  
**Canonical context:** Docs site architecture, crawlable URLs, AI-readable surface.

---

## Problem frame

The public docs at [openflagr.github.io/flagr](https://openflagr.github.io/flagr) are a **Docsify SPA** with hash routes (`#/flagr_overview`). Search engines and LLM crawlers mostly see a shell (`Loading …`), not page content. There is no `robots.txt`, `sitemap.xml`, `llms.txt`, per-page meta, or JSON-LD.

**Success criteria:**

1. Real paths (e.g. `/flagr/flagr_overview.html` or clean `/flagr/flagr_overview`) with SSR/static HTML content.
2. `robots.txt`, `sitemap` (VitePress built-in), `llms.txt` published at site root.
3. Site-wide + homepage JSON-LD (`Organization`, `SoftwareApplication`, `WebSite`); rich default OG/meta.
4. `api_docs/` still at `/flagr/api_docs` (ReDoc + `bundle.yaml`).
5. `make serve-docs` / `make build-docs`; GitHub Pages deploys VitePress `dist`.
6. Hash-route redirects from old Docsify URLs.
7. Visual language preserved (warm zinc + steel-blue accent, Geist if feasible).

---

## As-built (target)

| Piece | Location |
|-------|----------|
| VitePress app | `docs/` + `docs/.vitepress/` |
| Content | `docs/*.md` (`index.md` = home) |
| Theme | VitePress default (light); `docs/.vitepress/theme/index.ts` re-export only |
| Static public | `docs/public/` (`robots.txt`, `llms.txt`, `images/`, favicon) |
| OpenAPI artifacts | `docs/api_docs/` (unchanged path for `make swagger`); **copied into dist** on build |
| Plans (internal) | `docs/plans/` — **excluded** from VitePress source |
| Deploy | `.github/workflows/pages.yml` → upload `docs/.vitepress/dist` |
| Dev | `make serve-docs` → `npm run docs:dev` |
| Prod build | `make build-docs` → `npm run docs:build` + copy `api_docs` |

**Base URL:** `/flagr/` (GitHub project pages).

---

## Implementation units

### U1. Plan + scaffold

- This plan file.
- `docs/package.json`, lockfile, `.vitepress/config.mts`, theme entry.
- `base: '/flagr/'`, `title`, `description`, `cleanUrls`, search, sidebar, nav, mermaid.

### U2. Content migration

- `home.md` → `index.md`.
- Keep page filenames (`flagr_overview.md`, …) so public paths stay predictable and README links map cleanly after dropping `#/`.
- `srcExclude`: `plans/**`, `api_docs/**`, theme leftovers.
- Image paths stay `/images/...`; files live under `docs/public/images/`.
- Port mermaid fenced blocks via `vitepress-plugin-mermaid` (or equivalent).

### U3. Theme + CSS

- Port tokens/rules from `flagr-docs.css` to VitePress selectors (`.VPDoc`, sidebar, content).
- Drop Docsify-only chrome; keep content density and accent family.

### U4. SEO / LLM assets

- `public/robots.txt` — allow crawlers + AI search bots; point to sitemap.
- `public/llms.txt` — product summary + key doc links (clean paths).
- Config `head`: OG/Twitter defaults, theme-color, canonical base awareness.
- JSON-LD via `transformHead` / layout: `Organization`, `WebSite`, `SoftwareApplication`.
- Optional FAQ JSON-LD later per page frontmatter (not required for v1).
- Hash redirect snippet for `#/...` → clean path.

### U5. Build, Makefile, Pages, docs pointers

- `make serve-docs` / `make build-docs` (local npm in `docs/`, no global docsify).
- `pages.yml`: Node setup → `make build-docs` → upload dist.
- Update CONTRIBUTING, README docs table, `docs/theme/README.md`, `AGENTS.md` if needed.
- Remove Docsify `docs/index.html`, `_sidebar.md`, obsolete `flagr-docs.css` after port.

### U6. Verify

- `make build-docs` succeeds.
- Dist contains `index.html` with real content (not only “Loading”), `robots.txt`, `llms.txt`, `api_docs/`.
- Spot-check sidebar, images, mermaid, external API link.

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
| Broken external `#/` links | Hash redirect + keep similar slugs |
| `api_docs` missing after build | Explicit copy step in `build-docs` |
| Theme drift | Port CSS tokens 1:1 first, polish later |
| Plans published | `srcExclude` + do not link in sidebar |

---

## Verification checklist

- [ ] `cd docs && npm ci && npm run docs:build`
- [ ] Dist has crawlable HTML for each sidebar page
- [ ] `/api_docs/` present with ReDoc
- [ ] `robots.txt` + `llms.txt` + sitemap
- [ ] JSON-LD present in homepage head
- [ ] Old `#/flagr_overview` redirects to clean URL
- [ ] README / CONTRIBUTING point at VitePress workflow
