<template>
  <el-card class="dc-container">
    <template #header>
      <div class="el-card-header">
        <h2>Debug Console</h2>
      </div>
    </template>
    <el-collapse>
      <el-collapse-item title="Evaluation">
        <el-row :gutter="10">
          <el-col :span="5">
            <span>Request</span>
          </el-col>
          <el-col
            :span="7"
            class="evaluation-button-col"
          >
            <el-button
              size="small"
              type="primary"
              plain
              @click="postEvaluation(evalContext)"
            >
              POST /api/v1/evaluation
            </el-button>
          </el-col>
          <el-col :span="6">
            <span>Response</span>
          </el-col>
        </el-row>
        <el-row :gutter="10">
          <el-col :span="12">
            <JsonEditorVue
              v-model="evalContext"
              :mode="'text'"
              :main-menu-bar="false"
              :navigation-bar="false"
              :status-bar="false"
              class="json-editor"
            />
          </el-col>
          <el-col :span="12">
            <JsonEditorVue
              v-model="evalResult"
              :mode="'text'"
              :main-menu-bar="false"
              :navigation-bar="false"
              :status-bar="false"
              class="json-editor"
            />
          </el-col>
        </el-row>
      </el-collapse-item>

      <el-collapse-item title="Batch Evaluation">
        <el-row :gutter="10">
          <el-col :span="5">
            <span>Request</span>
          </el-col>
          <el-col
            :span="7"
            class="evaluation-button-col"
          >
            <el-button
              size="small"
              type="primary"
              plain
              @click="postEvaluationBatch(batchEvalContext)"
            >
              POST /api/v1/evaluation/batch
            </el-button>
          </el-col>
          <el-col :span="6">
            <span>Response</span>
          </el-col>
        </el-row>
        <el-row :gutter="10">
          <el-col :span="12">
            <JsonEditorVue
              v-model="batchEvalContext"
              :mode="'text'"
              :main-menu-bar="false"
              :navigation-bar="false"
              :status-bar="false"
              class="json-editor"
            />
          </el-col>
          <el-col :span="12">
            <JsonEditorVue
              v-model="batchEvalResult"
              :mode="'text'"
              :main-menu-bar="false"
              :navigation-bar="false"
              :status-bar="false"
              class="json-editor"
            />
          </el-col>
        </el-row>
      </el-collapse-item>
    </el-collapse>
  </el-card>
</template>

<script setup>
import { ref } from "vue";
import Axios from "axios";
import JsonEditorVue from "json-editor-vue";
import { ElMessage } from "element-plus";

import constants from "@/constants";

const props = defineProps({
  flag: {
    type: Object,
    required: true,
  },
});

const { API_URL } = constants;

const evalContext = ref({
  entityID: "a1234",
  entityType: "report",
  entityContext: {
    hello: "world"
  },
  enableDebug: true,
  flagID: props.flag.id,
  flagKey: props.flag.key
});
const evalResult = ref({});

const batchEvalContext = ref({
  entities: [
    {
      entityID: "a1234",
      entityType: "report",
      entityContext: {
        hello: "world"
      }
    },
    {
      entityID: "a5678",
      entityType: "report",
      entityContext: {
        hello: "world"
      }
    }
  ],
  enableDebug: true,
  flagIDs: [props.flag.id]
});
const batchEvalResult = ref({});

function postEvaluation(evalCtx) {
  Axios.post(`${API_URL}/evaluation`, evalCtx).then(
    response => {
      ElMessage.success(`evaluation success`);
      evalResult.value = response.data;
    },
    () => {
      ElMessage.error(`evaluation error`);
    }
  );
}

function postEvaluationBatch(batchEvalCtx) {
  Axios.post(`${API_URL}/evaluation/batch`, batchEvalCtx).then(
    response => {
      ElMessage.success(`evaluation success`);
      batchEvalResult.value = response.data;
    },
    () => {
      ElMessage.error(`evaluation error`);
    }
  );
}
</script>

<style lang="less" scoped>
.json-editor {
  margin-top: 3px;
  :deep(.jse-main) {
    height: 400px;
  }
}
.evaluation-button-col {
  text-align: right;
}
</style>
