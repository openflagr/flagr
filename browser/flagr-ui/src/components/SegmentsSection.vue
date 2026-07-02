<template>
  <el-card class="segments-container is-card-secondary">
    <template #header>
      <div class="el-card-header">
        <div class="flex-row">
          <div class="flex-row-left">
            <h2>Segments</h2>
          </div>
          <div class="flex-row-right">
            <el-tooltip
              :content="reorderDirty ? 'Unsaved reorder — click to persist' : 'Use buttons to reorder, then click to persist'"
              placement="top"
              effect="light"
            >
              <el-button
                size="small"
                :type="reorderDirty ? 'warning' : undefined"
                @click="handleReorder"
              >
                Reorder{{ reorderDirty ? ' *' : '' }}
              </el-button>
            </el-tooltip>
            <el-button
              size="small"
              data-testid="open-new-segment-btn"
              @click="$emit('new-segment')"
            >
              New Segment
            </el-button>
          </div>
        </div>
      </div>
    </template>

    <div v-if="segments.length">
      <div
        v-for="(element, index) in segments"
        :key="element.id"
        class="segment-card"
      >
        <!-- Header row -->
        <div class="seg-header">
          <span class="seg-id ui-id-badge">#{{ element.id }}</span>
          <div class="seg-header-fields">
            <el-input
              size="small"
              placeholder="Description"
              :model-value="element.description"
              data-testid="segment-desc-input"
              @update:model-value="onSegmentFieldChange(element, 'description', $event)"
            />
            <el-input
              class="segment-rollout-percent"
              size="small"
              placeholder="0"
              :model-value="element.rolloutPercent"
              data-testid="segment-rollout-input"
              :min="0"
              :max="100"
              @update:model-value="onSegmentFieldChange(element, 'rolloutPercent', $event)"
            >
              <template #prepend>
                rollout
              </template>
              <template #append>
                %
              </template>
            </el-input>
          </div>
          <div class="seg-header-actions">
            <el-button-group class="seg-reorder-group">
              <el-tooltip
                content="Move up"
                placement="top"
              >
                <el-button
                  size="small"
                  :disabled="index===0"
                  data-testid="move-segment-up-btn"
                  @click="handleMoveUp(element,index)"
                >
                  <el-icon><ArrowUp /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip
                content="Move down"
                placement="top"
              >
                <el-button
                  size="small"
                  :disabled="index===segments.length-1"
                  data-testid="move-segment-down-btn"
                  @click="handleMoveDown(element,index)"
                >
                  <el-icon><ArrowDown /></el-icon>
                </el-button>
              </el-tooltip>
            </el-button-group>
            <el-tooltip
              :content="SAVE_DIRTY_TOOLTIP"
              placement="top"
              effect="light"
              :disabled="!isSegmentDirty(element)"
            >
              <el-button
                size="small"
                :plain="!isSegmentDirty(element)"
                :type="saveButtonType(isSegmentDirty(element))"
                data-testid="save-segment-btn"
                @click="handleSaveSegment(element)"
              >
                {{ saveButtonLabel(isSegmentDirty(element)) }}
              </el-button>
            </el-tooltip>
            <el-button
              size="small"
              data-testid="delete-segment-btn"
              @click="$emit('delete-segment', element)"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>

        <!-- Constraints + Distribution -->
        <div class="seg-panel-row">
          <div class="seg-panel ui-surface-inset">
            <div class="seg-section-title ui-section-title">
              Constraints <span class="seg-section-subtitle">— match ALL</span>
            </div>
            <div
              v-if="!(element.constraints ?? []).length"
              class="card--empty"
            >
              No constraints — all entities pass
            </div>
            <div
              v-if="element.id != null"
              class="constraint-table"
            >
              <div
                v-for="(constraint, cIdx) in (element.constraints ?? [])"
                :key="constraint.id"
                class="constraint-row"
              >
                <span class="constraint-logic">{{ cIdx === 0 ? 'IF' : 'AND' }}</span>
                <el-input
                  size="small"
                  class="constraint-cell constraint-control"
                  :placeholder="propertyPlaceholderFor(displayUiOperator(constraint))"
                  :model-value="constraint.property"
                  data-testid="constraint-prop-input"
                  @update:model-value="onConstraintFieldChange(element, constraint, 'property', $event)"
                />
                <el-select
                  class="constraint-cell constraint-control width--full constraint-op-select"
                  size="small"
                  :model-value="displayUiOperator(constraint)"
                  placeholder="Operator"
                  popper-class="constraint-op-select-popper"
                  data-testid="constraint-op-select"
                  @update:model-value="onConstraintOperatorChange(element, constraint, $event)"
                >
                  <template #label>
                    {{ operatorSelectClosedBadge(displayUiOperator(constraint)) }}
                  </template>
                  <el-option-group
                    v-for="group in groupedOperatorOptions"
                    :key="group.label"
                    :label="group.label"
                  >
                    <el-option
                      v-for="item in group.options"
                      :key="item.value"
                      :label="operatorSelectLabel(item)"
                      :value="item.value"
                    >
                      <ConstraintOperatorOption :item="item" />
                    </el-option>
                  </el-option-group>
                </el-select>
                <el-input
                  size="small"
                  class="constraint-cell constraint-control"
                  :placeholder="valuePlaceholderFor(displayUiOperator(constraint))"
                  :model-value="constraintValueForInput(constraint)"
                  data-testid="constraint-value-input"
                  @update:model-value="onConstraintFieldChange(element, constraint, 'value', $event)"
                />
                <div class="constraint-actions">
                  <el-tooltip
                    :content="SAVE_DIRTY_TOOLTIP"
                    placement="top"
                    effect="light"
                    :disabled="!isConstraintDirty(element, constraint)"
                  >
                    <el-button
                      size="small"
                      :plain="!isConstraintDirty(element, constraint)"
                      :type="saveButtonType(isConstraintDirty(element, constraint))"
                      data-testid="save-constraint-btn"
                      @click="handleSaveConstraint(element, constraint)"
                    >
                      {{ saveButtonLabel(isConstraintDirty(element, constraint)) }}
                    </el-button>
                  </el-tooltip>
                  <el-button
                    size="small"
                    plain
                    data-testid="delete-constraint-btn"
                    @click="() => $emit('delete-constraint', { segment: element, constraint })"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <ConstraintOperatorHint
                  class="constraint-hint-cell"
                  :operator="displayUiOperator(constraint)"
                  :operator-options="operatorOptions"
                  test-id="constraint-operator-hint"
                />
              </div>
              <div
                v-if="(element.constraints ?? []).length"
                class="constraint-add-divider"
                aria-hidden="true"
              />
              <p class="constraint-add-caption">
                {{ (element.constraints ?? []).length ? 'Add another constraint' : 'Add a constraint' }}
              </p>
              <div class="constraint-row constraint-row--add">
                <span
                  class="constraint-logic constraint-logic--add"
                  aria-hidden="true"
                >+</span>
                <el-input
                  v-model="newConstraints[element.id].property"
                  size="small"
                  class="constraint-cell constraint-control"
                  :placeholder="propertyPlaceholderFor(newConstraints[element.id].operator)"
                  data-testid="new-constraint-prop-input"
                />
                <el-select
                  v-model="newConstraints[element.id].operator"
                  size="small"
                  class="constraint-cell constraint-control width--full constraint-op-select"
                  placeholder="Operator"
                  popper-class="constraint-op-select-popper"
                  data-testid="new-constraint-op-select"
                >
                  <template #label>
                    {{ operatorSelectClosedBadge(newConstraints[element.id].operator) }}
                  </template>
                  <el-option-group
                    v-for="group in groupedOperatorOptions"
                    :key="group.label"
                    :label="group.label"
                  >
                    <el-option
                      v-for="item in group.options"
                      :key="item.value"
                      :label="operatorSelectLabel(item)"
                      :value="item.value"
                    >
                      <ConstraintOperatorOption :item="item" />
                    </el-option>
                  </el-option-group>
                </el-select>
                <el-input
                  v-model="newConstraints[element.id].value"
                  size="small"
                  class="constraint-cell constraint-control"
                  :placeholder="valuePlaceholderFor(newConstraints[element.id].operator)"
                  data-testid="new-constraint-value-input"
                />
                <el-button
                  size="small"
                  type="primary"
                  plain
                  class="constraint-add-btn"
                  :disabled="!newConstraints[element.id]?.operator || !newConstraints[element.id]?.property || !newConstraints[element.id]?.value"
                  data-testid="add-constraint-btn"
                  @click.prevent="handleCreateConstraint(element)"
                >
                  Add constraint
                </el-button>
                <ConstraintOperatorHint
                  class="constraint-hint-cell"
                  :operator="newConstraints[element.id].operator"
                  :operator-options="operatorOptions"
                  test-id="new-constraint-operator-hint"
                />
              </div>
            </div>
          </div>
          <div class="seg-panel seg-panel-dist ui-surface-inset">
            <div class="seg-section-title seg-section-title--with-action ui-section-title">
              <span>Distribution</span>
              <el-button
                size="small"
                link
                type="primary"
                data-testid="edit-distribution-btn"
                @click="$emit('edit-distribution', element)"
              >
                <el-icon><Edit /></el-icon> Edit
              </el-button>
            </div>
            <div
              v-if="(element.distributions ?? []).length"
              class="dist-list"
            >
              <div
                v-for="distribution in (element.distributions ?? [])"
                :key="distribution.id"
                class="dist-item"
              >
                <div class="dist-header">
                  <span class="dist-variant">{{ distribution.variantKey }}</span>
                  <span class="dist-pct">{{ distribution.percent }}%</span>
                </div>
                <el-progress
                  :percentage="distribution.percent"
                  color="var(--el-color-primary)"
                  :show-text="false"
                  :stroke-width="6"
                />
              </div>
            </div>
            <div
              v-else
              class="card--empty"
            >
              No distribution
            </div>
          </div>
        </div>
      </div>
    </div>
    <div
      v-else
      class="card--cue"
    >
      <p class="card--cue-title">
        No segments yet
      </p>
      <p class="card--cue-body">
        Segments are the targeting rules that decide which entities match. Each segment has constraints (e.g. <code>country EQ "US"</code>) and a distribution over variants. Add one to start targeting.
      </p>
    </div>
  </el-card>
