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
        <JevToggleSwitch
          :enabled="jevEnabled"
          test-id="new-constraint-jev-toggle"
          @update:enabled="toggleJev"
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
import JevToggleSwitch from '@/components/JevToggleSwitch.vue'
import {
  propertyPlaceholderFor,
  valuePlaceholderFor,
} from '@/helpers/constraintOperatorUi'
import {
  formatJevSummary,
  isJevQuestionReady,
  jevEnablePatch,
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
    JevToggleSwitch,
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
          ...jevEnablePatch(this.draft.jev, this.draft.property),
        })
      } else {
        this.$emit('update:draft', { ...this.draft, jev: undefined, property: '' })
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
</style>
