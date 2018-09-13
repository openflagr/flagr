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
            <el-button size="mini" @click="postEvaluation(evalContext)" type="primary" plain>POST /api/v1/evaluation</el-button>
          </el-col>
          <el-col :span="6">
            <span>Response</span>
          </el-col>
        </el-row>
        <el-row :gutter="10">
          <el-col :span="12">
            <vue-json-editor v-model="evalContext" :showBtns="false" ref="evalContextEditor" class="json-editor"></vue-json-editor>
          </el-col>
          <el-col :span="12">
            <vue-json-editor v-model="evalResult" :showBtns="false" ref="evalResultEditor" class="json-editor"></vue-json-editor>
          </el-col>
        </el-row>
      </el-collapse-item>

      <el-collapse-item title="Batch Evaluation">
        <el-row :gutter="10">
          <el-col :span="5">
            <span>Request</span>
          </el-col>
          <el-col :span="7" class="evaluation-button-col">
            <el-button size="mini" @click="postEvaluationBatch(batchEvalContext)" type="primary" plain>POST /api/v1/evaluation/batch</el-button>
          </el-col>
          <el-col :span="6">
            <span>Response</span>
          </el-col>
        </el-row>
        <el-row :gutter="10">
          <el-col :span="12">
            <vue-json-editor v-model="batchEvalContext" :showBtns="false" ref="batchEvalContextEditor" class="json-editor"></vue-json-editor>
          </el-col>
          <el-col :span="12">
            <vue-json-editor v-model="batchEvalResult" :showBtns="false" ref="batchEvalResultEditor" class="json-editor"></vue-json-editor>
          </el-col>
        </el-row>
      </el-collapse-item>
    </el-collapse>
  </el-card>
</template>

<script>
import Axios from 'axios'
import vueJsonEditor from 'vue-json-editor'

import constants from '@/constants'

const {
  API_URL
} = constants

export default {
  name: 'debug-console',
  props: ['flag'],
  data () {
    return {
      evalContext: {
        entityID: 'a1234',
        entityType: 'report',
        entityContext: {
          hello: 'world'
        },
        enableDebug: true,
        flagID: this.flag.id,
        flagKey: this.flag.key
      },
      evalResult: {},
      batchEvalContext: {
        entities: [
          {
            entityID: 'a1234',
            entityType: 'report',
            entityContext: {
              hello: 'world'
            }
          },
          {
            entityID: 'a5678',
            entityType: 'report',
            entityContext: {
              hello: 'world'
            }
          }
        ],
        enableDebug: true,
        flagIDs: [
          this.flag.id
        ]
      },
      batchEvalResult: {}
    }
  },
  methods: {
    postEvaluation (evalContext) {
      Axios.post(`${API_URL}/evaluation`, evalContext).then((response) => {
        this.$message.success(`evaluation success`)
        this.evalResult = response.data
      }, () => { this.$message.error(`evaluation error`) })
    },
    postEvaluationBatch (batchEvalContext) {
      Axios.post(`${API_URL}/evaluation/batch`, batchEvalContext).then((response) => {
        this.$message.success(`evaluation success`)
        this.batchEvalResult = response.data
      }, () => { this.$message.error(`evaluation error`) })
    }
  },
  components: {
    vueJsonEditor
  },
  mounted () {
    this.$refs.evalContextEditor.editor.setMode('code')
    this.$refs.evalResultEditor.editor.setMode('code')
    this.$refs.batchEvalContextEditor.editor.setMode('code')
    this.$refs.batchEvalResultEditor.editor.setMode('code')
  }
}
</script>

<style lang="less">
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
