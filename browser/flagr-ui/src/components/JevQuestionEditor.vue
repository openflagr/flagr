<template>
  <div
    class="jev-editor"
    data-testid="jev-question-editor"
  >
    <div class="jev-notice">
      <el-icon class="jev-notice__icon">
        <InfoFilled />
      </el-icon>
      <span>
        Jev constraints call a System One model on every evaluation, which makes
        them slower than normal constraints and adds model usage cost.
      </span>
    </div>

    <div class="jev-row">
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
          label="Noul (true / false)"
          value="noul"
        />
        <el-option
          label="Choice (pick one)"
          value="choice"
        />
        <el-option
          label="Scale (rate)"
          value="score"
        />
      </el-select>
    </div>

    <div class="jev-stack">
      <span class="jev-label">Instructions</span>
      <el-input
        type="textarea"
        :autosize="{ minRows: 1, maxRows: 4 }"
        size="small"
        placeholder="e.g. Is `message` about billing?"
        :model-value="instructionsText"
        :disabled="disabled"
        data-testid="jev-instructions"
        @update:model-value="setInstructions"
      />
      <span class="jev-hint">
        Name state fields with backticks, e.g. <code>`message`</code> or
        <code>`account.plan`</code>. The state also has <code>`entityID`</code>
        and <code>`entityType`</code>.
      </span>
    </div>

    <div class="jev-stack">
      <span class="jev-label">{{ criteriaLabel }}</span>

      <div
        v-if="isNoul"
        class="jev-criteria"
      >
        <div class="jev-noul-row">
          <span class="jev-noul-label">true</span>
          <el-input
            size="small"
            placeholder="what counts as true"
            :model-value="noulTrue"
            :disabled="disabled"
            @update:model-value="setNoul('true', $event)"
          />
        </div>
        <div class="jev-noul-row">
          <span class="jev-noul-label">false</span>
          <el-input
            size="small"
            placeholder="what counts as false"
            :model-value="noulFalse"
            :disabled="disabled"
            @update:model-value="setNoul('false', $event)"
          />
        </div>
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
            class="jev-option-name"
            placeholder="choice"
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
          Add choice
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
            placeholder="level (low → high)"
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
          @click="addLevel"
        >
          Add level
        </el-button>
      </div>
    </div>

    <div class="jev-row jev-row--match">
      <span class="jev-label">Match</span>

      <template v-if="isNoul">
        <span class="jev-hint">when P(true) is</span>
        <el-select
          class="jev-op-select"
          size="small"
          :model-value="noulOperator"
          :disabled="disabled"
          data-testid="jev-noul-operator"
          @update:model-value="setNoulOperator"
        >
          <el-option
            label="≥"
            value="GTE"
          />
          <el-option
            label="<"
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
        <span class="jev-value">{{ confidence.toFixed(2) }}</span>
      </template>

      <template v-else-if="isChoice">
        <span class="jev-hint">when the model picks</span>
        <el-select
          class="jev-op-select"
          size="small"
          :model-value="choiceDirection"
          :disabled="disabled"
          data-testid="jev-choice-direction"
          @update:model-value="setChoiceDirection"
        >
          <el-option
            label="any of"
            value="include"
          />
          <el-option
            label="none of"
            value="exclude"
          />
        </el-select>
        <el-select
          class="jev-choice-select"
          size="small"
          multiple
          collapse-tags
          collapse-tags-tooltip
          placeholder="Pick choice(s)"
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
        <span class="jev-label jev-label--inline">confidence</span>
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
        <span class="jev-value">{{ confidence.toFixed(2) }}</span>
      </template>

      <template v-else>
        <span class="jev-hint">when the score is</span>
        <el-select
          class="jev-op-select"
          size="small"
          :model-value="scoreDirection"
          :disabled="disabled"
          data-testid="jev-score-direction"
          @update:model-value="setScoreDirection"
        >
          <el-option
            label="≥"
            value="atleast"
          />
          <el-option
            label="<"
            value="below"
          />
        </el-select>
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
            :label="`level ${i} · ${level || 'level'}`"
            :value="i"
          />
        </el-select>
        <span class="jev-label jev-label--inline">confidence</span>
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
        <span class="jev-value">{{ confidence.toFixed(2) }}</span>
      </template>
    </div>

    <span class="jev-hint">{{ matchHint }}</span>

    <el-collapse class="jev-advanced">
      <el-collapse-item
        title="Edit as JSON (advanced)"
        name="json"
      >
        <el-input
          type="textarea"
          :rows="5"
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
import { InfoFilled } from '@element-plus/icons-vue'
import {
  choiceOptionsFromValue,
  choiceRowsFromCriteria,
  isNegatedOperator,
  nextChoiceName,
  noulCriteriaText,
  numberFromValue,
  reduceJevMatch,
  scoreLevelsFromCriteria,
  type ChoiceRow,
  type JevAction,
  type JevMatchState,
} from '@/helpers/jevQuestion'

const DEFAULT_NOUL_THRESHOLD = 0.7
const DEFAULT_CONFIDENCE = 0.5

