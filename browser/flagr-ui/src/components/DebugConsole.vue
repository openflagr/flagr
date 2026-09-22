<template>
  <el-card class="dc-container">
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
            :json-string="evalContextText"
            :main-menu-bar="false"
            :navigation-bar="false"
            :status-bar="false"
            mode="text"
            class="dc-json-editor"
            @update:json-string="onEvalContextText"
          />
          <div class="dc-response-col">
            <json-editor
              :json-string="evalResultText"
              :main-menu-bar="false"
              :navigation-bar="false"
              :status-bar="false"
              mode="text"
              class="dc-json-editor"
              @update:json-string="onEvalResultText"
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
            :json-string="batchEvalContextText"
            :main-menu-bar="false"
            :navigation-bar="false"
            :status-bar="false"
            mode="text"
            class="dc-json-editor"
            @update:json-string="onBatchEvalContextText"
          />
          <json-editor
            :json-string="batchEvalResultText"
            :main-menu-bar="false"
            :navigation-bar="false"
            :status-bar="false"
            mode="text"
            class="dc-json-editor"
            @update:json-string="onBatchEvalResultText"
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
} from '@/helpers/evaluation'
import type { BatchEvalContext, BatchEvalResult, EvalContext, EvalResult, EvalSummary } from '@/api/types'

/** Stable text form used as the editor's source of truth. */
function toJsonText(value: unknown): string {
  return JSON.stringify(value, null, 2)
}

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
  data() {
    return {
      evalContextText: toJsonText(this.evalContext),
      evalResultText: toJsonText(this.evalResult),
      batchEvalContextText: toJsonText(this.batchEvalContext),
      batchEvalResultText: toJsonText(this.batchEvalResult),
      // Snapshot of the last value we pushed to the parent, so the editor can
      // ignore its own echo instead of resetting the caret on every keystroke.
      emittedEvalContext: toJsonText(this.evalContext),
      emittedEvalResult: toJsonText(this.evalResult),
      emittedBatchEvalContext: toJsonText(this.batchEvalContext),
      emittedBatchEvalResult: toJsonText(this.batchEvalResult),
    }
  },
  watch: {
    evalContext: { handler(value: EvalContext) { this.syncEvalContextFromProp(value) }, deep: true },
    evalResult: { handler(value: EvalResult) { this.syncEvalResultFromProp(value) }, deep: true },
    batchEvalContext: { handler(value: BatchEvalContext) { this.syncBatchEvalContextFromProp(value) }, deep: true },
    batchEvalResult: { handler(value: BatchEvalResult) { this.syncBatchEvalResultFromProp(value) }, deep: true },
  },
  methods: {
    onEvalContextText(text: string) {
      this.evalContextText = text
      const parsed = parseEvalContextJson(text)
      if (parsed) {
        this.emittedEvalContext = toJsonText(parsed)
        this.$emit('update:evalContext', parsed)
      }
    },
    onEvalResultText(text: string) {
      this.evalResultText = text
      const parsed = parseEvalResultJson(text)
      if (parsed) {
        this.emittedEvalResult = toJsonText(parsed)
        this.$emit('update:evalResult', parsed)
      }
    },
    onBatchEvalContextText(text: string) {
      this.batchEvalContextText = text
      const parsed = parseBatchEvalContextJson(text)
      if (parsed) {
        this.emittedBatchEvalContext = toJsonText(parsed)
        this.$emit('update:batchEvalContext', parsed)
      }
    },
    onBatchEvalResultText(text: string) {
      this.batchEvalResultText = text
      const parsed = parseBatchEvalResultJson(text)
      if (parsed) {
        this.emittedBatchEvalResult = toJsonText(parsed)
        this.$emit('update:batchEvalResult', parsed)
      }
    },
    syncEvalContextFromProp(value: EvalContext) {
      const text = toJsonText(value)
      if (text === this.emittedEvalContext) return
      this.emittedEvalContext = text
      this.evalContextText = text
    },
    syncEvalResultFromProp(value: EvalResult) {
      const text = toJsonText(value)
      if (text === this.emittedEvalResult) return
      this.emittedEvalResult = text
      this.evalResultText = text
    },
    syncBatchEvalContextFromProp(value: BatchEvalContext) {
      const text = toJsonText(value)
      if (text === this.emittedBatchEvalContext) return
      this.emittedBatchEvalContext = text
      this.batchEvalContextText = text
    },
    syncBatchEvalResultFromProp(value: BatchEvalResult) {
      const text = toJsonText(value)
      if (text === this.emittedBatchEvalResult) return
      this.emittedBatchEvalResult = text
      this.batchEvalResultText = text
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
  letter-spacing: var(--letter-spacing-tight);
  background: var(--el-color-primary-light-8);
  border-radius: 4px;
  padding: 2px var(--space-2xs);
  line-height: 1.5;
}
.dc-result-variant-value {
  font-size: 16px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  font-family: var(--font-mono);
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
  font-family: var(--font-mono);
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
