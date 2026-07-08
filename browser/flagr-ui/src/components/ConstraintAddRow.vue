<template>
  <div class="constraint-add-block">
    <div
      v-if="showDivider"
      class="constraint-add-divider"
      aria-hidden="true"
    />
    <p
      v-if="showCaption"
      class="constraint-add-caption"
    >
      {{ caption }}
    </p>
    <div class="constraint-row constraint-row--add">
      <span
        class="constraint-logic constraint-logic--add"
        aria-hidden="true"
      >+</span>
      <el-input
        size="small"
        class="constraint-cell constraint-control"
        :placeholder="propertyPlaceholder"
        :model-value="draft.property"
        data-testid="new-constraint-prop-input"
        @update:model-value="patch('property', $event)"
      />
      <ConstraintOperatorSelect
        :model-value="draft.operator"
        :grouped-operator-options="groupedOperatorOptions"
        :operator-options="operatorOptions"
        test-id="new-constraint-op-select"
        @update:model-value="patch('operator', $event)"
      />
      <ConstraintValueCell
        :model-value="draft.value"
        :property="draft.property"
        :placeholder="valuePlaceholder"
        data-testid="new-constraint-value-input"
        @update:model-value="patch('value', $event)"
        @keyup.enter="canAdd && $emit('add')"
      />
      <el-button
        size="small"
        type="primary"
        plain
        class="constraint-add-btn"
        data-testid="add-constraint-btn"
        :disabled="!canAdd"
        @click.prevent="$emit('add')"
      >
        Add constraint
      </el-button>
    </div>
  </div>
</template>

<script lang="ts">
import type { PropType } from 'vue'
import ConstraintValueCell from '@/components/ConstraintValueCell.vue'
import ConstraintOperatorSelect from '@/components/ConstraintOperatorSelect.vue'
import {
  propertyPlaceholderFor,
  valuePlaceholderFor,
} from '@/helpers/constraintOperatorUi'
import type { OperatorOptionGroup, OperatorUiOption } from '@/helpers/constraintOperators'

export interface NewConstraintDraft {
  operator: string
  property: string
  value: string
}

export default {
  name: 'ConstraintAddRow',
  components: {
    ConstraintOperatorSelect,
    ConstraintValueCell,
  },
  props: {
    draft: { type: Object as PropType<NewConstraintDraft>, required: true },
    operatorOptions: {
      type: Array as PropType<OperatorUiOption[]>,
      required: true,
    },
    groupedOperatorOptions: {
      type: Array as PropType<OperatorOptionGroup[]>,
      required: true,
    },
    showDivider: { type: Boolean, default: false },
    showCaption: { type: Boolean, default: false },
    caption: { type: String, default: '' },
  },
  emits: ['update:draft', 'add'],
  computed: {
    propertyPlaceholder(): string {
      return propertyPlaceholderFor(this.draft.operator, this.operatorOptions)
    },
    valuePlaceholder(): string {
      return valuePlaceholderFor(this.draft.operator, this.operatorOptions)
    },
    canAdd(): boolean {
      const d = this.draft
      return Boolean(d.operator && d.property && d.value)
    },
  },
  methods: {
    patch(field: keyof NewConstraintDraft, value: string) {
      this.$emit('update:draft', { ...this.draft, [field]: value })
    },
  },
}
</script>

<style scoped>
.constraint-add-block {
  display: contents;
}
</style>
