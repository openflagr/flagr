# Flagr brand assets (option D)

Single approved identity: geometric **F monogram** + **FLAGR** wordmark.

## Source of truth (edit these)

| File | Use |
|------|-----|
| `flagr-mark.svg` | Icon monogram (`currentColor`) |
| `flagr-logo.svg` | Full lockup monogram + wordmark (`currentColor`) |

PNG derivatives are for favicon / Open Graph only (raster fallbacks).

| File | Use |
|------|-----|
| `favicon.png` | Browser tab icon |
| `apple-touch-icon.png` | iOS home screen |
| `flagr-logo.png` / `flagr-mark.png` | OG / crawlers that prefer raster |

## Installed copies

**Docs (VitePress)**

- Header logo: `docs/public/images/logo.svg`
- Monogram: `docs/public/images/logo-mark.svg`
- Favicon: `docs/public/favicon.png`

**UI (`browser/flagr-ui`)**

- Navbar: **inline SVG** monogram in `App.vue` (matches `flagr-mark.svg`, `currentColor` on primary bar)
- Public: `public/logo-mark.svg` kept in sync for reuse
- Favicon: `public/favicon.png`

When changing the monogram geometry, update `docs/brand/flagr-mark.svg` and the path data in `App.vue` together, then re-copy to `docs/public` / regenerate PNGs.
