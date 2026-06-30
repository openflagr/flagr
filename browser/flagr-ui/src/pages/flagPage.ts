import type { Router } from 'vue-router'
import * as evalApi from '@/api/eval'
import * as crudApi from '@/api/crud'
import { variantUsedInDistribution, normalizeFlag, normalizeSegment } from '@/helpers/flagModel'
import { evalSummaryFromResult } from '@/helpers/evaluation'
import type { EntityTypeOption } from '@/helpers/flagModel'
import type {
  BatchEvalContext,
  BatchEvalResult,
  Constraint,
  ConstraintFieldKey,
  Distribution,
  DistributionDraft,
  EvalContext,
  EvalResult,
  EvalSummary,
  FlagSnapshot,
  FlagView,
  PutVariantBody,
  Segment,
  SegmentFieldKey,
  Tag,
  Variant,
} from '@/api/types'
import {
  pluckSegmentIds,
  requireConstraintId,
  requireSegmentId,
  requireTagId,
  requireVariantId,
  isIdentifiedSegment,
} from '@/api/types'
import { confirmAndRunApi, type ConfirmVm } from '@/helpers/runApi'
import { runApi } from '@/helpers/runApi'


export interface FlagPageVm extends ConfirmVm {
  $router: Router
  flagId: string
  /** Bumped on each route flagId change; mountFlagPage ignores stale responses. */
  flagPageLoadGen?: number
  flag: FlagView
  newSegment: { description: string; rolloutPercent: number }
  newTag: { value: string }
  tagInputVisible: boolean
  allTags: Tag[]
  entityTypes: EntityTypeOption[]
  allowCreateEntityType: boolean
  dialogCreateSegmentOpen: boolean
  dialogDuplicateFlagVisible: boolean
  dialogEditDistributionOpen: boolean
  selectedSegment: Segment | null
  distributionDraft: Record<string, DistributionDraft>
  loaded: boolean
  historyLoaded: boolean
  historyKey: number
  flagSnapshots: FlagSnapshot[]
  evalContext: EvalContext
  evalResult: EvalResult
  evalSummary: EvalSummary | null
  batchEvalContext: BatchEvalContext
  batchEvalResult: BatchEvalResult
  duplicateInFlight?: boolean
}

export const DEFAULT_SEGMENT = { description: '', rolloutPercent: 50 }
export const DEFAULT_TAG = { value: '' }

export const DUPLICATE_FLAG_CONFIRM_MESSAGE =
  'Duplicate this feature flag? A new flag will be created with the same segments, variants, constraints, distributions, and tags.'

/** Element Plus message duration (ms); user can still dismiss early via showClose. */
export const DUPLICATE_SUCCESS_TOAST_DURATION_MS = 10_000

function showDuplicateSuccessToast(vm: FlagPageVm, cloneId: number): void {
  const clonePath = `#/flags/${cloneId}`
  vm.$message({
    type: 'success',
    duration: DUPLICATE_SUCCESS_TOAST_DURATION_MS,
    showClose: true,
    dangerouslyUseHTMLString: true,
    message:
      `Flag cloned successfully. New flag ID: <strong>${cloneId}</strong>. ` +
      `<a href="${clonePath}" class="duplicate-flag-toast-link" aria-label="Open cloned flag ${cloneId}">Open ${clonePath}</a> to update details.`,
  })
}


export function reloadFlag(vm: FlagPageVm): void {
  runApi(vm, crudApi.getFlag(vm.flagId), {
    onSuccess: (data) => {
      vm.flag = normalizeFlag(data)
      vm.loaded = true
    },
  })
}


export function deleteFlag(vm: FlagPageVm): void {
  const id = vm.flagId
  runApi(vm, crudApi.deleteFlag(id), {
    onSuccess: () => {
      vm.$router.replace({ name: 'home' })
      vm.$message.success(`You deleted flag ${id}`)
    },
  })
}

