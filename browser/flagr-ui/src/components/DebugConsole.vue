<template>
  <el-card class="dc-container is-card-utility">
    <template #header>
      <div class="el-card-header">
        <h2>Debug Console</h2>
      </div>
    </template>
    <el-collapse>
      <el-collapse-item
        title="Evaluation"
        class="dc-collapse-item"
      >
        <div class="dc-eval-header">
          <span class="dc-label">Request</span>
          <el-button
            size="small"
            type="primary"
            plain
            @click="$emit('post-evaluation', evalContext)"
          >
            POST /api/v1/evaluation
          </el-button>
        </div>
        <div class="dc-editor-row">
          <json-editor
            :json="evalContext"
            :main-menu-bar="false"
            :navigation-bar="false"
            :status-bar="false"
            mode="text"
            class="dc-json-editor"
            @update:json="onEvalContextJson"
            @update:json-string="syncEvalContext"
          />
          <div class="dc-response-col">
            <json-editor
              :json="evalResult"
              :main-menu-bar="false"
              :navigation-bar="false"
              :status-bar="false"
              mode="text"
              class="dc-json-editor"
              @update:json="onEvalResultJson"
              @update:json-string="syncEvalResult"
            />
          </div>
        </div>
        <div
          v-if="evalSummary"
          class="dc-summary"
        >
          <div class="dc-summary-header">
            Rendered Result
          </div>
          <div class="dc-summary-body">
            <div class="dc-result-variant">
              <span class="dc-result-variant-label">Variant</span>
              <span class="dc-result-variant-value">{{ evalSummary.variantKey }}</span>
            </div>
            <div class="dc-segment-log">
              <div
                v-for="seg in evalSummary.segments"
                :key="String(seg.segmentID)"
                class="dc-segment-log-item"
              >
                <div class="dc-segment-log-header">
                  <span class="dc-seg-name">segment #{{ seg.segmentID }}</span>
                </div>
                <div
                  v-if="seg.msg"
                  class="dc-seg-msg"
                >
                  {{ seg.msg }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-collapse-item>
      <el-collapse-item
        title="Batch Evaluation"
        class="dc-collapse-item"
      >
        <div class="dc-eval-header">
          <span class="dc-label">Request</span>
          <el-button
            size="small"
            type="primary"
            plain
            @click="$emit('post-evaluation-batch', batchEvalContext)"
          >
            POST /api/v1/evaluation/batch
          </el-button>
          <span class="dc-label">Response</span>
        </div>
        <div class="dc-editor-row">
          <json-editor
            :json="batchEvalContext"
            :main-menu-bar="false"
            :navigation-bar="false"
            :status-bar="false"
            mode="text"
            class="dc-json-editor"
            @update:json="onBatchEvalContextJson"
            @update:json-string="syncBatchEvalContext"
          />
          <json-editor
            :json="batchEvalResult"
            :main-menu-bar="false"
            :navigation-bar="false"
            :status-bar="false"
            mode="text"
            class="dc-json-editor"
            @update:json="onBatchEvalResultJson"
            @update:json-string="syncBatchEvalResult"
          />
        </div>
      </el-collapse-item>
    </el-collapse>
  </el-card>
</template>

<script lang="ts">
import JsonEditor from 'vue3-ts-jsoneditor'
import {
  parseBatchEvalContextJson,
  parseBatchEvalResultJson,
  parseEvalContextJson,
  parseEvalResultJson,
} from '@/helpers/evalParse'
import { asBatchEvalResult, asJsonObject } from '@/helpers/jsonEditorValue'
import type { BatchEvalContext, BatchEvalResult, EvalContext, EvalResult, EvalSummary } from '@/api/types'

export default {
  name: 'DebugConsole',
  components: { JsonEditor },
  props: {
    evalContext: { type: Object as () => EvalContext, required: true },
    evalResult: { type: Object as () => EvalResult, required: true },
    evalSummary: { type: Object as () => EvalSummary | null, default: null },
    batchEvalContext: { type: Object as () => BatchEvalContext, required: true },
    batchEvalResult: { type: Object as () => BatchEvalResult, required: true },
  },
  emits: [
    'update:evalContext',
    'update:evalResult',
    'update:batchEvalContext',
    'update:batchEvalResult',
    'post-evaluation',
    'post-evaluation-batch',
  ],
  methods: {
    onEvalContextJson(v: unknown) {
      const o = asJsonObject(v)
      if (o) this.$emit('update:evalContext', o as EvalContext)
    },
    onEvalResultJson(v: unknown) {
      const o = asJsonObject(v)
      if (o) this.$emit('update:evalResult', o as EvalResult)
    },
    onBatchEvalContextJson(v: unknown) {
      const o = asJsonObject(v)
      if (o) this.$emit('update:batchEvalContext', o as BatchEvalContext)
    },
    onBatchEvalResultJson(v: unknown) {
      const parsed = asBatchEvalResult(v)
      if (parsed) this.$emit('update:batchEvalResult', parsed)
    },
    syncEvalContext(text: string) {
      const parsed = parseEvalContextJson(text)
      if (parsed) this.$emit('update:evalContext', parsed)
    },
    syncEvalResult(text: string) {
      const parsed = parseEvalResultJson(text)
      if (parsed) this.$emit('update:evalResult', parsed)
    },
    syncBatchEvalContext(text: string) {
      const parsed = parseBatchEvalContextJson(text)
      if (parsed) this.$emit('update:batchEvalContext', parsed)
    },
    syncBatchEvalResult(text: string) {
      const parsed = parseBatchEvalResultJson(text)
      if (parsed) this.$emit('update:batchEvalResult', parsed)
    },
  },
}
</script>

<style lang="scss" scoped>
.dc-eval-header {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  margin-bottom: var(--space-2xs);
}
.dc-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-regular);
}
.dc-editor-row {
  display: flex;
  gap: var(--space-xs);
}
.dc-json-editor {
  flex: 1;
  height: 280px;
}
.dc-collapse-item {
  :deep(.el-collapse-item__content) { padding-bottom: var(--space-2xs); }
}

