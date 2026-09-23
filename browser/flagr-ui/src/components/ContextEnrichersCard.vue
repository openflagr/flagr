<template>
  <el-card class="enrichers-card is-card-secondary">
    <template #header>
      <div class="el-card-header">
        <div class="enrichers-header">
          <h2>Context enrichers</h2>
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
      </div>
    </template>

    <p class="enrichers-hint">
      Properties injected into the evaluation context before constraints run. Pick them when adding a constraint.
    </p>

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
          <el-tooltip
            placement="top"
            effect="dark"
          >
            <code
              class="enricher-ns"
              :class="{ 'enricher-ns--disabled': enricher.enabled === false }"
              :data-testid="`enricher-ns-${enricher.namespace}`"
            >{{ enricher.namespace }}</code>
            <template #content>
              <div class="enricher-help">
                <p>{{ namespaceHelp(enricher.namespace) }}</p>
                <p
                  v-if="enricher.enabled === false"
                  class="enricher-help-warn"
                >
                  Not enabled on this server — constraints using its properties fail closed.
                </p>
              </div>
            </template>
          </el-tooltip>
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
          class="enricher-edit"
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
    /** What a namespace injects, shown as the namespace's tooltip. */
    namespaceHelp(namespace: string): string {
      switch (namespace) {
        case 'ts':
          return 'Built-in time context: @ts (unix seconds), @ts_hour, @ts_weekday, @ts_month.'
        case 'http':
          return 'Built-in request headers: @http_<header> for each header exposed by FLAGR_INJECTED_CONTEXT_HTTP_HEADERS / _HTTP_HEADER_PREFIXES.'
        case 'jev':
          return 'Model answers from a System One endpoint: one @jev_<question> property per question authored below.'
        default:
          return `Context enricher namespace "${namespace}".`
      }
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
  --enricher-help-width: 280px;
  --enricher-line-min-height: var(--space-lg);
}

.enrichers-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-sm);
}

.enrichers-hint {
  margin: 0 0 var(--space-2xs);
  color: var(--el-text-color-secondary);
  font-size: var(--font-size-caption);
  line-height: var(--line-height-ui);
}

.enricher-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2xs);
}

.enricher-row {
  display: flex;
  flex-direction: column;
  gap: var(--space-3xs);
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
  min-height: var(--enricher-line-min-height);
}

.enricher-ns {
  font-family: var(--font-mono);
  font-size: var(--font-size-body-sm);
  font-weight: var(--font-weight-semibold);
  cursor: help;
  border-bottom: 1px dotted var(--el-border-color);
}

.enricher-ns--disabled {
  color: var(--el-color-warning);
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

.enricher-help {
  max-width: var(--enricher-help-width);
  font-size: var(--font-size-caption);
  line-height: var(--line-height-ui);
}

.enricher-help p {
  margin: 0;
}

.enricher-help-warn {
  margin-top: var(--space-3xs) !important;
  color: var(--el-color-warning);
}
</style>
