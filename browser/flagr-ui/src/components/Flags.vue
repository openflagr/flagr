<template>
  <el-row>
    <el-col :span="14" :offset="5">
      <div class="flags-container container">
        <el-breadcrumb separator="/" v-if="loaded && !loadError">
          <el-breadcrumb-item>Home page</el-breadcrumb-item>
        </el-breadcrumb>

        <spinner v-if="!loaded" />

        <div v-if="loadError" class="card--error">
          <span class="el-icon-circle-close"></span>
          Failed to load feature flags
        </div>

        <div v-if="loaded && !loadError">
          <ul v-if="flags.length">
            <li
              v-for="flag in flags" class="flag"
              :class="{new: flag._new}">
              <router-link
                class="flag-link flex-row"
                :to="{name: 'flag', params: {flagId: flag.id}}">
                <div class="flex-row-left">
                  <el-tag type="primary" :disable-transitions="true">Flag ID: {{ flag.id }}</el-tag> {{ flag.name }}
                </div>
                <div class="flex-row-right">
                  <span :class="{'flag-enabled-icon': true, enabled: flag.enabled}"></span>
                </div>
              </router-link>
            </li>
          </ul>
          <div class="card--empty" v-else>
            No feature flags created yet
          </div>
          <div>
            <p>
              <el-input
                placeholder="New flag name (must be unique)"
                v-model="newFlag.name">
                <template slot="prepend">
                  Name
                </template>
                <template slot="append">
                  <el-button
                    :disabled="!(newFlag.name && newFlag.description)"
                    @click.prevent="createFlag">
                    <span class="el-icon-plus"/> Create Flag
                  </el-button>
                </template>
              </el-input>
              <el-input
                placeholder="New flag description"
                v-model="newFlag.description">
                <template slot="prepend">
                  Description
                </template>
              </el-input>
            </p>
          </div>
        </div>
      </div>
    </el-col>
  </el-row>
</template>

<script>
import constants from '@/constants'
import Spinner from '@/components/Spinner'

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
      loadError: false,
      flags: [],
      newFlag: {
        name: '',
        description: ''
      }
    }
  },
  created () {
    this.$http.get(`${API_URL}/flags`)
      .then(response => {
        let flags = response.body
        this.loaded = true

        // Sort flags by name instead of by ID
        flags.sort((flag1, flag2) => {
          return flag1.name.localeCompare(flag2.name)
        })
        this.flags = flags
      }, (err) => {
        this.$message.error(err.body.message)
        this.loadError = true
      })
  },
  methods: {
    createFlag () {
      if (!this.newFlag.name || !this.newFlag.description) {
        return
      }

      this.$http.post(`${API_URL}/flags`, this.newFlag)
        .then(response => {
          let flag = response.body
          this.newFlag.description = ''
          this.newFlag.name = ''
          this.$message('flag created')

          flag._new = true
          this.flags.unshift(flag)
        }, err => {
          this.$message.error(err.body.message)
        })
    }
  }
}
</script>

<style lang="less" scoped>

.flags-container {
  ul {
    border-radius: 3px;
    border: 1px solid #ddd;
    background-color: white;
    overflow-x: hidden;
    list-style-type: none;
    padding: 0;
    li.flag {
      text-align: left;
      display: block;
      padding: 0;
      border-bottom: 1px solid #ccc;
      &.new {
        background-color: #13ce66;
        .flag-link {
          color: white;
        }
      }
      .flag-link {
        display: inline-block;
        box-sizing: border-box;
        padding: 8px 12px;
        font-size: 1.1em;
        width: 100%;
        text-decoration: none;
        color: #2c3e50;
        &:hover {
          background-color: #74E5E0;
          color: white;
        }
      }
      .flag-enabled-icon {
        display: inline-block;
        width: 10px;
        height: 10px;
        border-radius: 5px;
        background-color: #ff4949;
        &.enabled {
          background-color: #13ce66;
        }
      }
    }
  }
}

</style>