export function duplicateFlag(vm: FlagPageVm): void {
  if (vm.duplicateInFlight) return
  vm.duplicateInFlight = true
  runApi(vm, crudApi.duplicateFlag(vm.flagId), {
    onFailure: () => {
      vm.duplicateInFlight = false
    },
    onSuccess: (cloned) => {
      vm.duplicateInFlight = false
      const cloneId = cloned.id
      if (cloneId == null) {
        vm.$message.error('Duplicate succeeded but response had no flag id')
        return
      }
      vm.dialogDuplicateFlagVisible = false
      showDuplicateSuccessToast(vm, cloneId)
    },
  })
}

export function putFlag(vm: FlagPageVm): void {
  const f = vm.flag
  runApi(
    vm,
    crudApi.updateFlag(vm.flagId, {
      description: f.description,
      dataRecordsEnabled: f.dataRecordsEnabled,
      key: f.key || '',
      entityType: f.entityType || '',
      notes: f.notes || '',
    }),
    { successMessage: 'Flag updated', onSuccess: () => syncEvalContextFromFlag(vm) },
  )
}

export function handleToggleEnabled(vm: FlagPageVm, checked: boolean): void {
  runApi(vm, crudApi.setFlagEnabled(vm.flagId, checked), {
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
  runApi(vm, crudApi.createTagAndRefreshAllTags(vm.flagId, value), {
    successMessage: 'new tag created',
    onSuccess: ({ tag, allTags }) => {
      vm.newTag = { ...DEFAULT_TAG }
      if (!vm.flag.tags.some((t) => t.value === tag.value)) {
        vm.flag.tags.push(tag)
      }
      vm.tagInputVisible = false
      vm.allTags = allTags
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
  confirmAndRunApi(
    vm,
    `Are you sure you want to delete tag #${tag.value}`,
    crudApi.deleteTagAndReload(vm.flagId, tagId), {
    successMessage: 'tag deleted',
    onSuccess: ({ flag, allTags }) => {
      vm.flag = normalizeFlag(flag)
      vm.allTags = allTags
    },
  })
}

export function handleCreateVariant(vm: FlagPageVm, { key }: { key: string }): void {
  runApi(vm, crudApi.createVariant(vm.flagId, key), {
    successMessage: 'new variant created',
    onSuccess: (variant) => {
      vm.flag.variants = [...vm.flag.variants, variant]
    },
  })
}

export function handleUpdateVariantKey(
  _vm: FlagPageVm,
  { variant, key }: { variant: Variant; key: string },
): void {
  variant.key = key
}

export function handleVariantAttachmentChange(
  _vm: FlagPageVm,
  { variant, valid }: { variant: Variant; valid: boolean },
): void {
  variant.attachmentValid = valid
}

export function putVariant(vm: FlagPageVm, variant: Variant): void {
  if (variant.attachmentValid === false) {
    vm.$message.error('variant attachment is not valid')
    return
  }
  let attachment: PutVariantBody['attachment']
  const raw = variant.attachment
  if (raw === undefined) {
    attachment = undefined
  } else if (typeof raw === 'string') {
    attachment = undefined
  } else {
    attachment = raw
  }
  runApi(
    vm,
    crudApi.updateVariant(vm.flagId, requireVariantId(variant), {
      key: variant.key,
      attachment,
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
  confirmAndRunApi(
    vm,
    `Are you sure you want to delete variant #${variant.id} [${variant.key}]`,
    crudApi.deleteVariant(vm.flagId, variantId),
    {
      successMessage: 'variant deleted',
      onSuccess: () => reloadFlag(vm),
    },
  )
}

export function createSegment(vm: FlagPageVm): void {
  runApi(vm, crudApi.createSegment(vm.flagId, vm.newSegment), {
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
  runApi(
    vm,
    crudApi.updateSegment(vm.flagId, requireSegmentId(segment), {
      description: segment.description,
      rolloutPercent: parseInt(String(segment.rolloutPercent), 10),
    }),
    { successMessage: 'segment updated' },
  )
}

export function deleteSegment(vm: FlagPageVm, segment: Segment): void {
  confirmAndRunApi(
    vm,
    'Are you sure you want to delete this segment?',
    crudApi.deleteSegment(vm.flagId, requireSegmentId(segment)),
    {
      successMessage: 'segment deleted',
      onSuccess: () => reloadFlag(vm),
    },
  )
}

export function handleReorderSegments(vm: FlagPageVm, segments: Segment[]): void {
  runApi(vm, crudApi.reorderSegments(vm.flagId, pluckSegmentIds(segments)), {
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

export function handleUpdateSegmentField(
  _vm: FlagPageVm,
  {
    segment,
    field,
    value,
  }: {
    segment: Segment
    field: SegmentFieldKey
    value: string | number
  },
): void {
  if (field === 'description' && typeof value === 'string') segment.description = value
  else if (field === 'rolloutPercent') segment.rolloutPercent = Number(value)
}

export function createConstraint(
  vm: FlagPageVm,
  { segment, constraint }: { segment: Segment; constraint: Constraint },
): void {
  const c = { ...constraint, property: constraint.property.trim(), value: constraint.value.trim() }
  runApi(vm, crudApi.createConstraint(vm.flagId, requireSegmentId(segment), c), {
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
  runApi(
    vm,
    crudApi.updateConstraint(
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
  confirmAndRunApi(
    vm,
    'Are you sure you want to delete this constraint?',
    crudApi.deleteConstraint(
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

export function handleUpdateConstraintField(
  _vm: FlagPageVm,
  {
    constraint,
    field,
    value,
  }: {
    constraint: Constraint
    field: ConstraintFieldKey
    value: string
  },
): void {
  if (field === 'property') constraint.property = value
  else if (field === 'operator') constraint.operator = value
  else if (field === 'value') constraint.value = value
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
  if (!isIdentifiedSegment(segment)) return
  runApi(vm, crudApi.putSegmentDistributions(vm.flagId, segment.id, distributions), {
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
    loadFlagSnapshots(vm)
  }
}

export function loadFlagSnapshots(vm: FlagPageVm): void {
  runApi(vm, crudApi.listFlagSnapshots(vm.flagId), {
    onSuccess: (data) => {
      vm.flagSnapshots = data
    },
  })
}

export function postEvaluation(vm: FlagPageVm, evalContext: EvalContext): void {
  runApi(vm, evalApi.postEvaluation(evalContext), {
    successMessage: 'evaluation success',
    onSuccess: (response) => {
      vm.evalResult = response
      vm.evalSummary = evalSummaryFromResult(response)
    },
  })
}

export function postEvaluationBatch(vm: FlagPageVm, batchEvalContext: BatchEvalContext): void {
  runApi(vm, evalApi.postEvaluationBatch(batchEvalContext), {
    successMessage: 'evaluation success',
    onSuccess: (response) => {
      vm.batchEvalResult = response
    },
  })
}

export function syncEvalContextFromFlag(vm: FlagPageVm): void {
  const id = vm.flag.id
  if (id == null) return
  vm.evalContext.flagID = id
  vm.evalContext.flagKey = vm.flag.key
  vm.batchEvalContext.flagIDs = [id]
}

/** Loads flag page context; resets route-local state and bumps load generation (route watcher). */
export function mountFlagPage(vm: FlagPageVm): void {
  vm.flagPageLoadGen = (vm.flagPageLoadGen ?? 0) + 1
  vm.loaded = false
  vm.historyLoaded = false
  vm.historyKey++
  vm.flagSnapshots = []
  vm.dialogDuplicateFlagVisible = false
  vm.dialogEditDistributionOpen = false
  vm.dialogCreateSegmentOpen = false
  vm.selectedSegment = null

  const flagId = vm.flagId
  const gen = vm.flagPageLoadGen ?? 0
  runApi(vm, crudApi.loadFlagPageContext(flagId), {
    onSuccess: (load) => {
      if (vm.flagId !== flagId || (vm.flagPageLoadGen ?? 0) !== gen) {
        return
      }
      vm.flag = normalizeFlag(load.flag)
      vm.loaded = true
      vm.allTags = load.allTags
      vm.entityTypes = load.entityTypes
      vm.allowCreateEntityType = load.allowCreateEntityType
      syncEvalContextFromFlag(vm)
    },
  })
}
