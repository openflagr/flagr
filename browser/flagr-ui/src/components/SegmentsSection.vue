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
              <ConstraintExistingRow
                v-for="(constraint, cIdx) in (element.constraints ?? [])"
                :key="constraint.id"
                :constraint="constraint"
                :index="cIdx"
                :operator-options="operatorOptions"
                :grouped-operator-options="groupedOperatorOptions"
                :dirty="isConstraintDirty(element, constraint)"
                :save-button-label="saveButtonLabel(isConstraintDirty(element, constraint))"
                :save-button-type="saveButtonType(isConstraintDirty(element, constraint)) ?? ''"
                @update-field="onConstraintFieldChange(element, constraint, $event.field, $event.value)"
                @update-operator="onConstraintOperatorFieldChange(element, constraint, $event.uiOperator)"
                @save="handleSaveConstraint(element, constraint)"
                @delete="$emit('delete-constraint', { segment: element, constraint })"
              />
              <ConstraintAddRow
                :draft="newConstraints[element.id]"
                :operator-options="operatorOptions"
                :grouped-operator-options="groupedOperatorOptions"
                :show-divider="(element.constraints ?? []).length > 0"
                show-caption
                :caption="(element.constraints ?? []).length ? 'Add another constraint' : 'Add a constraint'"
                @update:draft="newConstraints[element.id] = $event"
                @add="handleCreateConstraint(element)"
              />
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

import ConstraintAddRow, { type NewConstraintDraft } from '@/components/ConstraintAddRow.vue'
import ConstraintExistingRow from '@/components/ConstraintExistingRow.vue'
import { Delete, Edit, ArrowUp, ArrowDown } from '@element-plus/icons-vue'
import type { PropType } from 'vue'
import type { Constraint, ConstraintFieldKey, Segment, SegmentFieldKey } from '@/api/types'
import { applyUiOperatorSelection } from '@/helpers/constraintOperatorSugar'
import { operatorOptionGroups, type OperatorUiOption } from '@/helpers/constraintOperators'

function emptyNewConstraintDraft(): NewConstraintDraft {
  return { operator: '', property: '', value: '' }
}


export default {
  name: 'SegmentsSection',
  components: {
    ConstraintAddRow,
    ConstraintExistingRow,
    Delete,
    Edit,
    ArrowUp,
    ArrowDown,
  },
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
        const ids = new Set<number>()
        for (const seg of segs) {
          if (seg.id == null) continue
          ids.add(seg.id)
          if (!(seg.id in this.newConstraints)) {
            this.newConstraints[seg.id] = emptyNewConstraintDraft()
          }
        }
        for (const key of Object.keys(this.newConstraints)) {
          if (!ids.has(Number(key))) delete this.newConstraints[Number(key)]
        }
      },
    },
  },
  methods: {
    onConstraintOperatorFieldChange(
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
  .seg-panel {
    width: 100%;
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

