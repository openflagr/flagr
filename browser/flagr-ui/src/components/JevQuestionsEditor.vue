<template>
  <div class="jev-editor">
    <div
      v-for="(row, index) in rows"
      :key="index"
      class="jev-question"
      :data-testid="`jev-question-${index}`"
    >
      <div class="jev-question-head">
        <el-input
          v-model="row.name"
          :formatter="formatJevName"
          :parser="formatJevName"
          size="small"
          class="jev-name"
          placeholder="question_name"
          data-testid="jev-question-name"
          @update:model-value="sync()"
        />
        <el-select
          :model-value="row.type"
          size="small"
          class="jev-type"
          data-testid="jev-question-type"
          @update:model-value="setType(index, $event)"
        >
          <el-option
            label="True / False"
            value="noul"
          />
          <el-option
            label="Choice"
            value="choice"
          />
          <el-option
            label="Scale"
            value="score"
          />
        </el-select>
        <el-button
          size="small"
          link
          type="danger"
          class="jev-remove"
          :disabled="disabled || rows.length <= 1"
          :data-testid="`jev-remove-question-${index}`"
          @click="removeQuestion(index)"
        >
          Remove
        </el-button>
      </div>

      <div class="jev-question-meta">
        <code
          class="jev-prop"
          data-testid="jev-property-preview"
        >{{ propertyFor(row) }}</code>
        <span class="jev-value-hint">{{ valueHint(row.type) }}</span>
      </div>

      <el-input
        v-model="row.instructions"
        type="textarea"
        :autosize="{ minRows: 1, maxRows: 3 }"
        placeholder="Instructions for the model (keep it atomic)"
        data-testid="jev-question-instructions"
        @update:model-value="sync()"
      />

      <div
        v-if="row.type === 'noul'"
        class="jev-criteria"
      >
        <div class="jev-criteria-row jev-criteria-row--noul">
          <span class="jev-criteria-label">True means</span>
          <el-input
            v-model="row.trueText"
            size="small"
            placeholder="optional — e.g. about billing"
            data-testid="jev-noul-true"
            @update:model-value="sync()"
          />
        </div>
        <div class="jev-criteria-row jev-criteria-row--noul">
          <span class="jev-criteria-label">False means</span>
          <el-input
            v-model="row.falseText"
            size="small"
            placeholder="optional — e.g. unrelated to billing"
            data-testid="jev-noul-false"
            @update:model-value="sync()"
          />
        </div>
      </div>

      <div
        v-else-if="row.type === 'choice'"
        class="jev-criteria"
      >
        <div
          v-for="(choice, choiceIndex) in row.choices"
          :key="choiceIndex"
          class="jev-criteria-row jev-criteria-row--choice"
        >
          <el-input
            v-model="choice.name"
            size="small"
            placeholder="option"
            data-testid="jev-choice-name"
            @update:model-value="sync()"
          />
          <el-input
            v-model="choice.description"
            size="small"
            placeholder="description"
            data-testid="jev-choice-description"
            @update:model-value="sync()"
          />
          <el-button
            size="small"
            link
            type="danger"
            :disabled="disabled"
            @click="removeChoice(index, choiceIndex)"
          >
            x
          </el-button>
        </div>
        <el-button
          size="small"
          link
          type="primary"
          :disabled="disabled"
          @click="addChoice(index)"
        >
          + option
        </el-button>
      </div>

      <div
        v-else-if="row.type === 'score'"
        class="jev-criteria"
      >
        <div
          v-for="(level, levelIndex) in row.levels"
          :key="levelIndex"
          class="jev-criteria-row jev-criteria-row--score"
        >
          <el-input
            v-model="row.levels[levelIndex]"
            size="small"
            placeholder="level (low to high)"
            data-testid="jev-score-level"
            @update:model-value="sync()"
          />
          <el-button
            size="small"
            link
            type="danger"
            :disabled="disabled"
            @click="removeLevel(index, levelIndex)"
          >
            x
          </el-button>
        </div>
        <el-button
          size="small"
          link
          type="primary"
          :disabled="disabled"
          @click="addLevel(index)"
        >
          + level
        </el-button>
      </div>

      <div
        v-if="row.type !== 'noul'"
        class="jev-confidence"
      >
        <span class="jev-confidence-label">Confidence ≥</span>
        <el-input-number
          :model-value="row.threshold"
          :min="0"
          :max="1"
          :step="0.05"
          :precision="2"
          size="small"
          controls-position="right"
          :disabled="disabled"
          data-testid="jev-question-confidence"
          @update:model-value="setThreshold(index, $event)"
        />
        <span class="jev-confidence-hint">below this the answer is dropped</span>
      </div>
    </div>

    <el-button
      size="small"
      link
      type="primary"
      class="jev-add-question"
      data-testid="jev-add-question-btn"
      :disabled="disabled"
      @click="addQuestion"
    >
      + Question
    </el-button>
  </div>
