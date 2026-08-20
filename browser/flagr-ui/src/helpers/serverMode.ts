import { ref } from 'vue'
import { getHealth } from '@/api/health'

/**
 * Whether the server runs in eval-only (read-only) mode — the json_file /
 * json_http GitOps drivers. Single source of truth for the UI: chrome hides
 * write affordances from this ref, and crud reads derive the data plane
 * from it. The backend independently rejects writes with 403.
 */
export const evalOnlyMode = ref(false)

/**
 * Fetch the server mode once at app start. Fail-open: an unreachable health
 * endpoint, an aborted wait, or an older server without the evalOnlyMode
 * field renders the normal editable UI — a broken health check must not
 * lock the UI.
 */
export async function initServerMode(signal?: AbortSignal): Promise<void> {
  const res = await getHealth(signal)
  evalOnlyMode.value = res.ok && res.value?.evalOnlyMode === true
}
