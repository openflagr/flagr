<template>
  <div class="constraint-row">
    <span class="constraint-logic">{{ index === 0 ? 'IF' : 'AND' }}</span>
    <el-input
      size="small"
      class="constraint-cell constraint-control"
      :placeholder="propertyPlaceholder"
      :model-value="constraint.property"
      data-testid="constraint-prop-input"
      @update:model-value="onField('property', $event)"
    />
    <ConstraintOperatorSelect
      :model-value="uiOperator"
      :grouped-operator-options="groupedOperatorOptions"
      :operator-options="operatorOptions"
      test-id="constraint-op-select"
      @update:model-value="onOperator"
    />
    <div class="constraint-value-cell">
      <el-input
        size="small"
        class="constraint-cell constraint-control"
        :placeholder="valuePlaceholder"
        :model-value="valueForInput"
        data-testid="constraint-value-input"
        @update:model-value="onField('value', $event)"
      />
      <el-tooltip
        v-if="valueHint"
        :content="valueHint"
        placement="top"
        effect="light"
      >
        <el-icon class="constraint-value-hint-icon">
          <InfoFilled />
        </el-icon>
      </el-tooltip>
    </div>
    <div class="constraint-actions">
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
  </div>
</template>

<script lang="ts">
import type { PropType } from 'vue'
import type { Constraint } from '@/api/types'

import ConstraintOperatorSelect from '@/components/ConstraintOperatorSelect.vue'
import {
  propertyPlaceholderFor,
  valuePlaceholderFor,
} from '@/helpers/constraintOperatorUi'
import { constraintValueForInput, resolveUiOperator } from '@/helpers/constraintOperatorSugar'
import { contextKeyHint } from '@/helpers/contextKeyHints'
import type { OperatorOptionGroup, OperatorUiOption } from '@/helpers/constraintOperators'
import { Delete, InfoFilled } from '@element-plus/icons-vue'
import { SAVE_DIRTY_TOOLTIP } from '@/helpers/saveDirtyUi'

export default {
  name: 'ConstraintExistingRow',
  components: {
    ConstraintOperatorSelect,
    Delete,
    InfoFilled,
  },
  props: {
    constraint: { type: Object as PropType<Constraint>, required: true },
    index: { type: Number, required: true },
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
  emits: ['update-field', 'update-operator', 'save', 'delete'],
  computed: {
    uiOperator(): string {
      return resolveUiOperator(this.constraint.operator, this.constraint.value)
    },
    valueForInput(): string {
      return constraintValueForInput(this.constraint)
    },
    propertyPlaceholder(): string {
      return propertyPlaceholderFor(this.uiOperator, this.operatorOptions)
    },
    valuePlaceholder(): string {
      return valuePlaceholderFor(this.uiOperator, this.operatorOptions)
    },
    valueHint(): string | null {
      return contextKeyHint(this.constraint.property, this.constraint.value)
    },
  },
  methods: {
    onField(field: 'property' | 'value', value: string) {
      this.$emit('update-field', { field, value })
    },
    onOperator(uiOperator: string) {
      this.$emit('update-operator', { uiOperator })
    },
  },
}
</script>

<style scoped>
.constraint-value-cell {
  display: flex;
  align-items: center;
  gap: var(--space-3xs);
}
.constraint-value-cell .constraint-control {
  flex: 1;
  min-width: 0;
}
.constraint-value-hint-icon {
  color: var(--el-color-success-light-5);
  cursor: help;
  font-size: var(--font-size-body-sm);
}
.constraint-value-hint-icon:hover {
  color: var(--el-color-success);
}
</style>
