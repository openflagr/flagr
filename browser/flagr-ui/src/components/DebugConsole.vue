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
            <json-editor
              v-model:json="evalContext"
              :main-menu-bar="false"
              :navigation-bar="false"
              :status-bar="false"
              mode="text"
              class="json-editor"
            />
          </el-col>
          <el-col :span="12">
            <json-editor
              v-model:json="evalResult"
              :main-menu-bar="false"
              :navigation-bar="false"
              :status-bar="false"
              mode="text"
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
            <json-editor
              v-model:json="batchEvalContext"
              :main-menu-bar="false"
              :navigation-bar="false"
              :status-bar="false"
              mode="text"
              class="json-editor"
            />
          </el-col>
          <el-col :span="12">
            <json-editor
              v-model:json="batchEvalResult"
              :main-menu-bar="false"
              :navigation-bar="false"
              :status-bar="false"
              mode="text"
              class="json-editor"
            />
          </el-col>
        </el-row>
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
    return {
      evalContext: {
        entityID: "a1234",
        entityType: "report",
        entityContext: {
          hello: "world"
        },
        enableDebug: true,
        flagID: this.flag.id,
        flagKey: this.flag.key
      },
      evalResult: {},
      batchEvalContext: {
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
        flagIDs: [this.flag.id]
      },
      batchEvalResult: {}
    };
  },
  methods: {
    postEvaluation(evalContext) {
      Axios.post(`${API_URL}/evaluation`, evalContext).then(
        response => {
          this.$message.success(`evaluation success`);
          this.evalResult = response.data;
        },
        () => {
          this.$message.error(`evaluation error`);
        }
      );
    },
    postEvaluationBatch(batchEvalContext) {
      Axios.post(`${API_URL}/evaluation/batch`, batchEvalContext).then(
        response => {
          this.$message.success(`evaluation success`);
          this.batchEvalResult = response.data;
        },
        () => {
          this.$message.error(`evaluation error`);
        }
      );
    }
  }
};
</script>

<style lang="less" scoped>
.json-editor {
  margin-top: 3px;
  height: 400px;
}
.evaluation-button-col {
  text-align: right;
}
</style>