</template>

<script lang="ts">
import {
  SAVE_DIRTY_TOOLTIP,
  saveButtonLabel as fmtSaveLabel,
  saveButtonType as fmtSaveType,
} from '@/helpers/saveDirtyUi'

import ConstraintOperatorHint from '@/components/ConstraintOperatorHint.vue'
import ConstraintOperatorOption from '@/components/ConstraintOperatorOption.vue'
import { Delete, Edit, ArrowUp, ArrowDown } from '@element-plus/icons-vue'
import type { PropType } from 'vue'
import type { Constraint, ConstraintFieldKey, Segment, SegmentFieldKey } from '@/api/types'
import {
  findOperatorUi,
  operatorOptionGroups,
  type OperatorUiOption,
} from '@/helpers/constraintOperators'
import {
  operatorSelectClosedBadge as formatOperatorSelectClosedBadge,
  operatorSelectLabel as formatOperatorSelectLabel,
} from '@/helpers/constraintOperatorUi'
import {
  applyUiOperatorSelection,
  constraintValueForInput as sugarConstraintValueForInput,
  resolveUiOperator,
} from '@/helpers/constraintOperatorSugar'

interface NewConstraintDraft {
  operator: string
  property: string
  value: string
}

function emptyNewConstraintDraft(): NewConstraintDraft {
  return { operator: '', property: '', value: '' }
}

