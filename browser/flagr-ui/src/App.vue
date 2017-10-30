<template>
  <div id="app">
    <el-menu mode="horizontal" class="navbar">
      <el-row>
        <el-col :span="6">
          <div class="img-frame">
            <a href="#"><img src="./assets/logo.png"></a>
          </div>
        </el-col>
        <el-col :span="3" :offset="12">
          <a href="/api_docs" target="_blank"><h3>API</h3></a>
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
export default {
  name: 'app'
}
</script>

<style lang="less">
body {
  background-color: #f7f7f7;
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
    .img-frame {
      height: 80px;
      width: 150px;
      position: relative;
    }
    img {
      max-height: 100%;
      max-width: 100%;
      width: auto;
      height: auto;
      position: absolute;
      top: 0;
      bottom: 0;
      left: 0;
      right: 0;
      margin: auto;
    }
    h3 {
      color: white;
      text-align: right;
      margin-right: 1em;
      &:hover {
        color: #f7f7f7;
      }
    }
    a {
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
      background-color: white;
      padding: 10px;
      text-align: center;
      color: #ed2d2d;
      border: 1px solid #ed2d2d;
      border-radius: 3px;
      width: 100%;
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
    }
    margin-bottom: 1em;
    h2 {
      margin: -0.2em;
      color: white;
    }
  }

  .jsoneditor{
    border-color: #d8dce5;
    .jsoneditor-menu {
      background-color: #74E5E0;
      border-bottom-color: #74E5E0;
    }
    .jsoneditor-poweredBy {
      display: none;
    }
  }
}
</style>
