<template>
  <div class="constraint-row">
    <span class="constraint-logic">{{ index === 0 ? 'IF' : 'AND' }}</span>
    <el-input
      size="small"
      class="constraint-cell constraint-control"
      :placeholder="jevEnabled ? 'question_name' : propertyPlaceholder"
      :model-value="jevEnabled ? jevName : constraint.property"
      :disabled="readonly"
      data-testid="constraint-prop-input"
      @update:model-value="setProperty"
    >
      <template #prefix>
        <span
          v-if="jevEnabled"
          class="jev-prefix"
        >@jev.</span>
      </template>
    </el-input>
    <template v-if="jevEnabled">
      <span
        class="constraint-cell jev-summary"
        data-testid="constraint-jev-summary"
      >{{ jevSummary }}</span>
    </template>
    <template v-else>
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
    </template>
    <div
      v-if="!readonly"
      class="constraint-actions"
    >
      <el-tooltip
        placement="top"
        effect="light"
        :enterable="true"
        popper-class="jev-toggle-tooltip"
      >
        <template #content>
          <div class="jev-toggle-tooltip__body">
            Turn on to back this constraint with a Jev / System One question.
            The model answer is used as <code>@jev.&lt;name&gt;</code> and
            compared with the operator.
          </div>
        </template>
        <div
          class="constraint-jev"
          :class="{ 'constraint-jev--on': jevEnabled }"
        >
          <el-switch
            :model-value="jevEnabled"
            size="small"
            data-testid="constraint-jev-toggle"
            @update:model-value="toggleJev"
          />
          <span
            class="constraint-jev__label"
            @click="toggleJev(!jevEnabled)"
          >JEV</span>
        </div>
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
        :operator="constraint.operator"
        :value="constraint.value"
        :disabled="readonly"
        @update:all="setAll"
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
  formatJevSummary,
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
    jevSummary(): string {
      return formatJevSummary(this.constraint.jev, this.constraint.operator, this.constraint.value)
    },
    propertyPlaceholder(): string {
      return propertyPlaceholderFor(this.uiOperator, this.operatorOptions)
    },
    valuePlaceholder(): string {
      return valuePlaceholderFor(this.uiOperator, this.operatorOptions)
    },
  },
  methods: {
    onField(field: 'property' | 'operator' | 'value', value: string) {
      this.$emit('update-field', { field, value })
    },
    setProperty(value: string) {
      if (this.jevEnabled) this.setJevName(value)
      else this.onField('property', value)
    },
    onOperator(uiOperator: string) {
      this.$emit('update-operator', { uiOperator })
    },
    toggleJev(enabled: boolean) {
      if (enabled) {
        this.$emit('update-jev', {
          jev: this.constraint.jev ?? defaultJevQuestion('noul'),
          property: jevPropertyFor(this.jevName || 'question'),
          operator: 'GTE',
          value: '0.70',
        })
      } else {
        this.$emit('update-jev', { jev: undefined, property: '' })
      }
    },
    setJevName(name: string) {
      this.$emit('update-jev', { property: jevPropertyFor(name) })
    },
    setAll(payload: { jev: JevQuestion; operator: string; value: string }) {
      this.$emit('update-jev', payload)
    },
  },
}
</script>

<style scoped>
.jev-panel {
  grid-column: 2 / -1;
  margin: var(--space-3xs) 0 var(--space-2xs);
  padding: var(--space-2xs) var(--space-xs);
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: var(--radius-md);
}
@media (max-width: 768px) {
  .jev-panel {
    grid-column: 1 / -1;
  }
}
.jev-prefix {
  font-family: var(--font-mono);
  font-size: var(--font-size-body-sm);
  color: var(--el-color-primary);
}
.constraint-jev {
  display: flex;
  align-items: center;
  gap: var(--space-3xs);
  margin-right: var(--space-3xs);
}
.constraint-jev__label {
  font-size: var(--font-size-micro);
  font-weight: var(--font-weight-semibold);
  letter-spacing: var(--letter-spacing-wide);
  text-transform: uppercase;
  color: var(--el-text-color-placeholder);
  cursor: pointer;
  user-select: none;
}
.constraint-jev--on .constraint-jev__label {
  color: var(--el-color-primary);
}
.jev-summary {
  display: flex;
  align-items: center;
  min-height: var(--el-component-size-small);
  padding: 0 var(--space-2xs);
  font-family: var(--font-mono);
  font-size: var(--font-size-body-sm);
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: var(--radius-sm);
  grid-column: span 2;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
