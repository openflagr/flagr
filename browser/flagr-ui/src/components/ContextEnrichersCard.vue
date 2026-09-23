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
        <el-input
          v-model="drafts[enricher.namespace]"
          type="textarea"
          :rows="6"
          :data-testid="`enricher-config-${enricher.namespace}`"
        />
        <div class="enricher-actions">
          <span class="enricher-error">{{ errors[enricher.namespace] }}</span>
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
import type { Enricher } from '@/api/types'

/** Minimal starting config so the editor is never empty. */
const DEFAULT_JEV_CONFIG = JSON.stringify(
  {
    questions: {
      example_question: {
        type: 'noul',
        instructions: 'Describe the atomic yes/no question for the model.',
      },
    },
  },
  null,
  2,
)

export default {
  name: 'ContextEnrichersCard',
  props: {
    enrichers: { type: Array as PropType<Enricher[]>, default: () => [] },
    readonly: { type: Boolean, default: false },
  },
  emits: ['create-enricher', 'save-enricher', 'delete-enricher'],
  data() {
    return {
      drafts: {} as Record<string, string>,
      errors: {} as Record<string, string>,
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
        this.seedDrafts()
      },
    },
  },
  methods: {
    seedDrafts() {
      for (const enricher of this.enrichers) {
        if (enricher.scope === 'flag' && this.drafts[enricher.namespace] === undefined) {
          this.drafts[enricher.namespace] = JSON.stringify(enricher.config ?? {}, null, 2)
        }
      }
    },
    save(namespace: string) {
      const config = this.parseConfig(namespace)
      if (config === undefined) return
      this.$emit('save-enricher', { namespace, config })
    },
    addJev() {
      let config: unknown
      try {
        config = JSON.parse(DEFAULT_JEV_CONFIG)
      } catch {
        config = {}
      }
      this.$emit('create-enricher', { namespace: 'jev', config })
    },
    parseConfig(namespace: string): unknown | undefined {
      try {
        const config = JSON.parse(this.drafts[namespace] ?? '{}')
        this.errors[namespace] = ''
        return config
      } catch (error) {
        this.errors[namespace] = `Invalid JSON: ${(error as Error).message}`
        return undefined
      }
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

.enricher-error {
  color: var(--el-color-danger);
  font-size: var(--font-size-caption);
}

.enricher-add {
  margin-top: var(--space-sm);
}
</style>
