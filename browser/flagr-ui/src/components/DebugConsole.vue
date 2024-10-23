<template>
  <el-card class="dc-container">
    <div slot="header" class="el-card-header">
      <h2>Debug Console</h2>
    </div>
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
            <Vue3JsonEditor
              v-model="evalContext"
              :show-btns="false"
              mode="code"
              :expandedOnStart="true"
              class="json-editor"
              @json-change="onJsonChange('evalContext', $event)"
            ></Vue3JsonEditor>
          </el-col>
          <el-col :span="12">
            <Vue3JsonEditor
              v-model="evalResult"
              :show-btns="false"
              mode="code"
              :expandedOnStart="true"
              class="json-editor"
              @json-change="onJsonChange('evalResult', $event)"
            ></Vue3JsonEditor>
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
            <Vue3JsonEditor
              v-model="batchEvalContext"
              :show-btns="false"
              mode="code"
              :expandedOnStart="true"
              class="json-editor"
              @json-change="onJsonChange('batchEvalContext', $event)"
            ></Vue3JsonEditor>
          </el-col>
          <el-col :span="12">
            <Vue3JsonEditor
              v-model="batchEvalResult"
              :show-btns="false"
              mode="code"
              :expandedOnStart="true"
              class="json-editor"
              @json-change="onJsonChange('batchEvalResult', $event)"
            ></Vue3JsonEditor>
          </el-col>
        </el-row>
      </el-collapse-item>
    </el-collapse>
  </el-card>
</template>

<script>
import { Vue3JsonEditor } from 'vue3-json-editor'
import { getAxiosFlagrInstance } from '../utils/apiUtil';

export default {
  name: "debug-console",
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
    onJsonChange(editorType, value) {
      if(editorType === 'evalContext'){
        this.evalContext = value;
      } else if(editorType === 'evalResult'){
        this.evalResult = value;
      } else if(editorType === 'batchEvalContext'){
        this.batchEvalContext = value;
      } else if(editorType === 'batchEvalResult'){
        this.batchEvalResult = value;
      }
    },
    postEvaluation(evalContext) {
      getAxiosFlagrInstance().post(`/evaluation`, evalContext).then(
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
      getAxiosFlagrInstance().post(`/evaluation/batch`, batchEvalContext).then(
        response => {
          this.$message.success(`evaluation success`);
          this.batchEvalResult = response.data;
        },
        () => {
          this.$message.error(`evaluation error`);
        }
      );
    }
  },
  components: {
    Vue3JsonEditor
  },
  mounted() {
    this.$refs.evalContextEditor?.targeteditor?.setMode("code");
    this.$refs.evalResultEditor?.editor?.setMode("code");
    this.$refs.batchEvalContextEditor?.editor?.setMode("code");
    this.$refs.batchEvalResultEditor?.editor?.setMode("code");
  }
};
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
