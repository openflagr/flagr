<template>
  <el-tooltip
    placement="top"
    effect="light"
    :enterable="true"
    popper-class="jev-toggle-tooltip"
  >
    <template #content>
      <div class="jev-toggle-tooltip__body">
        Turn on to back this constraint with a Jev / System One question.
        The model answer is used as <code>@jev.&lt;name&gt;</code> and compared
        with the operator.
      </div>
    </template>
    <div
      class="constraint-jev"
      :class="{ 'constraint-jev--on': enabled }"
    >
      <el-switch
        :model-value="enabled"
        size="small"
        :data-testid="testId"
        @update:model-value="$emit('update:enabled', $event)"
      />
      <span
        class="constraint-jev__label"
        @click="$emit('update:enabled', !enabled)"
      >JEV</span>
    </div>
  </el-tooltip>
</template>

<script lang="ts">
export default {
  name: 'JevToggleSwitch',
  props: {
    enabled: { type: Boolean, required: true },
    testId: { type: String, default: 'constraint-jev-toggle' },
  },
  emits: ['update:enabled'],
}
</script>

<style scoped>
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
</style>
