<template>
  <div class="constraint-op-option">
    <span class="constraint-op-option-text">{{ displayText }}</span>
    <el-tag
      size="small"
      type="info"
      effect="plain"
      class="constraint-op-api-tag"
    >
      {{ operatorApiBadge(item.value) }}
    </el-tag>
  </div>
</template>

<script lang="ts">
import type { PropType } from 'vue'
import type { OperatorUiOption } from '@/helpers/constraintOperators'
import { operatorApiBadge, operatorOptionDisplayText } from '@/helpers/constraintOperatorUi'

export default {
  name: 'ConstraintOperatorOption',
  props: {
    item: { type: Object as PropType<OperatorUiOption>, required: true },
  },
  computed: {
    displayText(): string {
      return operatorOptionDisplayText(this.item)
    },
  },
  methods: {
    operatorApiBadge,
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
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-weight: var(--font-weight-semibold);
  border: none;
}
</style>