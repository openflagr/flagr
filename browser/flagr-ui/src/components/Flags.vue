<template>
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
      <el-alert
        v-if="createSuccess"
        title="Feature flag created!"
        type="success"
        show-icon>
      </el-alert>
      <h2>
        Feature Flags
        <span v-if="flags.length">({{ flags.length }})</span>
      </h2>
      <ul v-if="flags.length">
        <li
          v-for="flag in flags" class="flag"
          :class="{new: flag._new}">
          <router-link
            class="flag-link flex-row"
            :to="{name: 'flag', params: {flagId: flag.id}}">
            <div class="flex-row-left">
              <el-tag>{{ flag.id }}</el-tag> {{ flag.description }}
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
            placeholder="Feature flag description"
            v-model="newFlag.description">  
          </el-input>
        </p>
        <el-button
          class="width--full"
          :disabled="!newFlag.description"
          @click.prevent="createFlag">
          Create Feature Flag
        </el-button>
      </div>
    </div>
  </div>
</template>

<script>
import constants from '@/constants'
import fetchHelpers from '@/helpers/fetch'
import Spinner from '@/components/Spinner'
import { Tag, Button, Input, Alert, Breadcrumb, BreadcrumbItem } from 'element-ui'

const {
  getJson,
  postJson
} = fetchHelpers

const {
  API_URL
} = constants

export default {
  name: 'flags',
  components: {
    spinner: Spinner,
    'el-tag': Tag,
    'el-input': Input,
    'el-button': Button,
    'el-alert': Alert,
    'el-breadcrumb': Breadcrumb,
    'el-breadcrumb-item': BreadcrumbItem
  },
  data () {
    return {
      loaded: false,
      loadError: false,
      createSuccess: false,
      flags: [],
      newFlag: {
        description: ''
      }
    }
  },
  created () {
    getJson(`${API_URL}/flags`)
      .then(flags => {
        this.loaded = true
        flags.reverse()
        this.flags = flags
      })
      .catch(() => {
        this.loadError = true
      })
  },
  methods: {
    createFlag () {
      if (!this.newFlag.description) {
        return
      }

      postJson(`${API_URL}/flags`, this.newFlag)
        .then(flag => {
          this.newFlag.description = ''
          this.createSuccess = true

          flag._new = true
          this.flags.unshift(flag)
        })
    }
  }
}
</script>

<style lang="less" scoped>

.flags-container {
  width: 700px;
  ul {
    border-radius: 3px;
    border: 1px solid #ddd;
    background-color: white;
    overflow-y: scroll;
    overflow-x: hidden;
    max-height: 800px;
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
