<template>
  <div class="jev-editor">
    <div
      v-for="(row, index) in rows"
      :key="index"
      class="jev-question ui-surface-inset"
    >
      <div class="jev-question-head">
        <el-input
          v-model="row.name"
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
            label="Yes / No"
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
          :disabled="disabled || rows.length <= 1"
          @click="removeQuestion(index)"
        >
          Remove
        </el-button>
      </div>

      <el-input
        v-model="row.instructions"
        type="textarea"
        :rows="2"
        placeholder="Instructions for the model (keep it atomic)"
        data-testid="jev-question-instructions"
        @update:model-value="sync()"
      />

      <div
        v-if="row.type === 'choice'"
        class="jev-criteria"
      >
        <div
          v-for="(choice, choiceIndex) in row.choices"
          :key="choiceIndex"
          class="jev-criteria-row"
        >
          <el-input
            v-model="choice.name"
            size="small"
            placeholder="option"
            @update:model-value="sync()"
          />
          <el-input
            v-model="choice.description"
            size="small"
            placeholder="description"
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
          class="jev-criteria-row"
        >
          <el-input
            v-model="row.levels[levelIndex]"
            size="small"
            placeholder="level (low to high)"
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

      <div class="jev-confidence">
        <span class="jev-confidence-label">Confidence ≥</span>
        <el-input-number
          :model-value="row.threshold"
          :min="0"
          :max="1"
          :step="0.05"
          :precision="2"
          size="small"
          :disabled="disabled"
          data-testid="jev-question-confidence"
          @update:model-value="setThreshold(index, $event)"
        />
        <span class="jev-confidence-hint">below this, the answer is dropped</span>
      </div>
    </div>

    <el-button
      size="small"
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
  nextChoiceName,
} from '@/helpers/jevQuestion'

type JevQuestionType = 'noul' | 'choice' | 'score'

interface EditRow {
  name: string
  type: JevQuestionType
  instructions: string
  threshold?: number
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
    if (row.type === 'choice') question.criteria = choiceCriteriaFromRows(row.choices)
    if (row.type === 'score') question.criteria = row.levels.filter((level) => level.trim() !== '')
    if (row.threshold !== undefined && row.threshold !== null) question.confidenceThreshold = row.threshold
    out[name] = question
  }
  return out
}

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
  watch: {
    modelValue: {
      handler(value: Record<string, JevQuestion>) {
        this.rows = toRows(value ?? {})
      },
      deep: true,
    },
  },
  methods: {
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
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.jev-question {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.jev-question-head {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.jev-name {
  flex: 1;
}

.jev-type {
  width: 120px;
}

.jev-criteria {
  display: flex;
  flex-direction: column;
  gap: var(--space-3xs);
}

.jev-criteria-row {
  display: flex;
  align-items: center;
  gap: var(--space-3xs);
}

.jev-confidence {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.jev-confidence-label {
  font-size: var(--font-size-caption);
}

.jev-confidence-hint {
  color: var(--el-text-color-secondary);
  font-size: var(--font-size-micro);
}
</style>
