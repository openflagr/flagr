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
import vueJsonEditor from 'vue-json-editor'
import constants from '@/constants'

const {
  API_URL
} = constants

export default {
  name: 'debug-console',
  props: ['flagId'],
  data () {
    return {
      evalContext: {
        entityID: 'a1234',
        entityType: 'report',
        entityContext: {
          dl_state: 'CA',
          package: 'flagr-test'
        },
        enableDebug: true,
        flagID: this.$props.flagId
      },
      evalResult: {},
      batchEvalContext: {
        entities: [
          {
            entityID: 'a1234',
            entityType: 'report',
            entityContext: {
              dl_state: 'CA',
              package: 'flagr-test'
            }
          },
          {
            entityID: 'a5678',
            entityType: 'report',
            entityContext: {
              dl_state: 'NY',
              package: 'flagr-test'
            }
          }
        ],
        enableDebug: true,
        flagIDs: [
          this.$props.flagId
        ]
      },
      batchEvalResult: {}
    }
  },
  methods: {
    postEvaluation (evalContext) {
      this.$http.post(`${API_URL}/evaluation`, evalContext).then((response) => {
        this.$message.success(`evaluation success`)
        this.evalResult = response.body
      }, () => {
        this.$message.error(`evaluation error`)
      })
    },
    postEvaluationBatch (batchEvalContext) {
      this.$http.post(`${API_URL}/evaluation/batch`, batchEvalContext).then((response) => {
        this.$message.success(`evaluation success`)
        this.batchEvalResult = response.body
      }, () => {
        this.$message.error(`evaluation error`)
      })
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