</template>

<script lang="ts">
import type { PropType } from 'vue'
import type { JevQuestion } from '@/api/types'
import {
  type ChoiceRow,
  choiceCriteriaFromRows,
  choiceRowsFromCriteria,
  defaultJevQuestion,
  jevPropertyFor,
  nextChoiceName,
  noulCriteriaText,
  slugifyJevName,
  withNoulCriteria,
} from '@/helpers/jevQuestion'

type JevQuestionType = 'noul' | 'choice' | 'score'

interface EditRow {
  name: string
  type: JevQuestionType
  instructions: string
  threshold?: number
  trueText: string
  falseText: string
  choices: ChoiceRow[]
  levels: string[]
}

function instructionsText(instructions: unknown): string {
  return typeof instructions === 'string' ? instructions : ''
}

function toRows(questions: Record<string, JevQuestion>): EditRow[] {
  return Object.entries(questions).map(([name, question]) => ({
    name,
    type: (question.type as JevQuestionType) ?? 'noul',
    instructions: instructionsText(question.instructions),
    threshold: question.confidenceThreshold,
    trueText: noulCriteriaText(question.criteria, 'true'),
    falseText: noulCriteriaText(question.criteria, 'false'),
    choices: choiceRowsFromCriteria(question.criteria),
    levels: Array.isArray(question.criteria) ? question.criteria.map((level) => String(level)) : [],
  }))
}

function toQuestions(rows: EditRow[]): Record<string, JevQuestion> {
  const out: Record<string, JevQuestion> = {}
  for (const row of rows) {
    const name = row.name.trim()
    if (!name) continue
    const question: JevQuestion = {
      type: row.type,
      instructions: row.instructions,
    }
    if (row.type === 'noul') {
      // `criteria` is optional: describe what true and false mean so the model
      // resolves the boundary the way you intend.
      const criteria = withNoulCriteria(
        withNoulCriteria(undefined, 'true', row.trueText),
        'false',
        row.falseText,
      )
      if (criteria) question.criteria = criteria
    }
    if (row.type === 'choice') question.criteria = choiceCriteriaFromRows(row.choices)
    if (row.type === 'score') question.criteria = row.levels.filter((level) => level.trim() !== '')
    if (row.threshold !== undefined && row.threshold !== null) question.confidenceThreshold = row.threshold
    out[name] = question
  }
  return out
}

/**
 * Editor for a `jev` enricher's questions. The rows are the source of truth
 * once mounted — the parent re-mounts this component (via `:key`) when it swaps
 * in a different enricher, so there is no prop watcher to fight with typing.
 */
