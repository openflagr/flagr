<template>
  <div class="constraint-value-cell">
    <el-input
      size="small"
      class="constraint-cell constraint-control"
      :placeholder="placeholder"
      :model-value="modelValue"
      :data-testid="dataTestid"
      @update:model-value="$emit('update:modelValue', $event)"
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
</template>

<script lang="ts">
import { InfoFilled } from '@element-plus/icons-vue'
import { contextKeyHint } from '@/helpers/contextKeyHints'

export default {
  name: 'ConstraintValueCell',
  components: { InfoFilled },
  props: {
    modelValue: { type: String, required: true },
    property: { type: String, required: true },
    placeholder: { type: String, default: '' },
    dataTestid: { type: String, default: '' },
  },
  emits: ['update:modelValue'],
  computed: {
    valueHint(): string | null {
      return contextKeyHint(this.property, this.modelValue)
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
