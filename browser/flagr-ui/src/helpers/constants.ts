// Relative by default so API calls respect a FLAGR_WEB_PREFIX subpath
// deployment; resolves to /api/v1 when served at the root.
const API_URL = import.meta.env.VITE_API_URL || 'api/v1'
const rawEntityTypes = import.meta.env.VITE_FLAGR_UI_POSSIBLE_ENTITY_TYPES
const FLAGR_UI_POSSIBLE_ENTITY_TYPES =
  rawEntityTypes && rawEntityTypes !== 'null' ? rawEntityTypes : null

// Snapshots shown in the History tab. Heavily-edited flags can accumulate
// hundreds of snapshots (each a full flag JSON), so fetching them all is slow
// and memory-heavy on the server. 0 means unlimited (fetch full history).
const DEFAULT_SNAPSHOT_HISTORY_LIMIT = 50
const rawSnapshotLimit = import.meta.env.VITE_FLAGR_UI_SNAPSHOT_HISTORY_LIMIT
const parsedSnapshotLimit = rawSnapshotLimit ? Number(rawSnapshotLimit) : NaN
const FLAGR_UI_SNAPSHOT_HISTORY_LIMIT =
  Number.isInteger(parsedSnapshotLimit) && parsedSnapshotLimit >= 0
    ? parsedSnapshotLimit
    : DEFAULT_SNAPSHOT_HISTORY_LIMIT

export default {
  API_URL,
  FLAGR_UI_POSSIBLE_ENTITY_TYPES,
  FLAGR_UI_SNAPSHOT_HISTORY_LIMIT,
}