<template>
  <el-select
    :model-value="modelValue"
    size="small"
    placeholder="Operator"
    popper-class="constraint-op-select-popper"
    :data-testid="testId"
    class="constraint-cell constraint-control"
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
</template>

<script lang="ts">
import type { PropType } from 'vue'
import ConstraintOperatorOption from '@/components/ConstraintOperatorOption.vue'
import type { OperatorOptionGroup, OperatorUiOption } from '@/helpers/constraintOperators'
import {
  operatorSelectClosedBadge,
  operatorSelectLabel,
} from '@/helpers/constraintOperatorUi'

export default {
  name: 'ConstraintOperatorSelect',
  components: { ConstraintOperatorOption },
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
  },
  emits: ['update:modelValue'],
  computed: {
    closedBadge(): string {
      return operatorSelectClosedBadge(this.modelValue, this.operatorOptions)
    },
  },
  methods: {
    operatorSelectLabel,
  },
}
</script>