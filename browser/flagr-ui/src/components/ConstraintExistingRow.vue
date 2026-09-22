<template>
  <div class="constraint-row">
    <span class="constraint-logic">{{ index === 0 ? 'IF' : 'AND' }}</span>
    <el-input
      v-if="jevEnabled"
      size="small"
      class="constraint-cell constraint-control"
      placeholder="question name"
      :model-value="jevName"
      :disabled="readonly"
      data-testid="constraint-jev-name"
      @update:model-value="setJevName"
    />
    <el-input
      v-else
      size="small"
      class="constraint-cell constraint-control"
      :placeholder="propertyPlaceholder"
      :model-value="constraint.property"
      :disabled="readonly"
      data-testid="constraint-prop-input"
      @update:model-value="onField('property', $event)"
    />
    <ConstraintOperatorSelect
      :model-value="uiOperator"
      :disabled="readonly"
      :grouped-operator-options="groupedOperatorOptions"
      :operator-options="operatorOptions"
      test-id="constraint-op-select"
      @update:model-value="onOperator"
    />
    <ConstraintValueCell
      :model-value="valueForInput"
      :property="constraint.property"
      :placeholder="valuePlaceholder"
      :disabled="readonly"
      data-testid="constraint-value-input"
      @update:model-value="onField('value', $event)"
    />
    <div
      v-if="!readonly"
      class="constraint-actions"
    >
      <el-tooltip
        content="Back this constraint with a Jev / System One question"
        placement="top"
        effect="light"
      >
        <el-checkbox
          :model-value="jevEnabled"
          size="small"
          class="constraint-jev-toggle"
          data-testid="constraint-jev-toggle"
          @update:model-value="toggleJev"
        >
          Jev
        </el-checkbox>
      </el-tooltip>
      <el-tooltip
        :content="saveDirtyTooltip"
        placement="top"
        effect="light"
        :disabled="!dirty"
      >
        <el-button
          size="small"
          :plain="!dirty"
          :type="saveButtonType"
          data-testid="save-constraint-btn"
          @click="$emit('save')"
        >
          {{ saveButtonLabel }}
        </el-button>
      </el-tooltip>
      <el-button
        size="small"
        plain
        data-testid="delete-constraint-btn"
        @click="$emit('delete')"
      >
        <el-icon><Delete /></el-icon>
      </el-button>
    </div>
    <div
      v-if="jevEnabled"
      class="jev-panel"
    >
      <JevQuestionEditor
        :model-value="constraint.jev!"
        :disabled="readonly"
        @update:model-value="setJev"
      />
    </div>
  </div>
</template>

<script lang="ts">
import type { PropType } from 'vue'
import type { Constraint, JevQuestion } from '@/api/types'

import ConstraintOperatorSelect from '@/components/ConstraintOperatorSelect.vue'
import ConstraintValueCell from '@/components/ConstraintValueCell.vue'
import JevQuestionEditor from '@/components/JevQuestionEditor.vue'
import {
  propertyPlaceholderFor,
  valuePlaceholderFor,
} from '@/helpers/constraintOperatorUi'
import { constraintValueForInput, resolveUiOperator } from '@/helpers/constraintOperatorSugar'
import {
  defaultJevQuestion,
  jevPropertyFor,
  jevPropertyName,
} from '@/helpers/jevQuestion'
import type { OperatorOptionGroup, OperatorUiOption } from '@/helpers/constraintOperators'
import { Delete } from '@element-plus/icons-vue'
import { SAVE_DIRTY_TOOLTIP } from '@/helpers/saveDirtyUi'

export default {
  name: 'ConstraintExistingRow',
  components: {
    ConstraintOperatorSelect,
    Delete,
    ConstraintValueCell,
    JevQuestionEditor,
  },
  props: {
    constraint: { type: Object as PropType<Constraint>, required: true },
    index: { type: Number, required: true },
    readonly: { type: Boolean, default: false },
    operatorOptions: {
      type: Array as PropType<OperatorUiOption[]>,
      required: true,
    },
    groupedOperatorOptions: {
      type: Array as PropType<OperatorOptionGroup[]>,
      required: true,
    },
    dirty: { type: Boolean, default: false },
    saveButtonLabel: { type: String, required: true },
    saveButtonType: { type: String, default: undefined },
    saveDirtyTooltip: { type: String, default: SAVE_DIRTY_TOOLTIP },
  },
  emits: ['update-field', 'update-operator', 'update-jev', 'save', 'delete'],
  computed: {
    uiOperator(): string {
      return resolveUiOperator(this.constraint.operator, this.constraint.value)
    },
    valueForInput(): string {
      return constraintValueForInput(this.constraint)
    },
    jevEnabled(): boolean {
      return Boolean(this.constraint.jev)
    },
    jevName(): string {
      return jevPropertyName(this.constraint.property)
    },
    propertyPlaceholder(): string {
      return propertyPlaceholderFor(this.uiOperator, this.operatorOptions)
    },
    valuePlaceholder(): string {
      return valuePlaceholderFor(this.uiOperator, this.operatorOptions)
    },
  },
  methods: {
    onField(field: 'property' | 'value', value: string) {
      this.$emit('update-field', { field, value })
    },
    onOperator(uiOperator: string) {
      this.$emit('update-operator', { uiOperator })
    },
    toggleJev(enabled: boolean) {
      if (enabled) {
        this.$emit('update-jev', {
          jev: this.constraint.jev ?? defaultJevQuestion('noul'),
          property: jevPropertyFor(this.jevName || 'question'),
        })
      } else {
        this.$emit('update-jev', { jev: undefined })
      }
    },
    setJevName(name: string) {
      this.$emit('update-jev', { property: jevPropertyFor(name) })
    },
    setJev(jev: JevQuestion) {
      this.$emit('update-jev', { jev })
    },
  },
}
</script>

<style scoped>
.jev-panel {
  grid-column: 1 / -1;
  margin: var(--space-3xs) 0 var(--space-2xs);
  padding: var(--space-3xs) var(--space-2xs);
  border: 1px dashed var(--el-border-color);
  border-radius: var(--radius-md);
}
.constraint-jev-toggle {
  margin-right: var(--space-3xs);
}
</style>
