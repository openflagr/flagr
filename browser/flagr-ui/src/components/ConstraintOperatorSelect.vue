<template>
  <div class="constraint-op-cell constraint-cell">
    <el-select
      :model-value="modelValue"
      size="small"
      placeholder="Operator"
      popper-class="constraint-op-select-popper"
      :data-testid="testId"
      class="constraint-control constraint-op-select-wrap"
      @update:model-value="$emit('update:modelValue', $event)"
    >
      <template #label>
        {{ closedBadge }}
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
          :data-testid="`constraint-op-option-${item.value}`"
        >
          <ConstraintOperatorOption :item="item" />
        </el-option>
      </el-option-group>
    </el-select>
    <el-tooltip
      v-if="hintLine"
      placement="top"
      effect="light"
      :show-after="90"
      :enterable="true"
      popper-class="constraint-op-hint-tooltip"
    >
      <template #content>
        <span
          class="constraint-op-hint-tooltip__text"
          :data-testid="hintTestId || undefined"
        >{{ hintLine }}</span>
      </template>
      <button
        type="button"
        class="constraint-op-hint-icon"
        :data-testid="hintIconTestId"
        :aria-label="hintAriaLabel"
      >
        <el-icon>
          <InfoFilled />
        </el-icon>
      </button>
    </el-tooltip>
  </div>
</template>

<script lang="ts">
import type { PropType } from 'vue'
import { InfoFilled } from '@element-plus/icons-vue'
import ConstraintOperatorOption from '@/components/ConstraintOperatorOption.vue'
import type { OperatorOptionGroup, OperatorUiOption } from '@/helpers/constraintOperators'
import {
  getOperatorHintLine,
  operatorSelectClosedBadge,
  operatorSelectLabel,
} from '@/helpers/constraintOperatorUi'

export default {
  name: 'ConstraintOperatorSelect',
  components: { ConstraintOperatorOption, InfoFilled },
  props: {
    modelValue: { type: String, default: '' },
    groupedOperatorOptions: {
      type: Array as PropType<OperatorOptionGroup[]>,
      required: true,
    },
    operatorOptions: {
      type: Array as PropType<OperatorUiOption[]>,
      required: true,
    },
    testId: { type: String, required: true },
    hintTestId: { type: String, default: '' },
  },
  emits: ['update:modelValue'],
  computed: {
    closedBadge(): string {
      return operatorSelectClosedBadge(this.modelValue, this.operatorOptions)
    },
    hintLine(): string | null {
      return getOperatorHintLine(this.modelValue, this.operatorOptions)
    },
    hintIconTestId(): string | undefined {
      return this.hintTestId ? `${this.hintTestId}-icon` : undefined
    },
    hintAriaLabel(): string {
      return 'Operator help'
    },
  },
  methods: {
    operatorSelectLabel,
  },
}
</script>

<style scoped>
.constraint-op-cell {
  display: flex;
  align-items: center;
  gap: var(--space-3xs);
  min-width: 0;
  width: 100%;
}

.constraint-op-select-wrap {
  flex: 1;
  min-width: 0;
}

.constraint-op-hint-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 1.25rem;
  height: 1.25rem;
  padding: 0;
  border: none;
  border-radius: var(--el-border-radius-small);
  background: transparent;
  color: var(--el-text-color-placeholder);
  cursor: help;
  transition: color 0.15s ease;
}

.constraint-op-hint-icon:hover,
.constraint-op-hint-icon:focus-visible {
  color: var(--el-text-color-secondary);
  outline: none;
}

.constraint-op-hint-icon .el-icon {
  font-size: 0.95rem;
}
</style>