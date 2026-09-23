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
      >
        <template
          v-if="propertyGroups.length"
          #append
        >
          <el-popover
            placement="bottom-end"
            :width="280"
            trigger="click"
          >
            <template #reference>
              <el-button
                size="small"
                :icon="ArrowDown"
                data-testid="pick-enriched-property-btn"
              />
            </template>
            <div class="enricher-picker">
              <div
                v-for="group in propertyGroups"
                :key="group.namespace"
                class="enricher-picker-group"
              >
                <div class="enricher-picker-label">
                  {{ group.label }}
                </div>
                <div class="enricher-picker-options">
                  <el-tag
                    v-for="option in group.options"
                    :key="option"
                    size="small"
                    class="enricher-picker-option"
                    :data-testid="`enriched-property-${option}`"
                    @click="patch('property', option)"
                  >
                    {{ option }}
                  </el-tag>
                </div>
              </div>
            </div>
          </el-popover>
        </template>
      </el-input>
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
import { ArrowDown } from '@element-plus/icons-vue'
import ConstraintValueCell from '@/components/ConstraintValueCell.vue'
import ConstraintOperatorSelect from '@/components/ConstraintOperatorSelect.vue'
import {
  propertyPlaceholderFor,
  valuePlaceholderFor,
} from '@/helpers/constraintOperatorUi'
import type { OperatorOptionGroup, OperatorUiOption } from '@/helpers/constraintOperators'
import type { EnricherPropertyGroup } from '@/helpers/enricherOptions'

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
    propertyGroups: {
      type: Array as PropType<EnricherPropertyGroup[]>,
      default: () => [],
    },
  },
  emits: ['update:draft', 'add'],
  data() {
    return { ArrowDown }
  },
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

.enricher-picker {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.enricher-picker-label {
  color: var(--el-text-color-secondary);
  font-size: var(--font-size-caption);
  margin-bottom: var(--space-3xs);
}

.enricher-picker-options {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3xs);
}

.enricher-picker-option {
  cursor: pointer;
}
</style>
