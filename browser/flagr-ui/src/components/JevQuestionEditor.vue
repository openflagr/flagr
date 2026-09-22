<template>
  <div
    class="jev-editor"
    data-testid="jev-question-editor"
  >
    <div class="jev-field">
      <span class="jev-label">Question type</span>
      <el-select
        :model-value="question.type"
        size="small"
        class="jev-type"
        :disabled="disabled"
        data-testid="jev-type-select"
        @update:model-value="setType"
      >
        <el-option
          label="Noul (yes / no)"
          value="noul"
        />
        <el-option
          label="Choice (pick one)"
          value="choice"
        />
        <el-option
          label="Score (rate on a scale)"
          value="score"
        />
      </el-select>
    </div>

    <div class="jev-field jev-field--stack">
      <span class="jev-label">Instructions</span>
      <el-input
        type="textarea"
        :rows="2"
        size="small"
        placeholder="Question the model answers, e.g. Is `message` about billing?"
        :model-value="instructionsText"
        :disabled="disabled"
        data-testid="jev-instructions"
        @update:model-value="setInstructions"
      />
    </div>

    <div class="jev-field jev-field--stack">
      <span class="jev-label">{{ criteriaLabel }}</span>

      <div
        v-if="isNoul"
        class="jev-criteria"
      >
        <el-input
          size="small"
          placeholder="Optional: what &quot;yes&quot; means"
          :model-value="noulTrue"
          :disabled="disabled"
          @update:model-value="setNoul('true', $event)"
        />
        <el-input
          size="small"
          placeholder="Optional: what &quot;no&quot; means"
          :model-value="noulFalse"
          :disabled="disabled"
          @update:model-value="setNoul('false', $event)"
        />
      </div>

      <div
        v-else-if="isChoice"
        class="jev-criteria"
      >
        <div
          v-for="(row, i) in choiceRows"
          :key="i"
          class="jev-option-row"
        >
          <el-input
            size="small"
            placeholder="option"
            :model-value="row.name"
            :disabled="disabled"
            @update:model-value="setChoiceName(i, $event)"
          />
          <el-input
            size="small"
            placeholder="description"
            :model-value="row.description"
            :disabled="disabled"
            @update:model-value="setChoiceDescription(i, $event)"
          />
          <el-button
            size="small"
            plain
            :disabled="disabled"
            @click="removeChoice(i)"
          >
            ×
          </el-button>
        </div>
        <el-button
          size="small"
          plain
          :disabled="disabled"
          data-testid="jev-add-option"
          @click="addChoice"
        >
          Add option
        </el-button>
      </div>

      <div
        v-else
        class="jev-criteria"
      >
        <div
          v-for="(level, i) in scoreLevels"
          :key="i"
          class="jev-option-row"
        >
          <span class="jev-level-index">{{ i }}</span>
          <el-input
            size="small"
            placeholder="level description (low → high)"
            :model-value="level"
            :disabled="disabled"
            @update:model-value="setScoreLevelText(i, $event)"
          />
          <el-button
            size="small"
            plain
            :disabled="disabled || i === 0"
            @click="moveScoreLevel(i, -1)"
          >
            ↑
          </el-button>
          <el-button
            size="small"
            plain
            :disabled="disabled || i === scoreLevels.length - 1"
            @click="moveScoreLevel(i, 1)"
          >
            ↓
          </el-button>
          <el-button
            size="small"
            plain
            :disabled="disabled"
            @click="removeScoreLevel(i)"
          >
            ×
          </el-button>
        </div>
        <el-button
          size="small"
          plain
          :disabled="disabled"
          data-testid="jev-add-level"
          @click="addScoreLevel"
        >
          Add level
        </el-button>
      </div>
    </div>

    <div
      v-if="!isNoul"
      class="jev-field jev-field--stack"
    >
      <span class="jev-label">Match</span>
      <div class="jev-match">
        <template v-if="isChoice">
          <el-select
            class="jev-choice-select"
            size="small"
            multiple
            collapse-tags
            placeholder="Pick option(s)"
            :model-value="selectedOptions"
            :disabled="disabled"
            data-testid="jev-choice-options"
            @update:model-value="setChoiceOptions"
          >
            <el-option
              v-for="row in choiceRows"
              :key="row.name"
              :label="row.name"
              :value="row.name"
            />
          </el-select>
          <el-checkbox
            :model-value="choiceNegate"
            :disabled="disabled"
            data-testid="jev-choice-negate"
            @update:model-value="setChoiceNegate"
          >
            is not
          </el-checkbox>
        </template>

        <template v-else>
          <span class="jev-hint">{{ scoreNegate ? 'below level' : 'at least level' }}</span>
          <el-select
            class="jev-score-select"
            size="small"
            :model-value="selectedLevel"
            :disabled="disabled"
            data-testid="jev-score-level"
            @update:model-value="setScoreLevel"
          >
            <el-option
              v-for="(level, i) in scoreLevels"
              :key="i"
              :label="`${i}: ${level || '(unnamed)'}`"
              :value="i"
            />
          </el-select>
          <el-checkbox
            :model-value="scoreNegate"
            :disabled="disabled"
            data-testid="jev-score-negate"
            @update:model-value="setScoreNegate"
          >
            is not
          </el-checkbox>
        </template>
      </div>
    </div>

    <div class="jev-field jev-field--stack">
      <span class="jev-label">{{ confidenceLabel }} ≥ {{ confidence.toFixed(2) }}</span>
      <div class="jev-match">
        <el-select
          v-if="isNoul"
          class="jev-op-select"
          size="small"
          :model-value="noulNegate ? 'LT' : 'GTE'"
          :disabled="disabled"
          data-testid="jev-noul-operator"
          @update:model-value="setNoulOperator"
        >
          <el-option
            label="is"
            value="GTE"
          />
          <el-option
            label="is not"
            value="LT"
          />
        </el-select>
        <el-slider
          class="jev-slider"
          :model-value="confidence"
          :min="0"
          :max="1"
          :step="0.05"
          :disabled="disabled"
          data-testid="jev-confidence-slider"
          @update:model-value="setConfidence"
        />
      </div>
      <span class="jev-hint">{{ confidenceHint }}</span>
    </div>

    <el-collapse class="jev-advanced">
      <el-collapse-item
        title="Edit as JSON (advanced)"
        name="json"
      >
        <el-input
          type="textarea"
          :rows="6"
          :model-value="jsonDraft"
          :disabled="disabled"
          data-testid="jev-json"
          @update:model-value="jsonDraft = $event"
        />
        <div class="jev-json-actions">
          <el-button
            size="small"
            type="primary"
            plain
            :disabled="disabled"
            @click="applyJson"
          >
            Apply JSON
          </el-button>
          <span
            v-if="jsonError"
            class="jev-error"
          >{{ jsonError }}</span>
        </div>
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script lang="ts">
import type { PropType } from 'vue'
import type { JevQuestion, JevQuestionType } from '@/api/types'
import {
  choiceCriteriaFromRows,
  choiceOperatorFor,
  choiceOptionsFromValue,
  choiceRowsFromCriteria,
  choiceValueFromOptions,
  defaultJevQuestion,
  isNegatedOperator,
  nextChoiceName,
  noulCriteriaText,
  numberFromValue,
  scoreLevelsFromCriteria,
  withNoulCriteria,
  type ChoiceRow,
} from '@/helpers/jevQuestion'

