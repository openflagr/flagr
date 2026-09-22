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
        :placeholder="jevEnabled ? 'question_name' : propertyPlaceholder"
        :model-value="jevEnabled ? jevName : draft.property"
        data-testid="new-constraint-prop-input"
        @update:model-value="setProperty"
      >
        <template #prefix>
          <span
            v-if="jevEnabled"
            class="jev-prefix"
          >@jev.</span>
        </template>
        <template #suffix>
          <span
            class="jev-toggle"
            :class="{ 'jev-toggle--on': jevEnabled }"
            data-testid="new-constraint-jev-toggle"
            @click.stop.prevent="toggleJev(!jevEnabled)"
          >Jev</span>
        </template>
      </el-input>
      <template v-if="jevEnabled">
        <span
          class="constraint-cell jev-summary"
          data-testid="new-constraint-jev-summary"
        >{{ jevSummary }}</span>
      </template>
      <template v-else>
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
      </template>
      <div class="constraint-actions">
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
    <div
      v-if="jevEnabled"
      class="jev-panel"
    >
      <JevQuestionEditor
        :model-value="draft.jev!"
        :operator="draft.operator"
        :value="draft.value"
        @update:all="setAll"
      />
    </div>
  </div>
</template>

<script lang="ts">
import type { PropType } from 'vue'
import type { JevQuestion } from '@/api/types'
import ConstraintValueCell from '@/components/ConstraintValueCell.vue'
import ConstraintOperatorSelect from '@/components/ConstraintOperatorSelect.vue'
import JevQuestionEditor from '@/components/JevQuestionEditor.vue'
import {
  propertyPlaceholderFor,
  valuePlaceholderFor,
} from '@/helpers/constraintOperatorUi'
import {
  defaultJevQuestion,
  formatJevSummary,
  isJevQuestionReady,
  jevPropertyFor,
  jevPropertyName,
} from '@/helpers/jevQuestion'
import type { OperatorOptionGroup, OperatorUiOption } from '@/helpers/constraintOperators'

export interface NewConstraintDraft {
  operator: string
  property: string
  value: string
  jev?: JevQuestion
}

export default {
  name: 'ConstraintAddRow',
  components: {
    ConstraintOperatorSelect,
    ConstraintValueCell,
    JevQuestionEditor,
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
    jevEnabled(): boolean {
      return Boolean(this.draft.jev)
    },
    jevName(): string {
      return jevPropertyName(this.draft.property)
    },
    jevSummary(): string {
      return formatJevSummary(this.draft.jev, this.draft.operator, this.draft.value)
    },
    propertyPlaceholder(): string {
      return propertyPlaceholderFor(this.draft.operator, this.operatorOptions)
    },
    valuePlaceholder(): string {
      return valuePlaceholderFor(this.draft.operator, this.operatorOptions)
    },
    canAdd(): boolean {
      const d = this.draft
      if (d.jev) {
        return Boolean(d.operator && d.property && isJevQuestionReady(d.jev))
      }
      return Boolean(d.operator && d.property && d.value)
    },
  },
  methods: {
    patch(field: 'property' | 'value' | 'operator', value: string) {
      this.$emit('update:draft', { ...this.draft, [field]: value })
    },
    setProperty(value: string) {
      if (this.jevEnabled) this.setJevName(value)
      else this.patch('property', value)
    },
    toggleJev(enabled: boolean) {
      if (enabled) {
        this.$emit('update:draft', {
          ...this.draft,
          property: jevPropertyFor(this.jevName || 'question'),
          operator: 'GTE',
          value: '0.70',
          jev: this.draft.jev ?? defaultJevQuestion('noul'),
        })
      } else {
        this.$emit('update:draft', { ...this.draft, jev: undefined })
      }
    },
    setJevName(name: string) {
      this.$emit('update:draft', { ...this.draft, property: jevPropertyFor(name) })
    },
    setAll(payload: Partial<NewConstraintDraft>) {
      this.$emit('update:draft', { ...this.draft, ...payload })
    },
  },
}
</script>

<style scoped>
.constraint-add-block {
  display: contents;
}
.jev-panel {
  grid-column: 1 / -1;
  margin: var(--space-3xs) 0 var(--space-2xs);
  padding: var(--space-2xs) var(--space-xs);
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: var(--radius-md);
}
.jev-prefix {
  font-family: var(--font-mono);
  font-size: var(--font-size-body-sm);
  color: var(--el-color-primary);
}
.jev-toggle {
  font-size: var(--font-size-micro);
  font-weight: var(--font-weight-semibold);
  letter-spacing: var(--letter-spacing-wide);
  text-transform: uppercase;
  line-height: 1.6;
  padding: 0 var(--space-3xs);
  border-radius: var(--radius-sm);
  color: var(--el-text-color-placeholder);
  cursor: pointer;
  user-select: none;
}
.jev-toggle:hover {
  color: var(--el-color-primary);
}
.jev-toggle--on {
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}
.jev-summary {
  display: flex;
  align-items: center;
  min-height: var(--el-component-size-small);
  padding: 0 var(--space-2xs);
  font-family: var(--font-mono);
  font-size: var(--font-size-body-sm);
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: var(--radius-sm);
  grid-column: span 2;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
