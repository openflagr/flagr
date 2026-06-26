import { Effect } from 'effect'
import type { ApiError } from '@/api/errors'
import type { Router } from 'vue-router'
import * as flagsApi from '@/api/flags'
import constants from '@/helpers/constants'
import { entityTypeOptionsFromKeys } from '@/helpers/flagModel'
import { variantUsedInDistribution } from '@/helpers/flagModel'
import { normalizeFlag, normalizeSegment } from '@/helpers/flagModel'
import type {
  Constraint,
  Distribution,
  DistributionDraft,
  FlagView,
  Segment,
  Tag,
  Variant,
} from '@/api/types'
import {
  pluckSegmentIds,
  requireConstraintId,
  requireSegmentId,
  requireTagId,
  requireVariantId,
} from '@/api/types'
import { confirmAndRunApi, type ConfirmVm, type RunApiOptions } from '@/helpers/runApi'
import { runApi } from '@/helpers/runApi'

const { FLAGR_UI_POSSIBLE_ENTITY_TYPES } = constants

export interface FlagPageVm extends ConfirmVm {
  $router: Router
  flagId: string
  flag: FlagView
  newSegment: { description: string; rolloutPercent: number }
  newTag: { value: string }
  tagInputVisible: boolean
  allTags: Tag[]
  entityTypes: { label: string; value: string }[]
  allowCreateEntityType: boolean
  dialogCreateSegmentOpen: boolean
  dialogEditDistributionOpen: boolean
  selectedSegment: Segment | null
  distributionDraft: Record<string, DistributionDraft>
  loaded: boolean
  historyLoaded: boolean
  historyKey: number
}

export const DEFAULT_SEGMENT = { description: '', rolloutPercent: 50 }
export const DEFAULT_TAG = { value: '' }

function runMutation<A>(
  vm: FlagPageVm,
  program: Effect.Effect<A, ApiError>,
  options: RunApiOptions<A>,
): void {
  runApi(vm, program, options)
}

function confirmMutation<A>(
  vm: FlagPageVm,
  message: string,
  program: Effect.Effect<A, ApiError>,
  options: RunApiOptions<A>,
): void {
  confirmAndRunApi(vm, message, program, options)
}

export function reloadFlag(vm: FlagPageVm): void {
  runMutation(vm, flagsApi.getFlag(vm.flagId), {
    onSuccess: (data) => {
      vm.flag = normalizeFlag(data)
      vm.loaded = true
    },
  })
}

function applyEntityTypesToVm(vm: FlagPageVm, entityTypesFromApi: string[]): void {
  if (FLAGR_UI_POSSIBLE_ENTITY_TYPES) {
    vm.entityTypes = entityTypeOptionsFromKeys(
      FLAGR_UI_POSSIBLE_ENTITY_TYPES.split(','),
    )
    vm.allowCreateEntityType = false
    return
  }
  vm.entityTypes = entityTypeOptionsFromKeys(entityTypesFromApi)
  vm.allowCreateEntityType = true
}

export function deleteFlag(vm: FlagPageVm): void {
  const id = vm.flagId
  runMutation(vm, flagsApi.deleteFlag(id), {
    onSuccess: () => {
      vm.$router.replace({ name: 'home' })
      vm.$message.success(`You deleted flag ${id}`)
    },
  })
}

export function putFlag(vm: FlagPageVm): void {
  const f = vm.flag
  runMutation(
    vm,
    flagsApi.updateFlag(vm.flagId, {
      description: f.description,
      dataRecordsEnabled: f.dataRecordsEnabled,
      key: f.key || '',
      entityType: f.entityType || '',
      notes: f.notes || '',
    }),
    { successMessage: 'Flag updated' },
  )
}

export function handleToggleEnabled(vm: FlagPageVm, checked: boolean): void {
  runMutation(vm, flagsApi.setFlagEnabled(vm.flagId, checked), {
    successMessage: `You turned ${checked ? 'on' : 'off'} this feature flag`,
    onSuccess: () => {
      vm.flag.enabled = checked
    },
  })
}

export function handleUpdateFlag(vm: FlagPageVm, patch: Partial<FlagView>): void {
  Object.assign(vm.flag, patch)
}

export function handleCreateTag(vm: FlagPageVm, { value }: { value: string }): void {
  vm.newTag.value = value
  runMutation(vm, flagsApi.createTag(vm.flagId, value), {
    successMessage: 'new tag created',
    onSuccess: (tag) => {
      vm.newTag = { ...DEFAULT_TAG }
      if (!vm.flag.tags!.some((t) => t.value === tag.value)) {
        vm.flag.tags!.push(tag)
      }
      vm.tagInputVisible = false
      runMutation(vm, flagsApi.listAllTags(), {
        onSuccess: (data) => {
          vm.allTags = data
        },
      })
    },
  })
}