const DEFAULT_NOUL_THRESHOLD = 0.7
const DEFAULT_CONFIDENCE = 0.5

export default {
  name: 'JevQuestionEditor',
  props: {
    modelValue: { type: Object as PropType<JevQuestion>, required: true },
    operator: { type: String, required: true },
    value: { type: String, required: true },
    disabled: { type: Boolean, default: false },
  },
  emits: ['update:modelValue', 'update:operator', 'update:value'],
  data() {
    return {
      jsonDraft: '',
      jsonError: '',
    }
  },
  computed: {
    question(): JevQuestion {
      return this.modelValue
    },
    isNoul(): boolean {
      return this.question.type === 'noul'
    },
    isChoice(): boolean {
      return this.question.type === 'choice'
    },
    instructionsText(): string {
      return typeof this.question.instructions === 'string' ? this.question.instructions : ''
    },
    criteriaLabel(): string {
      if (this.isChoice) return 'Options (name + description)'
      if (this.isNoul) return 'Boundary (optional)'
      return 'Levels (low → high)'
    },
    choiceRows(): ChoiceRow[] {
      return choiceRowsFromCriteria(this.question.criteria)
    },
    scoreLevels(): string[] {
      return scoreLevelsFromCriteria(this.question.criteria)
    },
    noulTrue(): string {
      return noulCriteriaText(this.question.criteria, 'true')
    },
    noulFalse(): string {
      return noulCriteriaText(this.question.criteria, 'false')
    },
    selectedOptions(): string[] {
      return choiceOptionsFromValue(this.value)
    },
    choiceNegate(): boolean {
      return isNegatedOperator(this.operator)
    },
    selectedLevel(): number {
      return numberFromValue(this.value) ?? 0
    },
    scoreNegate(): boolean {
      return this.operator === 'LT' || this.operator === 'LTE'
    },
    noulNegate(): boolean {
      return this.operator === 'LT' || this.operator === 'LTE'
    },
    /** noul has no model confidence; its probability threshold is the gate. */
    confidence(): number {
      if (this.isNoul) return numberFromValue(this.value) ?? DEFAULT_NOUL_THRESHOLD
      return this.question.confidenceThreshold ?? DEFAULT_CONFIDENCE
    },
    confidenceLabel(): string {
      return this.isNoul ? 'Confidence (P(yes))' : 'Answer confidence'
    },
    confidenceHint(): string {
      if (this.isNoul) {
        return 'Minimum probability that the answer is yes. Noul has no separate model confidence.'
      }
      return 'Below this confidence the constraint evaluates false and the segment falls through.'
    },
  },
  watch: {
    modelValue: {
      handler() {
        this.jsonDraft = JSON.stringify(this.modelValue, null, 2)
        this.jsonError = ''
      },
      deep: true,
      immediate: true,
    },
  },
  methods: {
    emitQuestion(patch: Partial<JevQuestion>): void {
      this.$emit('update:modelValue', { ...this.question, ...patch })
    },
    setType(type: JevQuestionType): void {
      const question = { ...defaultJevQuestion(type), instructions: this.question.instructions }
      this.$emit('update:modelValue', question)
      // Reset the match to a type-appropriate default.
      if (type === 'choice') {
        this.$emit('update:operator', 'EQ')
        this.$emit('update:value', '')
      } else if (type === 'score') {
        this.$emit('update:operator', 'GTE')
        this.$emit('update:value', '0')
      } else {
        this.$emit('update:operator', 'GTE')
        this.$emit('update:value', DEFAULT_NOUL_THRESHOLD.toFixed(2))
      }
    },
    setInstructions(value: string): void {
      this.emitQuestion({ instructions: value })
    },
    setNoul(key: 'true' | 'false', value: string): void {
      this.emitQuestion({ criteria: withNoulCriteria(this.question.criteria, key, value) })
    },
    setChoiceCriteria(rows: ChoiceRow[]): void {
      this.emitQuestion({ criteria: choiceCriteriaFromRows(rows) })
    },
    setChoiceName(index: number, name: string): void {
      const rows = [...this.choiceRows]
      rows[index] = { ...rows[index], name }
      this.setChoiceCriteria(rows)
    },
    setChoiceDescription(index: number, description: string): void {
      const rows = [...this.choiceRows]
      rows[index] = { ...rows[index], description }
      this.setChoiceCriteria(rows)
    },
    addChoice(): void {
      const rows = [...this.choiceRows]
      rows.push({ name: nextChoiceName(rows), description: '' })
      this.setChoiceCriteria(rows)
    },
    removeChoice(index: number): void {
      this.setChoiceCriteria(this.choiceRows.filter((_, i) => i !== index))
    },
    setScoreLevels(levels: string[]): void {
      this.emitQuestion({ criteria: levels })
    },
    setScoreLevelText(index: number, value: string): void {
      const levels = [...this.scoreLevels]
      levels[index] = value
      this.setScoreLevels(levels)
    },
    addScoreLevel(): void {
      this.setScoreLevels([...this.scoreLevels, ''])
    },
    removeScoreLevel(index: number): void {
      this.setScoreLevels(this.scoreLevels.filter((_, i) => i !== index))
    },
    moveScoreLevel(index: number, delta: number): void {
      const levels = [...this.scoreLevels]
      const target = index + delta
      if (target < 0 || target >= levels.length) return
      const [moved] = levels.splice(index, 1)
      levels.splice(target, 0, moved)
      this.setScoreLevels(levels)
    },
    setChoiceOptions(options: string[]): void {
      this.$emit('update:value', choiceValueFromOptions(options))
      this.$emit('update:operator', choiceOperatorFor(options, this.choiceNegate))
    },
    setChoiceNegate(negate: boolean): void {
      this.$emit('update:operator', choiceOperatorFor(this.selectedOptions, negate))
    },
    setScoreLevel(level: number): void {
      this.$emit('update:value', String(level))
      this.$emit('update:operator', this.scoreNegate ? 'LT' : 'GTE')
    },
    setScoreNegate(negate: boolean): void {
      this.$emit('update:operator', negate ? 'LT' : 'GTE')
    },
    setNoulOperator(operator: string): void {
      this.$emit('update:operator', operator)
    },
    setConfidence(value: number): void {
      if (this.isNoul) {
        this.$emit('update:value', value.toFixed(2))
        return
      }
      this.emitQuestion({ confidenceThreshold: value })
    },
    applyJson(): void {
      try {
        const parsed = JSON.parse(this.jsonDraft) as JevQuestion
        if (!parsed || typeof parsed !== 'object' || !parsed.type) {
          throw new Error('question must be an object with a type')
        }
        this.jsonError = ''
        this.$emit('update:modelValue', parsed)
      } catch (err) {
        this.jsonError = err instanceof Error ? err.message : 'invalid JSON'
      }
    },
  },
}
</script>

