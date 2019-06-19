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

          <el-row>
            <el-button
              type="primary"
              icon="el-icon-caret-bottom"
              v-on:click="exportFlags"
            >
              Export all flags
            </el-button>
            <a ref="exportFile"/>

            <el-button
              class="import-btn"
              icon="el-icon-upload2"
              @click="importFlags"
            >
              Import flags
            </el-button>
            <input
              type="file"
              hidden
              id="importFlags"
              @change="importFlagsChanged"
            />
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
    },
    async getFlag (flagId) {
      const { data } = await Axios.get(`${API_URL}/flags/${flagId}`)
      return data;
    },
    async getAllFlags () {
      const flagIds = this.flags.map(flag => flag.id)
      const getAllFlagsTasks = flagIds.map(this.getFlag)
      const flags = await Promise.all(getAllFlagsTasks)

      return flags
    },
    async exportFlags () {
      const exportFileHref = this.$refs.exportFile
      const flagsText = await this.getAllFlags()

      exportFileHref.setAttribute(
        "href",
        "data:text/plain;charset=utf-8," + encodeURIComponent(JSON.stringify(flagsText))
      );
      exportFileHref.setAttribute("download", 'flags.json');
      exportFileHref.click();
    },
    onFileReaderLoaded (loadedFile) {
      const fileContent = loadedFile.target.result;

      try {
        const flags = JSON.parse(fileContent);
        debugger
      }
      catch (err) {
        console.error('Failed to load flags from file', err)
      }
    },
    importFlags (e) {
      e.preventDefault()
      if (!window.FileReader) {
        console.info('Reading files is not supported for that browser. Please try using anthor one.')
        return;
      }

      document.getElementById('importFlags').click()
    },
    importFlagsChanged (e) {
      const file = e.target.files[0];

      if (file.type !== 'application/json') {
        console.info('Import flags supports only JSON format, please upload a new file')
        return;
      }

      const fileReader = new FileReader()
      fileReader.onload = this.onFileReaderLoaded
      fileReader.readAsText(file)
    }
  }
}
</script>

<style lang="less">

.flags-container {
  .el-table {
    margin-top: 2em;
  }
  .el-table__row {
    cursor: pointer;
  }
  .import-btn {
    margin-left: 10px;
  }
  // .import-flags-input {
  //   display: none;
  // }
}

</style>
