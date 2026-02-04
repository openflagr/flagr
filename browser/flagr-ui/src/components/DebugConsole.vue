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
          <el-col :span="7" class="evaluation-button-col">
            <el-button
              size="small"
              @click="postEvaluation(evalContext)"
              type="primary"
              plain
            >POST /api/v1/evaluation</el-button>
          </el-col>
          <el-col :span="6">
            <span>Response</span>
          </el-col>
        </el-row>
        <el-row :gutter="10">
          <el-col :span="12">
            <vue3-json-editor
              v-model="evalContext"
              :show-btns="false"
              :mode="'code'"
              class="json-editor"
            />
          </el-col>
          <el-col :span="12">
            <vue3-json-editor
              v-model="evalResult"
              :show-btns="false"
              :mode="'code'"
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
          <el-col :span="7" class="evaluation-button-col">
            <el-button
              size="small"
              @click="postEvaluationBatch(batchEvalContext)"
              type="primary"
              plain
            >POST /api/v1/evaluation/batch</el-button>
          </el-col>
          <el-col :span="6">
            <span>Response</span>
          </el-col>
        </el-row>
        <el-row :gutter="10">
          <el-col :span="12">
            <vue3-json-editor
              v-model="batchEvalContext"
              :show-btns="false"
              :mode="'code'"
              class="json-editor"
            />
          </el-col>
          <el-col :span="12">
            <vue3-json-editor
              v-model="batchEvalResult"
              :show-btns="false"
              :mode="'code'"
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
import { Vue3JsonEditor } from "vue3-json-editor";
import { ElMessage } from "element-plus";

import constants from "@/constants";

const { API_URL } = constants;

const props = defineProps(["flag"]);

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
  .jsoneditor {
    height: 400px;
  }
}
.evaluation-button-col {
  text-align: right;
}
</style>