export function handleCancelCreateTag(vm: FlagPageVm): void {
  vm.newTag = { ...DEFAULT_TAG }
  vm.tagInputVisible = false
}

export function handleShowTagInput(vm: FlagPageVm): void {
  vm.tagInputVisible = true
}

export function deleteTag(vm: FlagPageVm, tag: Tag): void {
  const tagId = requireTagId(tag)
  const program = Effect.gen(function* () {
    yield* flagsApi.deleteTag(vm.flagId, tagId)
    return yield* flagsApi.loadFlagAndAllTags(vm.flagId)
  })
  confirmMutation(vm, `Are you sure you want to delete tag #${tag.value}`, program, {
    successMessage: 'tag deleted',
    onSuccess: ({ flag, allTags }) => {
      vm.flag = normalizeFlag(flag)
      vm.allTags = allTags
    },
  })
}

export function handleCreateVariant(vm: FlagPageVm, { key }: { key: string }): void {
  runMutation(vm, flagsApi.createVariant(vm.flagId, key), {
    successMessage: 'new variant created',
    onSuccess: (variant) => {
      vm.flag.variants = [...vm.flag.variants, variant]
    },
  })
}

export function handleUpdateVariantKey({
  variant,
  key,
}: {
  variant: Variant
  key: string
}): void {
  variant.key = key
}

export function handleVariantAttachmentChange({
  variant,
  valid,
}: {
  variant: Variant
  valid: boolean
}): void {
  variant.attachmentValid = valid
}

export function putVariant(vm: FlagPageVm, variant: Variant): void {
  if (variant.attachmentValid === false) {
    vm.$message.error('variant attachment is not valid')
    return
  }
  runMutation(
    vm,
    flagsApi.updateVariant(vm.flagId, requireVariantId(variant), {
      key: variant.key,
      attachment: variant.attachment,
    }),
    { successMessage: 'variant updated' },
  )
}

export function deleteVariant(vm: FlagPageVm, variant: Variant): void {
  const variantId = requireVariantId(variant)
  if (variantUsedInDistribution(vm.flag, variantId)) {
    vm.$message.warning(
      'This variant is being used by a segment distribution. Please remove the segment or edit the distribution in order to remove this variant.',
    )
    return
  }
  confirmMutation(
    vm,
    `Are you sure you want to delete variant #${variant.id} [${variant.key}]`,
    flagsApi.deleteVariant(vm.flagId, variantId),
    {
      successMessage: 'variant deleted',
      onSuccess: () => reloadFlag(vm),
    },
  )
}

export function createSegment(vm: FlagPageVm): void {
  runMutation(vm, flagsApi.createSegment(vm.flagId, vm.newSegment), {
    successMessage: 'new segment created',
    onSuccess: (segment) => {
      const normalized = normalizeSegment(segment)
      vm.newSegment = { ...DEFAULT_SEGMENT }
      vm.flag.segments = [...(vm.flag.segments ?? []), normalized]
      vm.dialogCreateSegmentOpen = false
    },
  })
}

export function putSegment(vm: FlagPageVm, segment: Segment): void {
  runMutation(
    vm,
    flagsApi.updateSegment(vm.flagId, requireSegmentId(segment), {
      description: segment.description,
      rolloutPercent: parseInt(String(segment.rolloutPercent), 10),
    }),
    { successMessage: 'segment updated' },
  )
}

export function deleteSegment(vm: FlagPageVm, segment: Segment): void {
  confirmMutation(
    vm,
    'Are you sure you want to delete this segment?',
    flagsApi.deleteSegment(vm.flagId, requireSegmentId(segment)),
    {
      successMessage: 'segment deleted',
      onSuccess: () => reloadFlag(vm),
    },
  )
}

export function handleReorderSegments(vm: FlagPageVm, segments: Segment[]): void {
  runMutation(vm, flagsApi.reorderSegments(vm.flagId, pluckSegmentIds(segments)), {
    successMessage: 'segment reordered',
  })
}

export function moveSegmentUp(vm: FlagPageVm, _element: Segment, index: number): void {
  if (index <= 0) return
  const arr = [...(vm.flag.segments ?? [])]
  ;[arr[index - 1], arr[index]] = [arr[index], arr[index - 1]]
  vm.flag.segments = arr
}

export function moveSegmentDown(vm: FlagPageVm, _element: Segment, index: number): void {
  if (index >= (vm.flag.segments ?? []).length - 1) return
  const arr = [...(vm.flag.segments ?? [])]
  ;[arr[index + 1], arr[index]] = [arr[index], arr[index + 1]]
  vm.flag.segments = arr
}

