# Docs site theme

- **Styles:** [`../flagr-docs.css`](../flagr-docs.css) (linked from [`../index.html`](../index.html))
- **Docsify:** `subMaxLevel: 0` in `index.html` hides the in-page heading sub-sidebar (sidebar + search only). Raise it if you want per-page TOC back.
- **Tokens:** edit `:root` (`--flagr-text-*`, `--content-max`, `--prism-*`, surface/hover/pre-border). Avoid new literal `rgba(...)` in rules — add a token instead.
- **Do not** add one-off `font-size` on `.markdown-section` rules unless you add a matching token.
- **Figures:** screenshots span the full content column (same width as `pre` and tables).