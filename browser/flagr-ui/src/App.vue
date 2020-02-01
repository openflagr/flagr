<template>
  <div id="app">
    <el-menu mode="horizontal" class="navbar">
      <el-row>
        <el-col :span="6">
          <router-link :to="{name: 'home'}">
            <div class="logo-container">
              <h3 class="logo">Flagr</h3>
              <div>
                <span class="version">v{{ version }}</span>
              </div>
            </div>
          </router-link>
        </el-col>
        <el-col :span="2" :offset="13">
          <a href="https://checkr.github.io/flagr/api_docs" target="_blank"><h3>API</h3></a>
        </el-col>
        <el-col :span="2">
          <a href="https://checkr.github.io/flagr" target="_blank"><h3>Docs</h3></a>
        </el-col>
      </el-row>
    </el-menu>
    <div class="router-view-container">
      <router-view></router-view>
    </div>
  </div>
</template>

<script>
const version = require("../package.json").version || "1.0.0"
export default {
  name: 'app', 
  data: () => ({
    version
  }),
}
</script>

<style lang="less">
body {
  margin: 0;
  padding: 0;
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

h1, h2 {
  font-weight: normal;
}

ol {
  margin: 0;
  padding-left: 20px;
}

ul {
  list-style-type: none;
  margin: 0;
  padding: 0;
  li {
    margin: 0;
    padding: 0;
  }
}

.width--full {
  width: 100%;
}

#app {
  color: #2c3e50;

  span[size="small"] {
    font-size: 0.85em
  }

  .navbar {
    background-color: #74E5E0;
    color: white;
    .logo-container {
      display: flex;
      justify-content: center;
      align-items: center;
      font-weight: bold;

      h3 {
        margin-left: 2em;
        margin-right: 10px;
        &:hover {
          color: #f7f7f7;
        }
      }

      span {
        font-size: 12px;
      }
    }
    a {
      color: white;
      text-decoration: none;
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

  .segment-rollout-percent input {
    text-align: right;
  }

  .el-card {
    .el-card__header {
      background-color: #74E5E0;
      h2 {
        margin: -0.2em;
        color: white;
      }
    }
    margin-bottom: 1em;
  }

  .jsoneditor{
    border-color: #E4E7ED;
    .jsoneditor-menu {
      background-color: #E4E7ED;
      border-bottom-color: #E4E7ED;
    }
    .jsoneditor-poweredBy {
      display: none;
    }
  }
}
</style>
