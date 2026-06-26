<template>
  <el-card class="dc-container is-card-utility">
    <template #header>
      <div class="el-card-header"><h2>Debug Console</h2></div>
    </template>
    <el-collapse>
      <el-collapse-item title="Evaluation" class="dc-collapse-item">
        <div class="dc-eval-header">
          <span class="dc-label">Request</span>
          <el-button size="small" @click="postEvaluation(evalContext)" type="primary" plain>POST /api/v1/evaluation</el-button>
        </div>
        <div class="dc-editor-row">
          <json-editor :json="evalContext" @update:json="onEvalContextJson" @update:jsonString="syncEvalContext" :main-menu-bar="false" :navigation-bar="false" :status-bar="false" mode="text" class="dc-json-editor" />
          <div class="dc-response-col">
            <json-editor :json="evalResult" @update:json="onEvalResultJson" @update:jsonString="syncEvalResult" :main-menu-bar="false" :navigation-bar="false" :status-bar="false" mode="text" class="dc-json-editor" />
          </div>
        </div>
        <div v-if="evalSummary" class="dc-summary">
          <div class="dc-summary-header">Rendered Result</div>
          <div class="dc-summary-body">
            <div class="dc-result-variant">
              <span class="dc-result-variant-label">Variant</span>
              <span class="dc-result-variant-value">{{ evalSummary.variantKey }}</span>
            </div>
            <div class="dc-segment-log">
              <div v-for="seg in evalSummary.segments" :key="String(seg.segmentID)" class="dc-segment-log-item">
                <div class="dc-segment-log-header">
                  <span class="dc-seg-name">segment #{{ seg.segmentID }}</span>
                </div>
                <div v-if="seg.msg" class="dc-seg-msg">{{ seg.msg }}</div>
              </div>
            </div>
          </div>
        </div>
      </el-collapse-item>
      <el-collapse-item title="Batch Evaluation" class="dc-collapse-item">
        <div class="dc-eval-header">
          <span class="dc-label">Request</span>
          <el-button size="small" @click="postEvaluationBatch(batchEvalContext)" type="primary" plain>POST /api/v1/evaluation/batch</el-button>
          <span class="dc-label">Response</span>
        </div>
        <div class="dc-editor-row">
          <json-editor :json="batchEvalContext" @update:json="onBatchEvalContextJson" @update:jsonString="syncBatchEvalContext" :main-menu-bar="false" :navigation-bar="false" :status-bar="false" mode="text" class="dc-json-editor" />
          <json-editor :json="batchEvalResult" @update:json="onBatchEvalResultJson" @update:jsonString="syncBatchEvalResult" :main-menu-bar="false" :navigation-bar="false" :status-bar="false" mode="text" class="dc-json-editor" />
        </div>
      </el-collapse-item>
    </el-collapse>
  </el-card>
</template>

<script lang="ts">
import JsonEditor from 'vue3-ts-jsoneditor'
import * as evalApi from '@/api/evaluation'
import type { BatchEvalContext, EvalContext, EvalResult, EvalSummary, Flag } from '@/api/types'
import { runApi } from '@/helpers/runApi'

export default {
  name: 'debug-console',
  components: { JsonEditor },
  props: ['flag'],
  data() {
    const flag = this.flag as Flag | undefined
    const flagId = flag?.id
    const flagKey = flag?.key
    return {
      evalContext: {
        entityID: 'a1234',
        entityType: 'report',
        entityContext: { hello: 'world' },
        enableDebug: true,
        flagID: flagId,
        flagKey: flagKey,
      } as EvalContext,
      evalResult: {} as EvalResult,
      evalSummary: null as EvalSummary | null,
      batchEvalContext: {
        entities: [
          { entityID: 'a1234', entityType: 'report', entityContext: { hello: 'world' } },
          { entityID: 'a5678', entityType: 'report', entityContext: { hello: 'world' } },
        ],
        enableDebug: true,
        flagIDs: flagId ? [flagId] : [],
      } as BatchEvalContext,
      batchEvalResult: {} as Record<string, unknown>,
    }
  },
  methods: {
    onEvalContextJson(v: unknown) {
      this.evalContext = v as EvalContext
    },
    onEvalResultJson(v: unknown) {
      this.evalResult = v as EvalResult
    },
    onBatchEvalContextJson(v: unknown) {
      this.batchEvalContext = v as BatchEvalContext
    },
    onBatchEvalResultJson(v: unknown) {
      this.batchEvalResult = v as Record<string, unknown>
    },
    syncEvalContext(text: string) {
      try {
        this.evalContext = JSON.parse(text) as EvalContext
      } catch {
        /* ignore */
      }
    },
    syncEvalResult(text: string) {
      try {
        this.evalResult = JSON.parse(text) as EvalResult
      } catch {
        /* ignore */
      }
    },
    syncBatchEvalContext(text: string) {
      try {
        this.batchEvalContext = JSON.parse(text) as BatchEvalContext
      } catch {
        /* ignore */
      }
    },
    syncBatchEvalResult(text: string) {
      try {
        this.batchEvalResult = JSON.parse(text) as Record<string, unknown>
      } catch {
        /* ignore */
      }
    },

    postEvaluation(evalContext: EvalContext) {
      runApi(this, evalApi.postEvaluation(evalContext), {
        successMessage: 'evaluation success',
        onSuccess: (response) => {
          this.evalResult = response
          this.evalSummary = this.buildSummary(response)
        },
      })
    },
    postEvaluationBatch(batchEvalContext: BatchEvalContext) {
      runApi(this, evalApi.postEvaluationBatch(batchEvalContext), {
        successMessage: 'evaluation success',
        onSuccess: (response) => {
          this.batchEvalResult = response
        },
      })
    },
    buildSummary(result: EvalResult): EvalSummary | null {
      if (!result || !result.evalDebugLog) return null
      const log = result.evalDebugLog as Record<string, unknown>
      const segments = ((log.segmentDebugLogs as unknown[]) || []).map((s) => {
        const seg = s as Record<string, unknown>
        return {
          segmentID: seg.segmentID,
          description: seg.description,
          rolloutPercent: seg.rolloutPercent,
          matched: seg.matched,
          msg: seg.msg,
          constraints: ((seg.constraintDebugLogs as unknown[]) || []).map((c) => {
            const con = c as Record<string, unknown>
            return {
              constraintID: con.constraintID,
              constraintProperty: con.constraintProperty,
              constraintOperator: con.constraintOperator,
              constraintValue: con.constraintValue,
              matched: con.matched,
            }
          }),
        }
      })
      return {
        variantKey: (result.variantKey as string) || '—',
        variantID: result.variantID,
        segments,
      }
    },
  },
  watch: {
    flag: {
      immediate: true,
      handler(f: Flag) {
        if (f?.id) {
          this.evalContext.flagID = f.id
          this.evalContext.flagKey = f.key
          this.batchEvalContext.flagIDs = [f.id]
        }
      },
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
