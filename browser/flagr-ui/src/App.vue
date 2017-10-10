<template>
  <div id="app">
    <el-menu theme="light" mode="horizontal" class="navbar">
      <img src="./assets/logo.png">
    </el-menu>
    <div class="router-view-container">
      <router-view></router-view>
    </div>
  </div>
</template>

<script>
import {
  Menu,
  MenuItem
} from 'element-ui'

export default {
  name: 'app',
  components: {
    'el-menu': Menu,
    'el-menu-item': MenuItem
  }
}
</script>

<style lang="less">
body {
  background-color: #f7f7f7;
  margin: 0;
  padding: 0;
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
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;

  .navbar {
    background-color: #74E5E0;
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
    }
  }
}
</style>
