<template>
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
</template>

<script lang="ts">
import type { PropType } from 'vue'
import type { JevQuestion } from '@/api/types'
import {
  choiceRowsFromCriteria,
  nextChoiceName,
  noulCriteriaText,
  scoreLevelsFromCriteria,
  type ChoiceRow,
  type JevAction,
} from '@/helpers/jevQuestion'

export default {
  name: 'JevCriteriaEditor',
  props: {
    jev: { type: Object as PropType<JevQuestion>, required: true },
    disabled: { type: Boolean, default: false },
  },
  emits: ['action'],
  computed: {
    isNoul(): boolean {
      return this.jev.type === 'noul'
    },
    isChoice(): boolean {
      return this.jev.type === 'choice'
    },
    criteriaLabel(): string {
      if (this.isChoice) return 'Choices'
      if (this.isNoul) return 'Criteria (true / false)'
      return 'Scale (low → high)'
    },
    choiceRows(): ChoiceRow[] {
      return choiceRowsFromCriteria(this.jev.criteria)
    },
    scoreLevels(): string[] {
      return scoreLevelsFromCriteria(this.jev.criteria)
    },
    noulTrue(): string {
      return noulCriteriaText(this.jev.criteria, 'true')
    },
    noulFalse(): string {
      return noulCriteriaText(this.jev.criteria, 'false')
    },
  },
  methods: {
    emitAction(action: JevAction): void {
      this.$emit('action', action)
    },
    setNoul(key: 'true' | 'false', value: string): void {
      this.emitAction({ type: 'setNoulCriteria', key, value })
    },
    setChoiceCriteria(rows: ChoiceRow[]): void {
      this.emitAction({ type: 'setChoiceCriteria', rows })
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
      this.emitAction({ type: 'setScoreLevels', levels })
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
  },
}
</script>

<style scoped>
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
</style>
