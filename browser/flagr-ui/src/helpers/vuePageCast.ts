import type { FlagPageVm } from '@/pages/flagPage'
import type { FlagsListVm } from '@/pages/flagsListPage'

/** Options API `this` matches page `data` + router/message chrome; one cast at the template edge. */
export function castFlagPage(host: unknown): FlagPageVm {
  return host as FlagPageVm
}

export function castFlagsList(host: unknown): FlagsListVm {
  return host as FlagsListVm
}