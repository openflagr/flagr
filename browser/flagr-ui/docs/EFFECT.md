# Effect in flagr-ui

This document explains **how Effect is used today**, **why it fits this app**, **where the real power is**, and **how to extend usage** without turning Vue into an Effect runtime.

## Mental model

Effect describes **programs** that can fail in **known ways**:

```text
Effect.Effect<Success, Error, Requirements>
```

In flagr-ui we only use the first two type parameters:

| Piece | Meaning in this repo |
|--------|----------------------|
| **Success (`A`)** | JSON decoded type (`Flag`, `void`, …) |
| **Error (`E`)** | Our `ApiError` union (HTTP, network, decode, 401) |
| **Requirements (`R`)** | Always `never` — we do not use Layers/services yet |

A function in `src/api/` should return **`Effect.Effect<A, ApiError>`**, not `Promise<A>`.  
Vue never `await`s those programs directly; **`runApi`** runs them once at the UI edge.

```text
  Vue (pages/components)
        │  runApi(vm, program, { onSuccess, successMessage })
        ▼
  helpers/runApi.ts  ── Effect.runPromiseExit + Exit.match + toasts
        ▼
  api/flags.ts, api/evaluation.ts  ── compose requestJson / Effect.gen
        ▼
  api/http.ts  ── fetch + decode + map errors
        ▼
  api/errors.ts  ── Data.TaggedError
```

**Rule:** Keep Effect **inside `api/`** (and small compositions like `listFlagsIfStale`). Do not sprinkle `Effect.gen` in `.vue` files.

---

## What we use today (concrete)

### 1. Tagged errors (`api/errors.ts`)

Failures are **data**, not random `throw` types:

- `ApiHttpError` — 4xx/5xx with server message  
- `ApiUnauthorized` — 401 + optional redirect from `WWW-Authenticate`  
- `ApiNetworkError` — fetch failed, offline, etc.  
- `ApiDecodeError` — JSON parse failed  

`Data.TaggedError` gives a stable `_tag` for narrowing and for **`Match.valueTags`** in the UI.

### 2. HTTP as Effect (`api/http.ts`)

- `Effect.tryPromise` wraps `fetch` and maps unknown failures with `ensureApiError`.
- Non-OK responses: `yield* Effect.fail(mapResponseError(...))` — **typed failure**, not thrown exceptions in business code.
- Empty bodies (DELETE, 204): decode returns `undefined` without throwing.
- `Effect.fn('flagr.requestJson')` names the step for tracing/debugging.

Every route in `flags.ts` / `evaluation.ts` is a thin wrapper:  
`get/post/putVoid/del` → `requestJson` / `requestVoid`.

### 3. Composition (`api/flags.ts`)

`listFlagsIfStale` is the one **multi-step** program:

```ts
Effect.fn('flags.listFlagsIfStale')(function* (cachedMaxId) {
  const { maxID } = yield* getSnapshotMaxId()
  if (cachedMaxId !== undefined && maxID === cachedMaxId) return null
  const flags = yield* listFlags()
  return { flags: [...flags].reverse(), maxSnapshotID: maxID }
})
```

Sequential logic reads like async/await but **stays in the error channel** (`ApiError`).

### 4. Vue boundary (`helpers/runApi.ts`)

- **`Effect.runPromiseExit`** — never `runPromise` alone for UI work; we need the full `Exit` to handle **defects** and mixed failure causes.
- **`Exit.match`** — success → toast + `onSuccess`; failure → `Cause.squash` → `presentApiError`.
- **`Match.valueTags`** — exhaustive user-facing messages per `ApiError` tag.
- **`confirmAndRunApi`** — confirm dialog then `runApi` (still fire-and-forget from Vue’s perspective).

Pages (`pages/flagPage.ts`, `Flags.vue`) only call **`runApi` / `confirmAndRunApi`** with programs from `api/*`.

### Flag page: after mutations (`pages/flagPage.ts`)

| Action | UI update |
|--------|-----------|
| Create tag / variant / segment / constraint | Optimistic patch on `vm.flag` (create tag also refreshes `listAllTags` for autocomplete) |
| Delete tag | One `Effect.gen`: `deleteTag` → `loadFlagAndAllTags` |
| Delete variant / segment / constraint | `reloadFlag` |
| Mount | `loadFlagPageContext` — parallel GETs (skips `GET /flags/entity_types` when `VITE_FLAGR_UI_POSSIBLE_ENTITY_TYPES` is set); env overrides options in `applyEntityTypesToVm` |

---

## Why Effect is good *here* (not hype)

| Problem without Effect | What Effect gives you |
|------------------------|------------------------|
| `axios` + `handleErr` with string matching and rethrows | One **`ApiError`** union end-to-end |
| `try/catch` around every call | Failures are **`Effect.fail`**; success path stays linear in `gen` |
| Double toasts / lost 401 redirect | Single boundary: **`runApi`** + tagged 401 handling |
| “What can this call throw?” | Signature says: **`Effect<Flag, ApiError>`** |
| Empty DELETE body breaks JSON parse | Centralized in **`decodeJsonBody`** |

For a **CRUD admin UI** with ~30 endpoints, the win is **consistent error typing + one runtime interpreter (`runApi`)**, not micro-optimizations.

---