// --- Summary ---
.dc-summary {
  margin-top: var(--space-xs);
  border: 1px solid var(--el-border-color-light);
  border-radius: 10px;
  overflow: hidden;
}
.dc-summary-header {
  font-size: 12px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  padding: var(--space-2xs) var(--space-xs);
  background: var(--el-fill-color-light);
  border-bottom: 1px solid var(--el-border-color-light);
}
.dc-summary-body {
  padding: var(--space-xs) var(--space-xs);
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}
.dc-result-variant {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
  background: var(--el-color-primary-light-9);
  border: 1px solid var(--el-color-primary-light-7);
  border-radius: 8px;
  padding: var(--space-2xs) var(--space-xs);
}
.dc-result-variant-label {
  font-size: 10px;
  font-weight: 700;
  color: var(--el-color-primary);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  background: var(--el-color-primary-light-8);
  border-radius: 4px;
  padding: 2px var(--space-2xs);
  line-height: 1.5;
}
.dc-result-variant-value {
  font-size: 16px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
}

.dc-segment-log {
  display: flex;
  flex-direction: column;
  gap: var(--space-2xs);
}


.dc-segment-log-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-3xs);
  background: var(--el-fill-color-light);
  border-radius: 8px;
  padding: var(--space-2xs) var(--space-xs);
}
.dc-segment-log-header {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
  flex-wrap: wrap;
}
.dc-seg-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.dc-seg-msg {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
  line-height: 1.4;
  font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
  word-break: break-all;
}
.dc-response-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--space-2xs);
}
.dc-response-col .dc-json-editor {
  flex: 1;
}
@media (max-width: 768px) {
  .dc-editor-row {
    flex-direction: column;
  }
  .dc-json-editor {
    height: 200px;
  }
}
</style>
