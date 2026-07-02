<template>
  <p
    v-if="line"
    class="constraint-hint-line"
    :data-testid="testId || undefined"
  >
    {{ line }}
  </p>
</template>

<script lang="ts">
import type { PropType } from 'vue'
import { getOperatorHintLine } from '@/helpers/constraintOperatorUi'
import type { OperatorUiOption } from '@/helpers/constraintOperators'

export default {
  name: 'ConstraintOperatorHint',
  props: {
    operator: { type: String, default: '' },
    operatorOptions: { type: Array as PropType<OperatorUiOption[]>, required: true },
    testId: { type: String, default: '' },
  },
  computed: {
    line(): string | null {
      return getOperatorHintLine(this.operator, this.operatorOptions)
    },
  },
}
</script>

<style scoped>
.constraint-hint-line {
  line-height: var(--line-height-tight);
  color: var(--el-text-color-placeholder);
}
</style>