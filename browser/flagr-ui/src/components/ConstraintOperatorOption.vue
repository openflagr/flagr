<template>
  <el-tooltip
    :disabled="!helpText"
    placement="right"
    effect="light"
    :show-after="0"
    :enterable="true"
    :offset="10"
    popper-class="constraint-op-hint-tooltip constraint-op-option-hint-popper"
  >
    <template #content>
      <span
        v-if="helpText"
        class="constraint-op-hint-tooltip__text"
        :data-testid="`constraint-op-option-hint-${item.value}`"
      >{{ helpText }}</span>
    </template>
    <div class="constraint-op-option">
      <span class="constraint-op-option-text">{{ displayText }}</span>
      <el-tag
        size="small"
        type="info"
        effect="plain"
        disable-transitions
        class="constraint-op-api-tag"
      >
        {{ badge }}
      </el-tag>
    </div>
  </el-tooltip>
</template>

<script lang="ts">
import type { PropType } from 'vue'
import type { OperatorUiOption } from '@/helpers/constraintOperators'
import {
  operatorHelpText,
  operatorOptionDisplayText,
} from '@/helpers/constraintOperatorUi'

export default {
  name: 'ConstraintOperatorOption',
  props: {
    item: { type: Object as PropType<OperatorUiOption>, required: true },
  },
  computed: {
    displayText(): string {
      return operatorOptionDisplayText(this.item)
    },
    badge(): string {
      return this.item.exprToken
    },
    helpText(): string | null {
      return operatorHelpText(this.item)
    },
  },
}
</script>

<style scoped>
.constraint-op-option {
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: center;
  column-gap: 12px;
  width: 100%;
  min-width: 0;
  line-height: 1.35;
}
.constraint-op-option-text {
  font-size: var(--font-size-body);
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: left;
}
.constraint-op-api-tag {
  justify-self: end;
  flex-shrink: 0;
  font-family: var(--font-mono);
  font-weight: var(--font-weight-semibold);
  border: none;
}
</style>