## Where the *real* power is (and what we have *not* turned on)

Effect’s depth is not “async syntax”. It is:

1. **Typed error channels** — you already use this (`ApiError`).
2. **Composition** — `yield*` chains programs without callback pyramids (`listFlagsIfStale`).
3. **Structured concurrency** — `Effect.all`, racing, timeouts, retries **across** steps with one failure model.
4. **Services & Layers** — swap HTTP client, base URL, or auth in tests without mocking `fetch` in every test.
5. **Testability** — run the same program with `Effect.runPromise` in Vitest and assert on `Exit` / `Effect.flip` for errors.
6. **Cross-cutting pipes** — `Effect.retry`, `Effect.timeout`, logging spans on **one** composed program.

We use **(1) lightly, (2) once**. We do **not** yet use (3)–(6) in production code.

That is intentional: Vue remains the app shell; Effect owns **I/O programs** only.

---

## How to apply Effect correctly (checklist for new code)

### Do

- Add new REST calls in `api/flags.ts` or `api/evaluation.ts` as  
  `(): Effect.Effect<T, ApiError> => requestJson(...)` or `requestVoid(...)`.
- Return **`Effect.fail(new ApiHttpError(...))`** (or other tags) for expected failures inside `api/`.
- Compose multi-step server workflows with **`Effect.gen`** or **`Effect.fn('name')`** in `api/`, not in Vue.
- Invoke from UI only via **`runApi(vm, program, options)`** or **`confirmAndRunApi`**.
- Add a new failure mode → **new `TaggedError` + branch in `apiErrorUserMessage`**.

### Don’t

- Call `fetch` from components or `pages/*` (keeps caching, auth, retries in one place).
- Use `Effect.runPromise` in components (bypasses `Exit` and defect handling).
- Use `Effect.catchAll` and erase `ApiError` unless you re-tag at the boundary.
- Put business rules that only touch in-memory Vue state inside Effect (normal functions in `helpers/flagModel.ts` are fine).

### Example: new endpoint

```ts
// api/flags.ts
export const archiveFlag = (flagId: FlagId): Effect.Effect<void, ApiError> =>
  putVoid(`${flag(flagId)}/archive`)
```

```ts
// pages/flagPage.ts
runApi(vm, flagsApi.archiveFlag(vm.flagId), { successMessage: 'Flag archived', onSuccess: () => reloadFlag(vm) })
```

---

## Where we *can* apply more Effect (sensible next steps)

Ordered by **value / effort** for flagr-ui:

### A. Parallel independent loads — **done for mount**

`loadFlagPageContext(flagId)` runs `Effect.all` on flag + tags (+ entity types only when env does not pin the list). `mountFlagPage` uses one `runApi`. Entity-type env override lives in `applyEntityTypesToVm` in `flagPage.ts`.

### B. Retries on transient failures (medium value, medium effort)

Wrap `requestJson` (or only GETs) with a schedule:

```ts
program.pipe(
  Effect.retry(Schedule.exponential('100 millis').pipe(Schedule.compose(Schedule.recurs(2)))),
)
```

Do this in **`api/http.ts`** or per-endpoint in `api/`, not in Vue.

### C. Unit tests for API programs (high value, medium effort)

Add `@effect/vitest` (or `Effect.runPromiseExit` in Vitest) and test:

- DELETE → empty body success  
- 401 → `ApiUnauthorized`  
- invalid JSON → `ApiDecodeError`  
- `listFlagsIfStale` → `null` when cache fresh  

No browser required; **`fetch` mocked once** at the HTTP layer.

### D. `FlagrApi` service + Layer (higher effort, best for larger teams)

Introduce a single service (e.g. `FlagrClient`) whose methods mirror `flags.ts`, with:

- `Layer` providing base URL + `fetch`  
- Test layer returning stubbed responses  

**When worth it:** many contributors, lots of API tests, or multiple apps sharing the same client.  
**When skip:** current SPA size is fine with module functions + `runApi`.

### E. Schema validation at the boundary (optional)

Decode JSON with `@effect/schema` (or keep manual types) so **`ApiDecodeError`** means “shape mismatch”, not just `JSON.parse` failure. Valuable if the backend schema drifts often.

### F. Usually *not* worth it here

- Rewriting Vue to `Effect.gen` everywhere  
- Replacing Element Plus confirm flows with Effect-only UX  
- Full `@effect/platform` HttpClient until you upgrade Effect major and want middleware stacks  

---

## Comparison: before vs after

| Before (axios era) | After (Effect) |
|--------------------|----------------|
| `axios.get` → catch → `handleErr` | `requestJson` → `ApiError` in type |
| String / status guessing | `_tag` + `Match.valueTags` |
| Errors as exceptions in UI | `runPromiseExit` + `Exit.match` once |
| Logic mixed in Vue methods | `api/*` programs + `pages/*` orchestration |

---

## References

- **Conventions (short):** root `AGENTS.md` (Frontend) — layout + `runApi` only.
- **Migration + Effect rules:** `docs/plans/2026-06-26-001-migrate-flagr-ui-js-to-ts-plan.md` (§ Effect).
- Effect v3: [effect.website](https://effect.website)
- Community skill (v4-oriented; map to v3): [joelhooks/effectts-skills](https://github.com/joelhooks/effectts-skills)