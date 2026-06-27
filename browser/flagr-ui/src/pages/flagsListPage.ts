import * as crudApi from '@/api/crud'
import type { Router } from 'vue-router'
import type { CreateFlagPayload, Flag } from '@/api/types'
import { requireFlagId } from '@/api/types'
import { confirmAndRunApi, type ConfirmVm } from '@/helpers/runApi'
import { runApi } from '@/helpers/runApi'

export interface FlagsCache {
  flags: Flag[]
  maxSnapshotID: number
}

let flagsCache: FlagsCache | null = null

export function getFlagsCache(): FlagsCache | null {
  return flagsCache
}

export function setFlagsCache(cache: FlagsCache): void {
  flagsCache = cache
}

export interface FlagsListVm extends ConfirmVm {
  $router: Router
  flags: Flag[]
  loaded: boolean
  deletedFlagsLoaded: boolean
  deletedFlags: Flag[]
  showCreateModal: boolean
  newFlag: { description: string }
}

export function refreshFlags(vm: FlagsListVm): void {
  const cachedId = getFlagsCache()?.maxSnapshotID
  runApi(vm, crudApi.listFlagsIfStale(cachedId), {
    onSuccess: (result) => {
      if (!result) return
      setFlagsCache(result)
      vm.flags = result.flags
      vm.loaded = true
    },
  })
}

export function mountFlagsList(vm: FlagsListVm): void {
  refreshFlags(vm)
}

export function createFlag(vm: FlagsListVm, params?: Partial<CreateFlagPayload>): void {
  if (!vm.newFlag.description) return
  const payload: CreateFlagPayload = params
    ? { ...vm.newFlag, ...params }
    : { ...vm.newFlag }
  runApi(vm, crudApi.createFlag(payload), {
    successMessage: 'Flag created',
    onSuccess: (flag) => {
      vm.newFlag.description = ''
      vm.showCreateModal = false
      vm.flags.unshift(flag)
    },
  })
}

export function createBooleanFlag(vm: FlagsListVm): void {
  createFlag(vm, { template: 'simple_boolean_flag' })
}

export function restoreFlag(vm: FlagsListVm, row: Flag): void {
  confirmAndRunApi(
    vm,
    'This will recover the deleted flag. Continue?',
    crudApi.restoreFlag(requireFlagId(row)),
    {
      successMessage: 'Flag restored',
      onSuccess: (flag) => {
        vm.flags.push(flag)
        vm.deletedFlags = vm.deletedFlags.filter((el) => el.id !== flag.id)
      },
    },
  )
}

export function fetchDeletedFlags(vm: FlagsListVm): void {
  if (vm.deletedFlagsLoaded) return
  runApi(vm, crudApi.listDeletedFlags(), {
    onSuccess: (data) => {
      vm.deletedFlags = [...data].reverse()
      vm.deletedFlagsLoaded = true
    },
  })
}

export function datetimeFormatter(_row: Flag, _col: unknown, val: string): string {
  return val ? val.split('.')[0] : ''
}

export function filterStatus(value: boolean, row: Flag): boolean {
  return row.enabled === value
}

export function goToFlag(vm: FlagsListVm, row: Flag): void {
  vm.$router.push({ name: 'flag', params: { flagId: String(row.id) } })
}