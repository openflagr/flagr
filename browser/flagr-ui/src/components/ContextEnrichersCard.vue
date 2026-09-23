<template>
  <el-card class="enrichers-card">
    <template #header>
      <div class="enrichers-header">
        <span class="ui-section-title">Context enrichers</span>
        <span class="enrichers-hint">
          Properties injected into the evaluation context before constraints run. Pick them when adding a constraint.
        </span>
      </div>
    </template>

    <div
      v-if="!enrichers.length"
      class="card--empty"
    >
      No enrichers enabled. Enable built-ins with <code>FLAGR_INJECTED_CONTEXT_ENABLED</code>, or add a Jev enricher below.
    </div>

    <div
      v-for="enricher in enrichers"
      :key="enricher.namespace"
      class="enricher-row ui-surface-inset"
      :data-testid="`enricher-row-${enricher.namespace}`"
    >
      <div class="enricher-row-head">
        <span class="enricher-ns">{{ enricher.namespace }}</span>
        <el-tag
          size="small"
          :type="enricher.scope === 'global' ? 'info' : 'primary'"
        >
          {{ enricher.scope === 'global' ? 'built-in' : 'flag' }}
        </el-tag>
        <el-button
          v-if="!readonly && enricher.scope === 'flag'"
          size="small"
          link
          type="danger"
          :data-testid="`delete-enricher-${enricher.namespace}`"
          @click="$emit('delete-enricher', enricher.namespace)"
        >
          Delete
        </el-button>
      </div>

      <div class="enricher-props">
        <el-tag
          v-for="property in enricher.properties"
          :key="property"
          size="small"
          effect="plain"
        >
          {{ property }}
        </el-tag>
        <span
          v-if="!(enricher.properties ?? []).length"
          class="enricher-no-props"
        >no properties</span>
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
          <el-button
            size="small"
            type="primary"
            :data-testid="`save-enricher-${enricher.namespace}`"
            @click="save(enricher.namespace)"
          >
            Save config
          </el-button>
        </div>
      </div>
    </div>

    <div
      v-if="!readonly && !hasFlagEnricher"
      class="enricher-add"
    >
      <el-button
        size="small"
        data-testid="add-jev-enricher-btn"
        @click="addJev"
      >
        Add Jev enricher
      </el-button>
    </div>
  </el-card>
</template>

<script lang="ts">
import type { PropType } from 'vue'
import type { Enricher, JevQuestion } from '@/api/types'
import JevQuestionsEditor from '@/components/JevQuestionsEditor.vue'
import { defaultJevQuestion } from '@/helpers/jevQuestion'

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
    seedQuestions() {
      for (const enricher of this.enrichers) {
        if (enricher.scope !== 'flag' || this.questions[enricher.namespace] !== undefined) continue
        const config = enricher.config as { questions?: Record<string, JevQuestion> } | undefined
        this.questions[enricher.namespace] = { ...(config?.questions ?? {}) }
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
  margin-bottom: var(--space-md);
}

.enrichers-header {
  display: flex;
  flex-direction: column;
  gap: var(--space-3xs);
}

.enrichers-hint {
  color: var(--el-text-color-secondary);
  font-size: var(--font-size-caption);
}

.enricher-row {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  margin-bottom: var(--space-sm);
}

.enricher-row-head {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.enricher-ns {
  font-weight: 600;
  font-size: var(--font-size-body-sm);
}

.enricher-props {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3xs);
}

.enricher-no-props {
  color: var(--el-text-color-secondary);
  font-size: var(--font-size-caption);
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
  gap: var(--space-sm);
}

.enricher-add {
  margin-top: var(--space-sm);
}
</style>
