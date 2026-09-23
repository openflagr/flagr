<template>
  <el-card
    class="enrichers-card"
    shadow="never"
  >
    <template #header>
      <div class="enrichers-header">
        <span class="ui-section-title">Context enrichers</span>
        <el-button
          v-if="!readonly && !hasFlagEnricher"
          size="small"
          type="primary"
          plain
          data-testid="add-jev-enricher-btn"
          @click="addJev"
        >
          + Jev
        </el-button>
      </div>
      <p class="enrichers-hint">
        Properties injected into the evaluation context before constraints run. Pick them when adding a constraint.
      </p>
    </template>

    <div
      v-if="!enrichers.length"
      class="card--empty"
    >
      No enrichers enabled. Enable built-ins with <code>FLAGR_INJECTED_CONTEXT_ENABLED</code>, or add a Jev enricher.
    </div>

    <div
      v-else
      class="enricher-list"
    >
      <div
        v-for="enricher in enrichers"
        :key="enricher.namespace"
        class="enricher-row"
        :class="{ 'enricher-row--flag': enricher.scope === 'flag' }"
        :data-testid="`enricher-row-${enricher.namespace}`"
      >
        <div class="enricher-line">
          <code class="enricher-ns">{{ enricher.namespace }}</code>
          <el-tag
            size="small"
            effect="plain"
            :type="enricher.scope === 'global' ? 'info' : 'primary'"
          >
            {{ enricher.scope === 'global' ? 'built-in' : 'flag' }}
          </el-tag>
          <el-tag
            v-if="enricher.enabled === false"
            size="small"
            type="warning"
            effect="plain"
            data-testid="enricher-disabled-tag"
            title="This namespace is not enabled on this server; its constraints fail closed."
          >
            disabled
          </el-tag>
          <span class="enricher-props">
            <el-tag
              v-for="property in enricher.properties"
              :key="property"
              size="small"
              effect="plain"
              type="info"
            >
              {{ property }}
            </el-tag>
            <span
              v-if="!(enricher.properties ?? []).length"
              class="enricher-no-props"
            >no properties</span>
          </span>
          <el-button
            v-if="!readonly && enricher.scope === 'flag'"
            size="small"
            link
            type="danger"
            class="enricher-delete"
            :data-testid="`delete-enricher-${enricher.namespace}`"
            @click="$emit('delete-enricher', enricher.namespace)"
          >
            Delete
          </el-button>
        </div>

        <div
          v-if="!readonly && enricher.scope === 'flag'"
          class="enricher-edit ui-surface-inset"
        >
          <JevQuestionsEditor
            v-if="enricher.namespace === 'jev'"
            v-model="questions[enricher.namespace]"
          />
          <div class="enricher-actions">
            <span
              v-if="problemsFor(enricher.namespace).length"
              class="enricher-problem"
            >
              {{ problemsFor(enricher.namespace)[0] }}
            </span>
            <el-button
              size="small"
              type="primary"
              :disabled="problemsFor(enricher.namespace).length > 0"
              :data-testid="`save-enricher-${enricher.namespace}`"
              @click="save(enricher.namespace)"
            >
              Save
            </el-button>
          </div>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script lang="ts">
import type { PropType } from 'vue'
import type { Enricher, JevQuestion } from '@/api/types'
import JevQuestionsEditor from '@/components/JevQuestionsEditor.vue'
import { defaultJevQuestion, jevQuestionProblems } from '@/helpers/jevQuestion'

export default {
  name: 'ContextEnrichersCard',
  components: { JevQuestionsEditor },
  props: {
    enrichers: { type: Array as PropType<Enricher[]>, default: () => [] },
    readonly: { type: Boolean, default: false },
  },
  emits: ['create-enricher', 'save-enricher', 'delete-enricher'],
  data() {
    return {
      questions: {} as Record<string, Record<string, JevQuestion>>,
    }
  },
  computed: {
    hasFlagEnricher(): boolean {
      return this.enrichers.some((enricher) => enricher.scope === 'flag')
    },
  },
  watch: {
    enrichers: {
      immediate: true,
      handler() {
        this.seedQuestions()
      },
    },
  },
  methods: {
    /** Seed each flag enricher's editor once, and drop editors for enrichers that went away. */
    seedQuestions() {
      const present = new Set(
        this.enrichers.filter((enricher) => enricher.scope === 'flag').map((enricher) => enricher.namespace),
      )
      for (const namespace of Object.keys(this.questions)) {
        if (!present.has(namespace)) delete this.questions[namespace]
      }
      for (const enricher of this.enrichers) {
        if (enricher.scope !== 'flag' || this.questions[enricher.namespace] !== undefined) continue
        const config = enricher.config as { questions?: Record<string, JevQuestion> } | undefined
        this.questions[enricher.namespace] = { ...(config?.questions ?? {}) }
      }
    },
    problemsFor(namespace: string): string[] {
      return jevQuestionProblems(this.questions[namespace] ?? {})
    },
    save(namespace: string) {
      this.$emit('save-enricher', {
        namespace,
        config: { questions: this.questions[namespace] ?? {} },
      })
    },
    addJev() {
      const fresh = defaultJevQuestion('noul')
      this.$emit('create-enricher', {
        namespace: 'jev',
        config: { questions: { example_question: fresh } },
      })
    },
  },
}
</script>

<style scoped>
.enrichers-card {
  margin-bottom: var(--space-md);
}

.enrichers-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-sm);
}

.enrichers-hint {
  margin: var(--space-3xs) 0 0;
  color: var(--el-text-color-secondary);
  font-size: var(--font-size-caption);
}

.enricher-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2xs);
}

.enricher-row {
  display: flex;
  flex-direction: column;
  gap: var(--space-2xs);
}

.enricher-row--flag {
  padding-top: var(--space-2xs);
  border-top: 1px solid var(--el-border-color-lighter);
}

.enricher-line {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-2xs);
}

.enricher-ns {
  font-family: var(--el-font-family-mono, monospace);
  font-size: var(--font-size-body-sm);
  font-weight: 600;
}

.enricher-props {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-3xs);
}

.enricher-no-props {
  color: var(--el-text-color-secondary);
  font-size: var(--font-size-caption);
}

.enricher-delete {
  margin-left: auto;
}

.enricher-edit {
  display: flex;
  flex-direction: column;
  gap: var(--space-2xs);
}

.enricher-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2xs);
}

.enricher-problem {
  color: var(--el-color-danger);
  font-size: var(--font-size-caption);
}
</style>
