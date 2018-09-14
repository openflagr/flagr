<template>
  <el-row>
    <el-col :span="20" :offset="2">
      <div class="flags-container container">
        <el-breadcrumb separator="/" v-if="loaded">
          <el-breadcrumb-item>Home page</el-breadcrumb-item>
        </el-breadcrumb>

        <spinner v-if="!loaded" />

        <div v-if="loaded">
          <el-row>
            <el-col>
              <el-input
                placeholder="Specific new flag description"
                v-model="newFlag.description">
                <template slot="prepend">
                  <span class="el-icon-plus"/>
                </template>
                <template slot="append">
                  <el-button
                    type="primary"
                    :disabled="!newFlag.description"
                    @click.prevent="createFlag">
                    Create New Flag
                  </el-button>
                </template>
              </el-input>
            </el-col>
          </el-row>

          <el-table
            :data="flags"
            :stripe="true"
            :highlight-current-row="false"
            :default-sort="{prop: 'id', order: 'descending'}"
            v-on:row-click="goToFlag"
            style="width: 100%">
            <el-table-column
              prop="id"
              align="center"
              label="Flag ID"
              sortable
              fixed
              width="100">
            </el-table-column>
            <el-table-column
              prop="description"
              label="Description"
              min-width="380">
            </el-table-column>
            <el-table-column
              prop="updatedBy"
              label="Last Updated By"
              sortable
              width="200">
            </el-table-column>
            <el-table-column
              prop="updatedAt"
              label="Updated At (UTC)"
              :formatter="datetimeFormatter"
              sortable
              width="200">
            </el-table-column>
            <el-table-column
              prop="enabled"
              label="Enabled"
              sortable
              align="center"
              fixed="right"
              width="100">
              <template slot-scope="scope">
                <el-tag
                  :type="scope.row.enabled ? 'primary' : 'danger'"
                  disable-transitions>{{scope.row.enabled ? 'on' : 'off'}}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-col>
  </el-row>
</template>

<script>
import Axios from 'axios'

import constants from '@/constants'
import Spinner from '@/components/Spinner'
import helpers from '@/helpers/helpers'

const {
  handleErr
} = helpers

const {
  API_URL
} = constants

export default {
  name: 'flags',
  components: {
    spinner: Spinner
  },
  data () {
    return {
      loaded: false,
      flags: [],
      newFlag: {
        description: ''
      }
    }
  },
  created () {
    Axios.get(`${API_URL}/flags`)
      .then(response => {
        let flags = response.data
        this.loaded = true
        flags.reverse()
        this.flags = flags
      }, handleErr.bind(this))
  },
  methods: {
    flagEnabledFormatter (row, col, val) {
      return val ? 'on' : 'off'
    },
    datetimeFormatter (row, col, val) {
      return val.split('.')[0]
    },
    goToFlag (row) {
      this.$router.push({name: 'flag', params: {flagId: row.id}})
    },
    createFlag () {
      if (!this.newFlag.description) {
        return
      }

      Axios.post(`${API_URL}/flags`, this.newFlag)
        .then(response => {
          let flag = response.data
          this.newFlag.description = ''
          this.$message.success('flag created')

          flag._new = true
          this.flags.unshift(flag)
        }, handleErr.bind(this))
    }
  }
}
</script>

<style lang="less">

.flags-container {
  .el-table {
    margin-top: 2em;
  }
}

</style>
