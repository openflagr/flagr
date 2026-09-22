<template>
  <div
    class="jev-editor"
    data-testid="jev-question-editor"
  >
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

      <template v-if="question.type !== 'noul'">
        <span class="jev-label">Confidence ≥ {{ confidence.toFixed(2) }}</span>
        <el-slider
          class="jev-confidence"
          :model-value="confidence"
          :min="0"
          :max="1"
          :step="0.05"
          :disabled="disabled"
          data-testid="jev-confidence-slider"
          @update:model-value="setConfidence"
        />
      </template>
      <span
        v-else
        class="jev-hint"
      >Noul has no separate confidence — the probability comparison is the gate.</span>
    </div>

    <el-input
      type="textarea"
      :rows="2"
      size="small"
      placeholder="Question the model answers (instructions)"
      :model-value="instructionsText"
      :disabled="disabled"
      data-testid="jev-instructions"
      @update:model-value="setInstructions"
    />

    <div
      v-if="question.type === 'noul'"
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
      v-else-if="question.type === 'choice'"
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
          @update:model-value="setScoreLevel(i, $event)"
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
  choiceRowsFromCriteria,
  defaultJevQuestion,
  nextChoiceName,
  noulCriteriaText,
  scoreLevelsFromCriteria,
  withNoulCriteria,
  type ChoiceRow,
} from '@/helpers/jevQuestion'

export default {
  name: 'JevQuestionEditor',
  props: {
    modelValue: { type: Object as PropType<JevQuestion>, required: true },
    disabled: { type: Boolean, default: false },
  },
  emits: ['update:modelValue'],
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
    confidence(): number {
      return this.question.confidenceThreshold ?? 0.5
    },
    instructionsText(): string {
      return typeof this.question.instructions === 'string' ? this.question.instructions : ''
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
      this.emitQuestion({ ...defaultJevQuestion(type), instructions: this.question.instructions })
    },
    setConfidence(value: number): void {
      this.emitQuestion({ confidenceThreshold: value })
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
    setScoreLevel(index: number, value: string): void {
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
.jev-row {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
  flex-wrap: wrap;
}
.jev-label {
  font-size: var(--font-size-body-sm);
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}
.jev-type {
  width: 190px;
  flex: 0 0 auto;
}
.jev-confidence {
  flex: 1;
  min-width: 120px;
}
.jev-hint {
  font-size: var(--font-size-body-sm);
  color: var(--el-text-color-secondary);
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
