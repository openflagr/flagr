# flagr-ui


Vue 3 + Vite. TypeScript: Vite build + `npm run typecheck` (`vue-tsc --noEmit`).

## Commands

| Command | Purpose |
|---------|---------|
| `npm install` | Dependencies |
| `npm run dev` | Vite dev server (repo: `make start` for backend + UI) |
| `npm run build` | Production → `dist/` |
| `npm run typecheck` | Typecheck only |
| `npm run lint` | ESLint |
| `npm run test:e2e` | Playwright (repo: `make test-e2e`) |

## Layout

`api/` · `pages/` · `components/` · `helpers/`

## Config

- `VITE_API_URL` (default `/api/v1`) — `helpers/constants.ts`