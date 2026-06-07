<template>
  <el-card class="segments-container">
    <template #header>
      <div class="el-card-header">
        <div class="flex-row">
          <div class="flex-row-left"><h2>Segments</h2></div>
          <div class="flex-row-right">
            <el-tooltip :content="reorderDirty ? 'Unsaved reorder — click to persist' : 'Use buttons to reorder, then click to persist'" placement="top" effect="light">
              <el-button size="small" :type="reorderDirty ? 'warning' : undefined" @click="handleReorder">Reorder{{ reorderDirty ? ' *' : '' }}</el-button>
            </el-tooltip>
            <el-button size="small" @click="$emit('new-segment')" data-testid="open-new-segment-btn">New Segment</el-button>
          </div>
        </div>
      </div>
    </template>

    <div v-if="segments.length">
      <div v-for="(element, index) in segments" :key="element.id" class="segment-card">
        <!-- Header row -->
        <div class="seg-header">
          <span class="seg-id">#{{ element.id }}</span>
          <div class="seg-header-fields">
            <el-input size="small" placeholder="Description" :model-value="element.description"
              @update:model-value="onSegmentFieldChange(element, 'description', $event)"
              data-testid="segment-desc-input" />
            <el-input class="segment-rollout-percent" size="small" placeholder="0" :model-value="element.rolloutPercent"
              @update:model-value="onSegmentFieldChange(element, 'rolloutPercent', $event)"
              data-testid="segment-rollout-input" :min="0" :max="100">
              <template #prepend>rollout</template>
              <template #append>%</template>
            </el-input>
          </div>
          <div class="seg-header-actions">
            <el-tooltip content="Move up" placement="top"><el-button size="small" :disabled="index===0" @click="handleMoveUp(element,index)" data-testid="move-segment-up-btn"><el-icon><ArrowUp /></el-icon></el-button></el-tooltip>
            <el-tooltip content="Move down" placement="top"><el-button size="small" :disabled="index===segments.length-1" @click="handleMoveDown(element,index)" data-testid="move-segment-down-btn"><el-icon><ArrowDown /></el-icon></el-button></el-tooltip>
            <el-button size="small" plain @click="$emit('save-segment', element)" data-testid="save-segment-btn">Save</el-button>
            <el-button size="small" @click="$emit('delete-segment', element)" data-testid="delete-segment-btn"><el-icon><Delete /></el-icon></el-button>
          </div>
        </div>

        <!-- Constraints + Distribution -->
        <div class="seg-panel-row">
          <div class="seg-panel">
            <div class="seg-section-title">Constraints <span class="seg-section-subtitle">— match ALL</span></div>
            <div v-if="element.constraints.length" class="constraint-grid">
              <div v-for="(constraint, cIdx) in element.constraints" :key="constraint.id" class="constraint-row">
                <span class="constraint-logic">{{ cIdx === 0 ? 'IF' : 'AND' }}</span>
                <div class="constraint-input-group">
                  <el-input size="small" placeholder="Property" :model-value="constraint.property"
                    @update:model-value="onConstraintFieldChange(element, constraint, 'property', $event)"
                    data-testid="constraint-prop-input" />
                  <el-select class="width--full" size="small" :model-value="constraint.operator"
                    @update:model-value="onConstraintFieldChange(element, constraint, 'operator', $event)"
                    placeholder="OP" data-testid="constraint-op-select">
                    <el-option v-for="item in operatorOptions" :key="item.value" :label="item.label" :value="item.value" />
                  </el-select>
                  <el-input size="small" :model-value="constraint.value"
                    @update:model-value="onConstraintFieldChange(element, constraint, 'value', $event)"
                    data-testid="constraint-value-input" />
                </div>
                <div class="constraint-actions">
                  <el-button size="small" plain @click="$emit('save-constraint', { segment: element, constraint })" data-testid="save-constraint-btn">Save</el-button>
                  <el-button size="small" plain @click="() => $emit('delete-constraint', { segment: element, constraint })" data-testid="delete-constraint-btn"><el-icon><Delete /></el-icon></el-button>
                </div>
              </div>
            </div>
            <div class="card--empty" v-if="!element.constraints.length">No constraints — all entities pass</div>
            <div class="constraint-row new-constraint-row">
              <span class="constraint-logic">AND</span>
              <div class="constraint-input-group">
                <el-input size="small" placeholder="Property" v-model="newConstraints[element.id].property" data-testid="new-constraint-prop-input" />
                <el-select size="small" v-model="newConstraints[element.id].operator" placeholder="OP" data-testid="new-constraint-op-select">
                  <el-option v-for="item in operatorOptions" :key="item.value" :label="item.label" :value="item.value" />
                </el-select>
                <el-input size="small" v-model="newConstraints[element.id].value" placeholder="Value" data-testid="new-constraint-value-input" />
              </div>
              <el-button size="small" type="primary" plain :disabled="!newConstraints[element.id]?.property || !newConstraints[element.id]?.value"
                @click.prevent="handleCreateConstraint(element)" data-testid="add-constraint-btn">Add</el-button>
            </div>
          </div>
          <div class="seg-panel seg-panel-dist">
            <div class="seg-section-title">
              <span>Distribution</span>
              <el-button size="small" link type="primary" @click="$emit('edit-distribution', element)"><el-icon><Edit /></el-icon> Edit</el-button>
            </div>
            <div v-if="element.distributions.length" class="dist-list">
              <div v-for="distribution in element.distributions" :key="distribution.id" class="dist-item">
                <div class="dist-header">
                  <span class="dist-variant">{{ distribution.variantKey }}</span>
                  <span class="dist-pct">{{ distribution.percent }}%</span>
                </div>
                <el-progress :percentage="distribution.percent" color="var(--el-color-primary)" :show-text="false" :stroke-width="6" />
              </div>
            </div>
            <div class="card--empty" v-else>No distribution</div>
          </div>
        </div>
      </div>
    </div>
    <div class="card--error" v-else>No segments created for this feature flag yet</div>
  </el-card>
