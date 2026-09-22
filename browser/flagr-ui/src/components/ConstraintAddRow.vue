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
        v-if="jevEnabled"
        size="small"
        class="constraint-cell constraint-control"
        placeholder="question_name"
        :model-value="jevName"
        data-testid="new-constraint-jev-name"
        @update:model-value="setJevName"
      >
        <template #prepend>
          @jev.
        </template>
      </el-input>
      <el-input
        v-else
        size="small"
        class="constraint-cell constraint-control"
        :placeholder="propertyPlaceholder"
        :model-value="draft.property"
        data-testid="new-constraint-prop-input"
        @update:model-value="patch('property', $event)"
      />
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
        <el-tooltip
          content="Back this constraint with a Jev / System One question"
          placement="top"
          effect="light"
        >
          <el-checkbox
            :model-value="jevEnabled"
            size="small"
            class="constraint-jev-toggle"
            data-testid="new-constraint-jev-toggle"
            @update:model-value="toggleJev"
          >
            Jev
          </el-checkbox>
        </el-tooltip>
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
        @update:model-value="setJev"
        @update:operator="patch('operator', $event)"
        @update:value="patch('value', $event)"
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
    setJev(jev: JevQuestion) {
      this.$emit('update:draft', { ...this.draft, jev })
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
  padding: var(--space-3xs) var(--space-2xs);
  border: 1px dashed var(--el-border-color);
  border-radius: var(--radius-md);
}
.constraint-jev-toggle {
  margin-right: var(--space-3xs);
}
.jev-summary {
  font-size: var(--font-size-body-sm);
  font-family: var(--font-mono);
  color: var(--el-text-color-regular);
  grid-column: span 2;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
