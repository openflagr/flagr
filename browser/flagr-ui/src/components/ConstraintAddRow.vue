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
      <div class="constraint-value-cell">
        <el-input
          size="small"
          class="constraint-cell constraint-control"
          :placeholder="valuePlaceholder"
          :model-value="draft.value"
          data-testid="new-constraint-value-input"
          @update:model-value="patch('value', $event)"
          @keyup.enter="canAdd && $emit('add')"
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
import { InfoFilled } from '@element-plus/icons-vue'
import ConstraintOperatorSelect from '@/components/ConstraintOperatorSelect.vue'
import {
  propertyPlaceholderFor,
  valuePlaceholderFor,
} from '@/helpers/constraintOperatorUi'
import { contextKeyHint } from '@/helpers/contextKeyHints'
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
    InfoFilled,
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
    valueHint(): string | null {
      return contextKeyHint(this.draft.property, this.draft.value)
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
