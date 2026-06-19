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
          <json-editor :json="evalContext" @update:json="evalContext = $event" @update:jsonString="syncEvalContext" :main-menu-bar="false" :navigation-bar="false" :status-bar="false" mode="text" class="dc-json-editor" />
          <div class="dc-response-col">
            <json-editor :json="evalResult" @update:json="evalResult = $event" @update:jsonString="syncEvalResult" :main-menu-bar="false" :navigation-bar="false" :status-bar="false" mode="text" class="dc-json-editor" />
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
              <div v-for="seg in evalSummary.segments" :key="seg.segmentID" class="dc-segment-log-item">
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
          <json-editor :json="batchEvalContext" @update:json="batchEvalContext = $event" @update:jsonString="syncBatchEvalContext" :main-menu-bar="false" :navigation-bar="false" :status-bar="false" mode="text" class="dc-json-editor" />
          <json-editor :json="batchEvalResult" @update:json="batchEvalResult = $event" @update:jsonString="syncBatchEvalResult" :main-menu-bar="false" :navigation-bar="false" :status-bar="false" mode="text" class="dc-json-editor" />
        </div>
      </el-collapse-item>
    </el-collapse>
  </el-card>
</template>

<script>
import Axios from "axios";
import JsonEditor from "vue3-ts-jsoneditor";
import constants from "@/constants";
const { API_URL } = constants;

export default {
  name: "debug-console",
  components: { JsonEditor },
  props: ["flag"],
  data() {
    const flagId = this.flag && this.flag.id;
    const flagKey = this.flag && this.flag.key;
    return {
      evalContext: { entityID: "a1234", entityType: "report", entityContext: { hello: "world" }, enableDebug: true, flagID: flagId, flagKey: flagKey },
      evalResult: {},
      evalSummary: null,
      batchEvalContext: { entities: [{ entityID: "a1234", entityType: "report", entityContext: { hello: "world" } }, { entityID: "a5678", entityType: "report", entityContext: { hello: "world" } }], enableDebug: true, flagIDs: [flagId] },
      batchEvalResult: {}
    };
  },
  methods: {
    syncEvalContext(text) { try { this.evalContext = JSON.parse(text) } catch(e) {} },
    syncEvalResult(text) { try { this.evalResult = JSON.parse(text) } catch(e) {} },
    syncBatchEvalContext(text) { try { this.batchEvalContext = JSON.parse(text) } catch(e) {} },
    syncBatchEvalResult(text) { try { this.batchEvalResult = JSON.parse(text) } catch(e) {} },

    postEvaluation(evalContext) {
      Axios.post(`${API_URL}/evaluation`, evalContext).then(response => {
        this.evalResult = response.data;
        this.evalSummary = this.buildSummary(response.data);
        this.$message.success("evaluation success");
      }, err => { this.$message.error(err?.response?.data?.message || 'evaluation error') });
    },
    postEvaluationBatch(batchEvalContext) {
      Axios.post(`${API_URL}/evaluation/batch`, batchEvalContext).then(response => {
        this.batchEvalResult = response.data;
        this.$message.success("evaluation success");
      }, err => { this.$message.error(err?.response?.data?.message || 'evaluation error') });
    },
    buildSummary(result) {
      if (!result || !result.evalDebugLog) return null;
      const log = result.evalDebugLog;
      const segments = (log.segmentDebugLogs || []).map(s => ({
        segmentID: s.segmentID,
        description: s.description,
        rolloutPercent: s.rolloutPercent,
        matched: s.matched,
        msg: s.msg,
        constraints: (s.constraintDebugLogs || []).map(c => ({
          constraintID: c.constraintID,
          constraintProperty: c.constraintProperty,
          constraintOperator: c.constraintOperator,
          constraintValue: c.constraintValue,
          matched: c.matched
        }))
      }));
      return {
        variantKey: result.variantKey || "—",
        variantID: result.variantID,
        segments
      };
    }
  },
  watch: {
    flag: {
      immediate: true,
      handler(f) {
        if (f && f.id) {
          this.evalContext.flagID = f.id;
          this.evalContext.flagKey = f.key;
          this.batchEvalContext.flagIDs = [f.id];
        }
      }
    }
  }
};
</script>

<style lang="less" scoped>
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