export default {
  name: 'SegmentsSection',
  components: { ConstraintOperatorHint, ConstraintOperatorOption, Delete, Edit, ArrowUp, ArrowDown },
  props: {
    segments: { type: Array as PropType<Segment[]>, required: true },
    operatorOptions: { type: Array as PropType<OperatorUiOption[]>, required: true },
  },
  emits: [
    'move-up',
    'move-down',
    'reorder',
    'new-segment',
    'save-segment',
    'delete-segment',
    'update-segment-field',
    'update-constraint-field',
    'save-constraint',
    'delete-constraint',
    'create-constraint',
    'edit-distribution',
  ],
  data() {
    return {
      SAVE_DIRTY_TOOLTIP,
      reorderDirty: false,
      segmentDirtyIds: {} as Record<number, boolean>,
      constraintDirtyKeys: {} as Record<string, boolean>,
      newConstraints: {} as Record<number, NewConstraintDraft>,
    }
  },
  computed: {
    groupedOperatorOptions() {
      return operatorOptionGroups(this.operatorOptions)
    },
  },
  watch: {
    segments: {
      immediate: true,
      handler(segs: Segment[]) {
        for (const seg of segs) {
          if (seg.id != null && !(seg.id in this.newConstraints)) {
            this.newConstraints[seg.id] = emptyNewConstraintDraft()
          }
        }
      },
    },
  },
  methods: {
    displayUiOperator(constraint: Constraint): string {
      return resolveUiOperator(constraint.operator, constraint.value)
    },
    constraintValueForInput(constraint: Constraint): string {
      return sugarConstraintValueForInput(constraint)
    },
    onConstraintOperatorChange(
      segment: Segment,
      constraint: Constraint,
      uiOperator: string,
    ): void {
      applyUiOperatorSelection(constraint, uiOperator)
      this.markConstraintDirty(segment, constraint)
      this.$emit('update-constraint-field', {
        segment,
        constraint,
        field: 'operator',
        value: constraint.operator,
      })
    },
    operatorSelectLabel(op: OperatorUiOption): string {
      return formatOperatorSelectLabel(op)
    },
    operatorSelectClosedBadge(operatorValue: string | undefined): string {
      return formatOperatorSelectClosedBadge(operatorValue)
    },
    operatorDescription(operator: string): string {
      return findOperatorUi(operator, this.operatorOptions)?.description ?? ''
    },
    propertyPlaceholderFor(operator: string): string {
      const op = findOperatorUi(operator, this.operatorOptions)
      return op?.propertyPlaceholder ?? 'Property'
    },
    valuePlaceholderFor(operator: string): string {
      const op = findOperatorUi(operator, this.operatorOptions)
      return op?.valuePlaceholder ?? 'Value'
    },
    saveButtonLabel(dirty: boolean) {
      return fmtSaveLabel(dirty)
    },
    saveButtonType(dirty: boolean) {
      return fmtSaveType(dirty)
    },
    constraintDirtyKey(segment: Segment, constraint: { id?: number }): string {
      return `${segment.id ?? 'x'}:${constraint.id ?? 'x'}`
    },
    isSegmentDirty(segment: Segment): boolean {
      return segment.id != null && !!this.segmentDirtyIds[segment.id]
    },
    isConstraintDirty(segment: Segment, constraint: { id?: number }): boolean {
      return !!this.constraintDirtyKeys[this.constraintDirtyKey(segment, constraint)]
    },
    markSegmentDirty(segment: Segment): void {
      if (segment.id != null) this.segmentDirtyIds[segment.id] = true
    },
    markConstraintDirty(segment: Segment, constraint: { id?: number }): void {
      this.constraintDirtyKeys[this.constraintDirtyKey(segment, constraint)] = true
    },
    clearSegmentDirty(segment: Segment): void {
      if (segment.id != null) delete this.segmentDirtyIds[segment.id]
    },
    clearConstraintDirty(segment: Segment, constraint: { id?: number }): void {
      delete this.constraintDirtyKeys[this.constraintDirtyKey(segment, constraint)]
    },
    handleSaveSegment(segment: Segment): void {
      this.$emit('save-segment', segment)
      this.clearSegmentDirty(segment)
    },
    handleSaveConstraint(segment: Segment, constraint: Constraint): void {
      this.$emit('save-constraint', { segment, constraint })
      this.clearConstraintDirty(segment, constraint)
    },

    handleMoveUp(element: Segment, index: number) {
      this.reorderDirty = true
      this.$emit('move-up', element, index)
    },
    handleMoveDown(element: Segment, index: number) {
      this.reorderDirty = true
      this.$emit('move-down', element, index)
    },
    handleReorder() {
      this.reorderDirty = false
      this.$emit('reorder', this.segments)
    },
    onSegmentFieldChange(
      segment: Segment,
      field: SegmentFieldKey,
      value: string | number,
    ) {
      this.markSegmentDirty(segment)
      this.$emit('update-segment-field', { segment, field, value })
    },
    onConstraintFieldChange(
      segment: Segment,
      constraint: { id?: number },
      field: ConstraintFieldKey,
      value: string,
    ) {
      this.markConstraintDirty(segment, constraint)
      this.$emit('update-constraint-field', { segment, constraint, field, value })
    },
    handleCreateConstraint(element: Segment) {
      const id = element.id!
      const c = this.newConstraints[id]
      if (!c.operator) return
      this.$emit('create-constraint', {
        segment: element,
        constraint: { operator: c.operator, property: c.property, value: c.value },
      })
      this.newConstraints[id] = emptyNewConstraintDraft()
    },
  },
}
</script>