</template>

<script>
import { Delete, Edit, ArrowUp, ArrowDown } from "@element-plus/icons-vue"

export default {
  name: "segments-section",
  components: { Delete, Edit, ArrowUp, ArrowDown },
  props: { segments: { type: Array, required: true }, operatorOptions: { type: Array, required: true } },
  emits: ["move-up","move-down","reorder","new-segment","save-segment","delete-segment","update-segment-field","update-constraint-field","save-constraint","delete-constraint","create-constraint","edit-distribution"],
  data() { return { reorderDirty: false, newConstraints: {} } },
  watch: {
    segments: {
      immediate: true,
      handler(segs) {
        for (const seg of segs) {
          if (!(seg.id in this.newConstraints)) {
            this.newConstraints[seg.id] = { operator: 'EQ', property: '', value: '' }
          }
        }
      }
    }
  },
  methods: {
    handleMoveUp(element, index) { this.reorderDirty = true; this.$emit('move-up', element, index) },
    handleMoveDown(element, index) { this.reorderDirty = true; this.$emit('move-down', element, index) },
    handleReorder() { this.reorderDirty = false; this.$emit('reorder', this.segments) },
    onSegmentFieldChange(segment, field, value) { this.$emit("update-segment-field", { segment, field, value }) },
    onConstraintFieldChange(segment, constraint, field, value) { this.$emit("update-constraint-field", { segment, constraint, field, value }) },
    handleCreateConstraint(element) {
      const c = this.newConstraints[element.id]
      this.$emit('create-constraint', { segment: element, constraint: { operator: c.operator, property: c.property, value: c.value } })
      this.newConstraints[element.id] = { operator: 'EQ', property: '', value: '' }
    }
  }
}
</script>

<style lang="less" scoped>
.segment-card {
  background: #fff;
  border: 1px solid var(--el-border-color);
  border-radius: 10px;
  padding: 10px 12px;
  margin-bottom: 10px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.04);
}
.seg-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.seg-id {
  font-size: 10px;
  font-weight: 600;
  color: var(--el-text-color-placeholder);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  letter-spacing: 0.02em;
}

.seg-header-fields {
  display: flex;
  gap: 6px;
  flex: 1;
  > * { flex: 1; }
}
.seg-header-actions {
  display: flex;
  gap: 3px;
  flex-shrink: 0;
}
.seg-panel-row {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}
.seg-panel {
  flex: 1;
  min-width: 0;
  background: var(--el-fill-color-light);
  border-radius: 8px;
  padding: 10px 12px;
}
.seg-panel-dist {
  flex: 0 0 220px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
  padding: 10px 12px;
}
.seg-section-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}
.seg-section-subtitle {
  font-weight: 400;
  color: var(--el-text-color-placeholder);
  text-transform: none;
  letter-spacing: 0;
}

// --- Constraint Input Group ---
.constraint-input-group {
  display: flex;
  flex: 1;
  gap: 4px;
}

.constraint-grid {
  display: grid;
  gap: 4px;
}
.constraint-row {
  display: grid;
  grid-template-columns: 36px 1fr auto;
  gap: 6px;
  align-items: center;
  background: var(--el-color-primary-light-9);
  border-radius: 6px;
  padding: 4px 6px;
}
.constraint-logic {
  font-size: 10px;
  font-weight: 700;
  color: var(--el-color-primary);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  text-align: right;
  padding-right: 2px;
  white-space: nowrap;
}
.constraint-actions {
  display: flex;
  gap: 3px;
}
.new-constraint-row {
  .constraint-input-group { flex: 1; }
  margin-top: 4px;
  padding: 6px 6px 4px;
  border-top: 1px dashed var(--el-border-color);
  border-radius: 0 0 6px 6px;
  background: var(--el-color-primary-light-9);
  grid-template-columns: 36px 1fr auto;
  .constraint-logic { align-self: flex-start; padding-top: 5px; }
}

// --- Distribution ---
.dist-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
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
  font-size: 12px;
  font-weight: 500;
  color: var(--el-text-color-regular);
}
.dist-pct {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  font-variant-numeric: tabular-nums;
}
</style>