export default {
  name: 'JevQuestionsEditor',
  props: {
    modelValue: { type: Object as PropType<Record<string, JevQuestion>>, default: () => ({}) },
    disabled: { type: Boolean, default: false },
  },
  emits: ['update:modelValue'],
  data() {
    return {
      rows: toRows(this.modelValue) as EditRow[],
    }
  },
  methods: {
    /** The constraint property this question will populate, e.g. `@jev_plan_tier`. */
    propertyFor(row: EditRow): string {
      return row.name.trim() ? jevPropertyFor(row.name) : '@jev_<name>'
    },
    /**
     * What the answer becomes in the evaluation context, and how to match it.
     * Kept in lockstep with jevAnswerValue in pkg/handler/jev_client.go.
     */
    valueHint(type: JevQuestionType): string {
      switch (type) {
        case 'choice':
          return 'value is the option label · match = / ≠ / in / not in'
        case 'score':
          return 'value is the level number · match ≥ / > / ≤ / <'
        default:
          return 'value is P(true), 0–1 · match ≥ / > / ≤ / <'
      }
    },
    /**
     * Keep the name in the canonical `@jev_<name>` form as the user types. Used
     * as both formatter and parser so the native input is re-synced even when a
     * keystroke does not change the model (e.g. a second stray separator).
     */
    formatJevName(value: string | number): string {
      return slugifyJevName(String(value))
    },
    sync() {
      this.$emit('update:modelValue', toQuestions(this.rows))
    },
    setType(index: number, type: JevQuestionType) {
      const fresh = defaultJevQuestion(type)
      this.rows[index] = {
        name: this.rows[index].name,
        type,
        instructions: this.rows[index].instructions || instructionsText(fresh.instructions),
        threshold: this.rows[index].threshold,
        trueText: '',
        falseText: '',
        choices: choiceRowsFromCriteria(fresh.criteria),
        levels: Array.isArray(fresh.criteria) ? fresh.criteria.map((level) => String(level)) : [],
      }
      this.sync()
    },
    addQuestion() {
      const fresh = defaultJevQuestion('noul')
      this.rows.push({
        name: `question_${this.rows.length + 1}`,
        type: 'noul',
        instructions: instructionsText(fresh.instructions),
        trueText: '',
        falseText: '',
        choices: [],
        levels: [],
      })
      this.sync()
    },
    removeQuestion(index: number) {
      this.rows.splice(index, 1)
      this.sync()
    },
    addChoice(index: number) {
      const row = this.rows[index]
      row.choices.push({ name: nextChoiceName(row.choices), description: '' })
      this.sync()
    },
    removeChoice(index: number, choiceIndex: number) {
      this.rows[index].choices.splice(choiceIndex, 1)
      this.sync()
    },
    addLevel(index: number) {
      this.rows[index].levels.push('')
      this.sync()
    },
    removeLevel(index: number, levelIndex: number) {
      this.rows[index].levels.splice(levelIndex, 1)
      this.sync()
    },
    setThreshold(index: number, value: number | undefined) {
      this.rows[index].threshold = value ?? undefined
      this.sync()
    },
  },
}
</script>

<style scoped>
.jev-editor {
  --jev-name-width: 220px;
  --jev-type-width: 118px;
  --jev-criteria-label-width: 74px;
  --jev-noul-row-width: 520px;
  --jev-choice-row-width: 560px;
  --jev-score-row-width: 320px;

  display: flex;
  flex-direction: column;
  gap: var(--space-2xs);
}

.jev-question {
  display: flex;
  flex-direction: column;
  gap: var(--space-3xs);
  padding: var(--space-2xs) var(--space-xs);
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-light);
  border-radius: var(--radius-md);
}

.jev-question-head {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
}

.jev-name {
  flex: 0 1 var(--jev-name-width);
}

.jev-type {
  flex: 0 0 var(--jev-type-width);
}

.jev-remove {
  margin-left: auto;
}

.jev-question-meta {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: var(--space-3xs) var(--space-2xs);
  font-size: var(--font-size-caption);
  line-height: var(--line-height-tight);
}

.jev-prop {
  font-family: var(--font-mono);
  color: var(--el-text-color-secondary);
}

.jev-value-hint {
  color: var(--el-text-color-placeholder);
}

.jev-criteria {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-3xs);
}

.jev-criteria-row {
  display: flex;
  align-items: center;
  gap: var(--space-3xs);
  width: 100%;
}

.jev-criteria-row--noul {
  max-width: var(--jev-noul-row-width);
}

.jev-criteria-row--choice {
  max-width: var(--jev-choice-row-width);
}

.jev-criteria-row--score {
  max-width: var(--jev-score-row-width);
}

.jev-criteria-row .el-input {
  flex: 1;
}

.jev-criteria-label {
  flex: 0 0 var(--jev-criteria-label-width);
  color: var(--el-text-color-secondary);
  font-size: var(--font-size-caption);
}

.jev-confidence {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
}

.jev-confidence-label,
.jev-confidence-hint {
  color: var(--el-text-color-secondary);
  font-size: var(--font-size-caption);
}

.jev-add-question {
  align-self: flex-start;
}
</style>