<style lang="scss" scoped>
.segment-card {
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  border-radius: var(--radius-lg);
  padding: var(--space-2xs) var(--space-xs);
  margin-bottom: var(--space-2xs);
  box-shadow: 0 1px 4px rgba(0,0,0,0.04);
}
.seg-header {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
  margin-bottom: var(--space-2xs);
  padding-bottom: var(--space-2xs);
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.seg-header-fields {
  display: flex;
  gap: var(--space-2xs);
  flex: 1;
  > * { flex: 1; }
}
.seg-header-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
  flex-shrink: 0;
}
.seg-reorder-group {
  // Connected up/down pair reads as a single "reorder" control.
  :deep(.el-button) { padding-left: var(--space-2xs); padding-right: var(--space-2xs); }
}
.seg-panel-row {
  display: flex;
  gap: var(--space-sm);
  align-items: flex-start;
}
.seg-panel {
  flex: 1;
  min-width: 0;
}
.seg-panel-dist {
  flex: 0 0 220px;
}
.seg-section-title {
  margin-bottom: var(--space-2xs);
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
}
.seg-section-title--with-action {
  justify-content: space-between;
  :deep(.el-button) {
    text-transform: none;
    letter-spacing: normal;
  }
}
.seg-section-subtitle {
  font-weight: 400;
  color: var(--el-text-color-placeholder);
  text-transform: none;
  letter-spacing: 0;
}

