<template>
  <div
    class="constraint-op-cell constraint-cell"
    :class="{ 'constraint-op-cell--has-operator': Boolean(modelValue) }"
  >
    <div class="constraint-op-shell">
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
      <div
        class="constraint-op-help-slot"
        aria-hidden="true"
      >
        <el-tooltip
          v-if="helpText"
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
            >{{ helpText }}</span>
          </template>
          <button
            type="button"
            class="constraint-op-help-btn"
            :data-testid="hintIconTestId"
            :aria-label="hintAriaLabel"
          >
            <el-icon>
              <InfoFilled />
            </el-icon>
          </button>
        </el-tooltip>
        <span
          v-else
          class="constraint-op-help-btn constraint-op-help-btn--idle"
        />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import type { PropType } from 'vue'
import { InfoFilled } from '@element-plus/icons-vue'
import ConstraintOperatorOption from '@/components/ConstraintOperatorOption.vue'
import type { OperatorOptionGroup, OperatorUiOption } from '@/helpers/constraintOperators'
import {
  getOperatorHelpText,
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
    helpText(): string | null {
      return getOperatorHelpText(this.modelValue, this.operatorOptions)
    },
    hintIconTestId(): string | undefined {
      return this.hintTestId ? `${this.hintTestId}-icon` : undefined
    },
    hintAriaLabel(): string {
      return 'How this operator uses property and value'
    },
  },
  methods: {
    operatorSelectLabel,
  },
}
</script>

<style scoped>
.constraint-op-cell {
  min-width: 0;
  width: 100%;
}

.constraint-op-shell {
  display: flex;
  align-items: stretch;
  width: 100%;
  min-width: 0;
  border: 1px solid var(--el-border-color);
  border-radius: var(--el-border-radius-base);
  background: var(--el-fill-color-blank);
  overflow: hidden;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.constraint-op-cell--has-operator .constraint-op-shell:hover {
  border-color: var(--el-border-color-hover);
}

.constraint-op-select-wrap {
  flex: 1;
  min-width: 0;
}

.constraint-op-select-wrap :deep(.el-select__wrapper) {
  box-shadow: none !important;
  border: none !important;
  border-radius: 0;
  background: transparent;
  padding-right: 0.35rem;
}

.constraint-op-help-slot {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 1.75rem;
  border-left: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-light);
}

.constraint-op-help-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.5rem;
  padding: 0;
  margin: 0;
  border: none;
  border-radius: 0;
  background: transparent;
  color: var(--el-text-color-secondary);
  cursor: help;
  transition: color 0.15s ease, background-color 0.15s ease;
}

.constraint-op-help-btn:hover,
.constraint-op-help-btn:focus-visible {
  color: var(--el-color-primary);
  background: var(--el-fill-color);
  outline: none;
}

.constraint-op-help-btn--idle {
  cursor: default;
  pointer-events: none;
  opacity: 0.35;
}

.constraint-op-help-btn .el-icon {
  font-size: 0.9rem;
}
</style>