export function handleUpdateSegmentField({
  segment,
  field,
  value,
}: {
  segment: Segment
  field: keyof Segment
  value: Segment[keyof Segment]
}): void {
  if (field === 'description' && typeof value === 'string') segment.description = value
  else if (field === 'rolloutPercent') segment.rolloutPercent = Number(value)
}

export function createConstraint(
  vm: FlagPageVm,
  { segment, constraint }: { segment: Segment; constraint: Constraint },
): void {
  const c = { ...constraint, property: constraint.property.trim(), value: constraint.value.trim() }
  runMutation(vm, flagsApi.createConstraint(vm.flagId, requireSegmentId(segment), c), {
    successMessage: 'new constraint created',
    onSuccess: (created) => {
      segment.constraints = [...(segment.constraints ?? []), created]
    },
  })
}

export function putConstraint(
  vm: FlagPageVm,
  { segment, constraint }: { segment: Segment; constraint: Constraint },
): void {
  constraint.property = constraint.property.trim()
  constraint.value = constraint.value.trim()
  runMutation(
    vm,
    flagsApi.updateConstraint(
      vm.flagId,
      requireSegmentId(segment),
      requireConstraintId(constraint),
      constraint,
    ),
    { successMessage: 'constraint updated' },
  )
}

export function deleteConstraint(
  vm: FlagPageVm,
  { segment, constraint }: { segment: Segment; constraint: Constraint },
): void {
  confirmMutation(
    vm,
    'Are you sure you want to delete this constraint?',
    flagsApi.deleteConstraint(
      vm.flagId,
      requireSegmentId(segment),
      requireConstraintId(constraint),
    ),
    {
      successMessage: 'constraint deleted',
      onSuccess: () => reloadFlag(vm),
    },
  )
}

export function handleUpdateConstraintField({
  constraint,
  field,
  value,
}: {
  constraint: Constraint
  field: keyof Constraint
  value: Constraint[keyof Constraint]
}): void {
  if (field === 'property' && typeof value === 'string') constraint.property = value
  else if (field === 'operator' && typeof value === 'string') constraint.operator = value
  else if (field === 'value' && typeof value === 'string') constraint.value = value
}

export function handleEditDistribution(vm: FlagPageVm, segment: Segment): void {
  vm.selectedSegment = segment
  const draft: Record<string, DistributionDraft> = {}
  for (const d of segment.distributions ?? []) {
    const { id: _id, ...rest } = d
    draft[d.variantID] = { ...rest }
  }
  vm.distributionDraft = draft
  vm.dialogEditDistributionOpen = true
}

export function handleSaveDistribution(
  vm: FlagPageVm,
  draft: Record<string, DistributionDraft>,
): void {
  const distributions: Distribution[] = Object.values(draft)
    .filter((d) => d.percent !== 0)
    .map((d) => ({ percent: d.percent, variantID: d.variantID, variantKey: d.variantKey }))
  const segment = vm.selectedSegment
  if (!segment?.id) return
  runMutation(vm, flagsApi.putSegmentDistributions(vm.flagId, segment.id, distributions), {
    successMessage: 'distributions updated',
    onSuccess: (data) => {
      segment.distributions = data
      vm.dialogEditDistributionOpen = false
    },
  })
}

export function handleHistoryTabClick(vm: FlagPageVm, tab: { props?: { name?: string } }): void {
  if (tab.props?.name === 'history') {
    vm.historyLoaded = true
    vm.historyKey++
  }
}

export function mountFlagPage(vm: FlagPageVm): void {
  runMutation(vm, flagsApi.loadFlagPageContext(vm.flagId), {
    onSuccess: (load) => {
      vm.flag = normalizeFlag(load.flag)
      vm.loaded = true
      vm.allTags = load.allTags
      applyEntityTypesToVm(vm, load.entityTypesFromApi)
    },
  })
}

/** Handlers that take `(vm, …)` — bind once in `Flag.vue` via `bindPageHandlers`. */
export const flagPageVmHandlers = {
  deleteFlag,
  putFlag,
  handleToggleEnabled,
  handleUpdateFlag,
  handleCreateTag,
  handleCancelCreateTag,
  handleShowTagInput,
  deleteTag,
  handleCreateVariant,
  putVariant,
  deleteVariant,
  createSegment,
  putSegment,
  deleteSegment,
  handleReorderSegments,
  moveSegmentUp,
  moveSegmentDown,
  createConstraint,
  putConstraint,
  deleteConstraint,
  handleEditDistribution,
  handleSaveDistribution,
  handleHistoryTabClick,
} as const