<template>
  <el-card class="segments-container">
    <template #header>
      <div class="el-card-header">
        <div class="flex-row">
          <div class="flex-row-left"><h2>Segments</h2></div>
          <div class="flex-row-right">
            <el-tooltip :content="reorderDirty ? 'Unsaved reorder — click to persist' : 'Use up/down buttons to reorder, then click Reorder to persist'" placement="top" effect="light">
              <el-button :type="reorderDirty ? 'warning' : undefined" @click="handleReorder">Reorder{{ reorderDirty ? ' *' : '' }}</el-button>
            </el-tooltip>
            <el-button @click="$emit('new-segment')" data-testid="open-new-segment-btn">New Segment</el-button>
          </div>
        </div>
      </div>
    </template>

    <div class="segments-container-inner" v-if="segments.length">
      <div v-for="(element, index) in segments" :key="element.id">
        <el-card shadow="hover" class="segment">
          <!-- Segment header -->
          <div class="flex-row id-row">
            <div class="flex-row-left">
              <el-tag type="primary">Segment ID: <b>{{ element.id }}</b></el-tag>
            </div>
            <div class="flex-row-right">
              <el-tooltip content="Move segment up" placement="top" effect="light">
                <el-button size="small" :disabled="index === 0" @click="handleMoveUp(element, index)" data-testid="move-segment-up-btn">
                  <el-icon><ArrowUp /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip content="Move segment down" placement="top" effect="light">
                <el-button size="small" :disabled="index === segments.length - 1" @click="handleMoveDown(element, index)" data-testid="move-segment-down-btn">
                  <el-icon><ArrowDown /></el-icon>
                </el-button>
              </el-tooltip>
              <el-button size="small" @click="$emit('save-segment', element)" data-testid="save-segment-btn">Save Segment Setting</el-button>
              <el-button @click="$emit('delete-segment', element)" size="small" data-testid="delete-segment-btn">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>

          <!-- Description + Rollout -->
          <el-row :gutter="10" class="id-row">
            <el-col :span="15">
              <el-input size="small" placeholder="Description" :model-value="element.description" @update:model-value="onSegmentFieldChange(element, 'description', $event)" data-testid="segment-desc-input">
                <template #prepend>Description</template>
              </el-input>
            </el-col>
            <el-col :span="9">
              <el-input class="segment-rollout-percent" size="small" placeholder="0" :model-value="element.rolloutPercent" @update:model-value="onSegmentFieldChange(element, 'rolloutPercent', $event)" data-testid="segment-rollout-input" :min="0" :max="100">
                <template #prepend>Rollout</template>
                <template #append>%</template>
              </el-input>
            </el-col>
          </el-row>

          <!-- Constraints -->
          <el-row>
            <el-col :span="24">
              <h5>Constraints (match ALL of them)</h5>
              <div class="constraints">
                <div class="constraints-inner" v-if="element.constraints.length">
                  <div v-for="constraint in element.constraints" :key="constraint.id">
                    <el-row :gutter="3" class="segment-constraint">
                      <el-col :span="20">
                        <el-input size="small" placeholder="Property" :model-value="constraint.property" @update:model-value="onConstraintFieldChange(element, constraint, 'property', $event)" data-testid="constraint-prop-input">
                          <template #prepend>Property</template>
                        </el-input>
                      </el-col>
                      <el-col :span="4">
                        <el-select class="width--full" size="small" :model-value="constraint.operator" @update:model-value="onConstraintFieldChange(element, constraint, 'operator', $event)" placeholder="operator" data-testid="constraint-op-select">
                          <el-option v-for="item in operatorOptions" :key="item.value" :label="item.label" :value="item.value"></el-option>
                        </el-select>
                      </el-col>
                      <el-col :span="20">
                        <el-input size="small" :model-value="constraint.value" @update:model-value="onConstraintFieldChange(element, constraint, 'value', $event)" data-testid="constraint-value-input">
                          <template #prepend>Value&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</template>
                        </el-input>
                      </el-col>
                      <el-col :span="2">
                        <el-button type="success" plain class="width--full" @click="$emit('save-constraint', { segment: element, constraint })" size="small" data-testid="save-constraint-btn">Save</el-button>
                      </el-col>
                      <el-col :span="2">
                        <el-button type="danger" plain class="width--full" @click="() => $emit('delete-constraint', { segment: element, constraint })" size="small" data-testid="delete-constraint-btn">
                          <el-icon><Delete /></el-icon>
                        </el-button>
                      </el-col>
                    </el-row>
                  </div>
                </div>
                <div class="card--empty" v-else>
                  <span>No constraints (ALL will pass)</span>
                </div>

                <!-- New constraint row -->
                <div>
                  <el-row :gutter="3">
                    <el-col :span="5">
                      <el-input size="small" placeholder="Property" v-model="element._newConstraint.property" data-testid="new-constraint-prop-input"></el-input>
                    </el-col>
                    <el-col :span="4">
                      <el-select size="small" v-model="element._newConstraint.operator" placeholder="operator" data-testid="new-constraint-op-select">
                        <el-option v-for="item in operatorOptions" :key="item.value" :label="item.label" :value="item.value"></el-option>
                      </el-select>
                    </el-col>
                    <el-col :span="11">
                      <el-input size="small" v-model="element._newConstraint.value" data-testid="new-constraint-value-input"></el-input>
                    </el-col>
                    <el-col :span="4">
                      <el-button class="width--full" size="small" type="primary" plain :disabled="!element._newConstraint.property || !element._newConstraint.value" @click.prevent="() => $emit('create-constraint', element)" data-testid="add-constraint-btn">Add Constraint</el-button>
                    </el-col>
                  </el-row>
                </div>
              </div>
            </el-col>

            <!-- Distribution display -->
            <el-col :span="24" class="segment-distributions">
              <h5>
                <span>Distribution</span>
                <el-button round size="small" @click="$emit('edit-distribution', element)">
                  <el-icon><Edit /></el-icon> edit
                </el-button>
              </h5>
              <el-row v-if="element.distributions.length" :gutter="20">
                <el-col v-for="distribution in element.distributions" :key="distribution.id" :span="6">
                  <el-card shadow="never" class="distribution-card">
                    <div><span>{{ distribution.variantKey }}</span></div>
                    <el-progress type="circle" color="#74E5E0" :width="70" :percentage="distribution.percent"></el-progress>
                  </el-card>
                </el-col>
              </el-row>
              <div class="card--error" v-else>No distribution yet</div>
            </el-col>
          </el-row>
        </el-card>
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
  props: {
    segments: { type: Array, required: true },
    operatorOptions: { type: Array, required: true }
  },
  emits: [
    "move-up",
    "move-down",
    "reorder",
    "new-segment",
    "save-segment",
    "delete-segment",
    "update-segment-field",
    "update-constraint-field",
    "save-constraint",
    "delete-constraint",
    "create-constraint",
    "edit-distribution"
  ],
  data() {
    return {
      reorderDirty: false
    }
  },
  methods: {

    handleMoveUp(element, index) {
      this.reorderDirty = true
      this.$emit('move-up', element, index)
    },
    handleMoveDown(element, index) {
      this.reorderDirty = true
      this.$emit('move-down', element, index)
    },
    handleReorder() {
      this.reorderDirty = false
      this.$emit('reorder', this.segments)
    },
    onSegmentFieldChange(segment, field, value) {
      this.$emit("update-segment-field", { segment, field, value })
    },
    onConstraintFieldChange(segment, constraint, field, value) {
      this.$emit("update-constraint-field", { segment, constraint, field, value })
    }
  }
}
</script>
