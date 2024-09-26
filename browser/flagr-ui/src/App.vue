<template>
  <div id="app">
    <el-menu class="navbar" v-if="showHeader">
      <el-row>
        <el-col :span="20" :offset="2">
          <el-row>
            <el-col :span="6">
              <RouterLink :to="{ name: 'home' }">
                <div class="logo-container">
                  <h3 class="logo">Flagr</h3>
                  <div>
                    <span class="version">v{{ version }}</span>
                  </div>
                </div>
              </RouterLink>
            </el-col>
            <el-col :span="2" :offset="12">
              <a href="https://openflagr.github.io/flagr/api_docs" target="_blank"
                ><h3>API</h3></a
              >
            </el-col>
            <el-col :span="1" :offset="1">
              <a href="https://openflagr.github.io/flagr" target="_blank"
                ><h3>Docs</h3></a
              >
            </el-col>
            <el-col :span="1" :offset="1" v-if="authenticated">
              <span href="https://openflagr.github.io/flagr" target="_blank"
              @click="logoutUser"
              ><h3>Logout</h3></span>
            </el-col>
          </el-row>
        </el-col>
      </el-row>
    </el-menu>
    <div class="router-view-container">
      <RouterView/>
    </div>
  </div>
</template>

<script>
import { RouterLink, RouterView } from "vue-router";
import {logout} from './utils/apiUtil';
import { mapState } from "vuex";
const version = require("../package.json").version || "1.0.0";
export default {
  name: "app",
  data: () => ({
    version,
    showHeader: false,
    authenticated: false
  }),
  components: {
    RouterView,
    RouterLink
  },
  computed: {
      ...mapState({
          isAuth: (state) => state.isAuth, // Map userDetails from Vuex state
      }),
    },
  mounted(){
    this.checkAuthentication();
  },
  methods: {
    checkAuthentication() {
      const isAuthenticated = !!localStorage.getItem('tokens');
      console.log('isAuthenticated', isAuthenticated)
      if (!isAuthenticated || !this.isAuth) {
        this.showHeader = false;
      } else {
        this.showHeader = true;
        this.authenticated = true;
      }
    },
    logoutUser() {
      console.log('logging out')
      logout();
      this.$router.push('/login');
    }
  },
  watch: {
    // Watch for route changes to recheck authentication on navigation
    isAuth() {
      console.log("isauth changed")
      this.checkAuthentication();
    },
    $route() {
      this.checkAuthentication();
    }
  }
};
</script>

<style lang="less">
body {
  margin: 0;
  padding: 0;
  font-family: -apple-system,BlinkMacSystemFont,Segoe UI,Helvetica,Arial,sans-serif,Apple Color Emoji,Segoe UI Emoji;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

h1,
h2 {
  font-weight: normal;
}

ol {
  margin: 0;
  padding-left: 20px;
}

.width--full {
  width: 100%;
}

#app {
  color: #2c3e50;

  .flex {
    display: flex;
  }

  .flex-col {
    flex-direction: column;
  }

  .justify-center {
    justify-content: center;
  }

  .items-center {
    align-items: center;
  }
  span[size="small"] {
    font-size: 0.85em;
  }

  .navbar {
    background-color: #74e5e0;
    color: #2e4960;
    border: 0;

    .logo-container {
      display: flex;
      align-items: center;
      font-weight: bold;

      h3 {
        margin-right: 10px;
        &:hover {
          color: #000;
        }
      }

      span {
        font-size: 12px;
      }
    }

    a {
      color: inherit;
      text-decoration: none;
    }

    .el-col {
      text-align: right;

      &:first-child {
        text-align: left;
      }
    }
  }

  .flex-row {
    display: flex;
    align-items: center;
    justify-content: center;
    &-right {
      margin-left: auto;
    }
    &.equal-width {
      > * {
        flex: 1;
      }
    }
    &.align-items-top {
      align-items: flex-start;
    }
  }

  .container {
    margin: 0 auto;
    margin-top: 20px;
  }

  img {
    height: 60px;
  }

  .card {
    &--error {
      box-sizing: border-box;
      background-color: #fff9f9;
      padding: 10px;
      text-align: center;
      color: #ed2d2d;
      border: 1px solid #ed2d2d;
      border-radius: 3px;
      width: 100%;
      margin-bottom: 12px;
    }
    &--empty {
      box-sizing: border-box;
      background-color: #eee;
      padding: 10px;
      text-align: center;
      color: #777;
      border: 1px solid #ccc;
      border-radius: 3px;
      width: 100%;
      margin-bottom: 12px;
    }
  }

  .el-breadcrumb {
    margin-bottom: 20px;
  }

  .el-input {
    margin-bottom: 2px;
  }
  .el-dropdown .el-button-group {
    display: block;
  }
  .segment-rollout-percent input {
    text-align: right;
  }

  .el-card {
    .el-card__header {
      background-color: #74e5e0;
      color: #2e4960;
      border: 0;

      h2 {
        margin: -0.2em;
        color: inherit;
        font-size: 20px;
      }
    }
    margin-bottom: 1em;
  }

  .jsoneditor {
    border-color: #e4e7ed;
    .jsoneditor-menu {
      background-color: #e4e7ed;
      border-bottom-color: #e4e7ed;
    }
    .jsoneditor-poweredBy {
      display: none;
    }
  }

  .el-tag {
    margin: 2.5px;
  }
}
</style>
