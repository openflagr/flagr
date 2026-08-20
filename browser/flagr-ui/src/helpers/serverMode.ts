import { ref } from 'vue'
import { getHealth } from '@/api/health'
import { setFlagReadSource } from '@/api/crud'

/**
 * Whether the server runs in eval-only (read-only) mode — the json_file /
 * json_http GitOps drivers. When true the UI hides write affordances; the
 * backend independently rejects writes with 403.
 */
export const evalOnlyMode = ref(false)

function applyMode(evalOnly: boolean): void {
  evalOnlyMode.value = evalOnly
  setFlagReadSource(evalOnly ? 'evalCache' : 'http')
}

/**
 * Fetch the server mode once at app start. Fail-open: an unreachable health
 * endpoint, an aborted wait, or an older server without the evalOnlyMode
 * field renders the normal editable UI — a broken health check must not
 * lock the UI.
 */
export async function initServerMode(signal?: AbortSignal): Promise<void> {
  const res = await getHealth(signal)
  applyMode(res.ok && res.value?.evalOnlyMode === true)
}