// --- Constraints ---
.constraint-table {
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr) minmax(0, 1.28fr) minmax(0, 1fr) auto;
  column-gap: var(--space-2xs);
  row-gap: var(--space-3xs);
  align-items: center;
  margin-top: var(--space-3xs);
}
.constraint-row {
  display: contents;
}
.constraint-logic {
  font-size: var(--font-size-micro);
  font-weight: var(--font-weight-bold);
  color: var(--el-text-color-secondary);
  letter-spacing: var(--letter-spacing-wide);
  text-transform: uppercase;
  text-align: right;
  justify-self: end;
  align-self: center;
  white-space: nowrap;
}
.constraint-logic--add {
  color: var(--el-color-primary);
}
.constraint-cell {
  min-width: 0;
  align-self: center;
}
.constraint-control {
  width: 100%;
  :deep(.el-input__wrapper) {
    border-radius: var(--radius-sm);
  }
}
.constraint-actions {
  display: flex;
  gap: var(--space-3xs);
  align-self: center;
  justify-self: end;
}
.constraint-hint-cell {
  grid-column: 2 / -1;
  align-self: start;
}

.constraint-add-caption {
  grid-column: 1 / -1;
  margin: 0;
  font-size: var(--font-size-caption);
  font-weight: var(--font-weight-semibold);
  letter-spacing: var(--letter-spacing-ui);
  color: var(--el-text-color-secondary);
}
.constraint-row--add {
  opacity: 0.88;

  .constraint-logic--add {
    color: var(--el-text-color-placeholder);
    font-size: var(--font-size-tab);
    font-weight: 600;
    line-height: 1;
  }

  :deep(.el-input__wrapper),
  :deep(.el-select__wrapper) {
    background-color: var(--el-fill-color-lighter);
  }
}

.constraint-add-divider {
  grid-column: 1 / -1;
  height: 0;
  border-top: 1px dashed var(--el-border-color-lighter);
  margin: var(--space-3xs) 0 var(--space-3xs);
}
.constraint-add-btn {
  justify-self: end;
  align-self: center;
}

:deep(.constraint-hint-line) {
  margin: 0 0 var(--space-3xs);
}

@media (max-width: 720px) {
  .constraint-table {
    grid-template-columns: 36px 1fr;
  }
  .constraint-cell--prop,
  .constraint-cell:nth-child(2) {
    grid-column: 2;
  }
  .constraint-actions {
    grid-column: 2;
    justify-self: start;
  }
  .constraint-hint-cell {
    grid-column: 2;
  }
  .constraint-add-btn {
    grid-column: 2;
    justify-self: start;
  }
}

// --- Distribution ---
.dist-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2xs);
}
.dist-item {
  padding: 0;
}
.dist-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2px;
}

.dist-variant {
  font-size: var(--font-size-body-sm);
  font-weight: var(--font-weight-medium);
  color: var(--el-text-color-regular);
}
.dist-pct {
  font-size: var(--font-size-body-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--el-text-color-primary);
  font-variant-numeric: tabular-nums;
}

@media (max-width: 768px) {
  .seg-header {
    flex-wrap: wrap;
    gap: var(--space-2xs);
  }
  .seg-header-fields {
    flex: 1 1 100%;
    order: 2;
  }
  .seg-header-actions {
    order: 3;
    margin-left: auto;
  }
  .seg-panel-row {
    flex-direction: column;
    gap: var(--space-2xs);
  }
  .seg-panel-dist {
    flex: 0 0 auto;
    width: 100%;
  }
}
@media (max-width: 480px) {
  .seg-header-fields {
    flex-direction: column;
  }
}
</style>