<style scoped>
.jev-editor {
  display: flex;
  flex-direction: column;
  gap: var(--space-3xs);
  padding: var(--space-3xs) 0;
  width: 100%;
}
.jev-field {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
  flex-wrap: wrap;
}
.jev-field--stack {
  flex-direction: column;
  align-items: stretch;
  gap: var(--space-3xs);
}
.jev-label {
  font-size: var(--font-size-body-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}
.jev-type {
  width: 190px;
  flex: 0 0 auto;
}
.jev-criteria {
  display: flex;
  flex-direction: column;
  gap: var(--space-3xs);
}
.jev-option-row {
  display: flex;
  align-items: center;
  gap: var(--space-3xs);
}
.jev-level-index {
  width: 1.5em;
  text-align: right;
  color: var(--el-text-color-secondary);
  font-size: var(--font-size-body-sm);
}
.jev-match {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
  flex-wrap: wrap;
}
.jev-choice-select {
  flex: 1;
  min-width: 180px;
}
.jev-score-select,
.jev-op-select {
  width: 180px;
  flex: 0 0 auto;
}
.jev-slider {
  flex: 1;
  min-width: 140px;
}
.jev-hint {
  font-size: var(--font-size-body-sm);
  color: var(--el-text-color-secondary);
}
.jev-json-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
  margin-top: var(--space-3xs);
}
.jev-error {
  color: var(--el-color-danger);
  font-size: var(--font-size-body-sm);
}
</style>