export default {
  name: 'JevQuestionEditor',
  components: { InfoFilled },
  props: {
    modelValue: { type: Object as PropType<JevQuestion>, required: true },
    operator: { type: String, required: true },
    value: { type: String, required: true },
    disabled: { type: Boolean, default: false },
  },
  // One atomic event: emitting operator and value separately clobbers one
  // another because the parent's draft prop has not re-rendered yet.
  emits: ['update:all'],
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
      if (this.isChoice) return 'Choices'
      if (this.isNoul) return 'Criteria (true / false)'
      return 'Scale (low → high)'
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
    noulOperator(): string {
      return this.noulNegate ? 'LT' : 'GTE'
    },
    choiceDirection(): string {
      return this.choiceNegate ? 'exclude' : 'include'
    },
    scoreDirection(): string {
      return this.scoreNegate ? 'below' : 'atleast'
    },
    /** noul has no model confidence; its probability threshold is the gate. */
    confidence(): number {
      if (this.isNoul) return numberFromValue(this.value) ?? DEFAULT_NOUL_THRESHOLD
      return this.question.confidenceThreshold ?? DEFAULT_CONFIDENCE
    },
    matchHint(): string {
      if (this.isNoul) {
        return 'Noul returns the probability that the answer is true; the slider is that probability threshold.'
      }
      if (this.isChoice) {
        return this.choiceNegate
          ? 'Matches when the model picks none of these choices, with at least this confidence.'
          : 'Matches when the model picks one of these choices, with at least this confidence.'
      }
      return 'Matches when the model score is at least this level with at least this confidence.'
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
    matchState(): JevMatchState {
      return { jev: this.question, operator: this.operator, value: this.value }
    },
    dispatch(action: JevAction): void {
      this.$emit('update:all', reduceJevMatch(this.matchState(), action))
    },
    setType(questionType: JevQuestionType): void {
      this.dispatch({ type: 'setType', questionType })
    },
    setInstructions(instructions: string): void {
      this.dispatch({ type: 'setInstructions', instructions })
    },
    setNoul(key: 'true' | 'false', value: string): void {
      this.dispatch({ type: 'setNoulCriteria', key, value })
    },
    setChoiceCriteria(rows: ChoiceRow[]): void {
      this.dispatch({ type: 'setChoiceCriteria', rows })
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
      this.dispatch({ type: 'setScoreLevels', levels })
    },
    setScoreLevelText(index: number, value: string): void {
      const levels = [...this.scoreLevels]
      levels[index] = value
      this.setScoreLevels(levels)
    },
    addLevel(): void {
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
      this.dispatch({ type: 'setChoiceOptions', options })
    },
    setChoiceNegate(negate: boolean): void {
      this.dispatch({ type: 'setChoiceNegate', negate })
    },
    setChoiceDirection(direction: string): void {
      this.setChoiceNegate(direction === 'exclude')
    },
    setScoreLevel(level: number): void {
      this.dispatch({ type: 'setScoreLevel', level })
    },
    setScoreNegate(negate: boolean): void {
      this.dispatch({ type: 'setScoreNegate', negate })
    },
    setScoreDirection(direction: string): void {
      this.setScoreNegate(direction === 'below')
    },
    setNoulOperator(operator: string): void {
      this.dispatch({ type: 'setNoulOperator', operator })
    },
    setConfidence(value: number): void {
      this.dispatch({ type: 'setConfidence', value })
    },
    applyJson(): void {
      try {
        const parsed = JSON.parse(this.jsonDraft) as JevQuestion
        if (!parsed || typeof parsed !== 'object' || !parsed.type) {
          throw new Error('question must be an object with a type')
        }
        this.jsonError = ''
        this.dispatch({ type: 'setJev', jev: parsed })
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
  gap: var(--space-2xs);
  width: 100%;
}
.jev-notice {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3xs);
  padding: var(--space-3xs) var(--space-2xs);
  background: var(--el-color-warning-light-9);
  border: 1px solid var(--el-color-warning-light-7);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-caption);
  color: var(--el-text-color-regular);
  line-height: 1.5;
}
.jev-notice__icon {
  color: var(--el-color-warning);
  margin-top: 0.15em;
  flex: 0 0 auto;
}
.jev-row {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
  flex-wrap: wrap;
}
.jev-row--match {
  align-items: center;
}
.jev-stack {
  display: flex;
  flex-direction: column;
  gap: var(--space-3xs);
}
.jev-label {
  font-size: var(--font-size-body-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}
.jev-label--inline {
  margin-left: var(--space-2xs);
}
.jev-type {
  width: 170px;
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
.jev-option-name {
  max-width: 180px;
}
.jev-level-index {
  width: 1.4em;
  text-align: right;
  color: var(--el-text-color-secondary);
  font-size: var(--font-size-body-sm);
}
.jev-noul-row {
  display: flex;
  align-items: center;
  gap: var(--space-3xs);
}
.jev-noul-label {
  width: 2.2em;
  font-size: var(--font-size-body-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--el-text-color-secondary);
}
.jev-choice-select {
  flex: 1;
  min-width: 160px;
}
.jev-score-select,
.jev-op-select {
  width: 120px;
  flex: 0 0 auto;
}
.jev-slider {
  flex: 1;
  min-width: 120px;
}
.jev-value {
  font-family: var(--font-mono);
  font-size: var(--font-size-body-sm);
  color: var(--el-text-color-regular);
  min-width: 2.6em;
  text-align: right;
}
.jev-hint {
  font-size: var(--font-size-caption);
  color: var(--el-text-color-placeholder);
}
.jev-hint code {
  font-family: var(--font-mono);
  background: var(--el-fill-color-light);
  border-radius: var(--radius-sm);
  padding: 0 0.15em;
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
.jev-advanced {
  border-top: none;
  border-bottom: none;

  :deep(.el-collapse-item__header) {
    height: auto;
    line-height: 1.5;
    padding: var(--space-3xs) 0;
    background: transparent;
    border-bottom: none;
    font-size: var(--font-size-caption);
    font-weight: var(--font-weight-normal);
    color: var(--el-text-color-placeholder);
  }

  :deep(.el-collapse-item__wrap) {
    background: transparent;
    border-bottom: none;
  }

  :deep(.el-collapse-item__content) {
    padding-bottom: var(--space-3xs);
  }
}
</style>
