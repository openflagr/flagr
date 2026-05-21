<template>
  <div id="app">
    <header class="navbar">
      <div class="navbar-inner">
        <router-link :to="{ name: 'home' }" class="navbar-brand">
          <h3 class="logo">Flagr <span class="version">v{{ version }}</span></h3>
        </router-link>
        <nav class="navbar-nav">
          <a href="https://openflagr.github.io/flagr/api_docs" target="_blank">API</a>
          <a href="https://openflagr.github.io/flagr" target="_blank">Docs</a>
        </nav>
      </div>
    </header>
    <div class="router-view-container">
      <router-view></router-view>
    </div>
  </div>
</template>

<script>
import pkg from "../package.json";
const version = pkg.version || "1.0.0";
export default {
  name: "app",
  data: () => ({ version }),
};
</script>

<style lang="less">
body {
  margin: 0;
  padding: 0;
  background-color: var(--el-bg-color-page);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji";
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  line-height: 1.5;
}
h1, h2 { font-weight: 600; }
ol { margin: 0; padding-left: 20px; }
.width--full { width: 100%; }

#app {
  color: var(--el-text-color-primary);
  min-height: 100vh;
  background-color: var(--el-bg-color-page);

  span[size="small"] { font-size: 0.85em; }

  // --- Navbar ---
  .navbar {
    background-color: var(--el-color-primary);
    color: #fff;
    padding: 0;
  }
  .navbar-inner {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 24px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    min-height: 50px;
  }
  .navbar-brand {
    color: inherit;
    text-decoration: none;
    .logo { margin: 0; font-size: 1.2em; font-weight: 700; letter-spacing: -0.03em; }
    .version { font-size: 11px; font-weight: 400; opacity: 0.7; margin-left: 6px; }
  }
  .navbar-nav {
    display: flex;
    gap: 24px;
    a { color: inherit; text-decoration: none; font-size: 0.9em; font-weight: 500; opacity: 0.85; }
    a:hover { opacity: 1; }
  }
  .router-view-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 24px;
  }

  // --- Flex helpers ---
  .flex-row {
    display: flex;
    align-items: center;
    justify-content: center;
    &-right { margin-left: auto; display: flex; align-items: center; gap: 12px; }
    &.equal-width > * { flex: 1; }
    &.align-items-top { align-items: flex-start; }
  }

  .container { margin: 0 auto; margin-top: 24px; }
  img { height: 60px; }

  // --- Notion-style cards ---
  .el-card {
    border-radius: 12px;
    border: 1px solid var(--el-border-color);
    box-shadow: 0 1px 3px rgba(0,0,0,0.04);
    margin-bottom: 20px;

    .el-card__header {
      background: transparent;
      color: var(--el-text-color-primary);
      border-bottom: 1px solid var(--el-border-color-light);
      padding: 14px 20px;
      position: relative;
      &::before {
        content: '';
        position: absolute;
        left: 0;
        top: 0;
        bottom: 0;
        width: 3px;
        background: var(--el-color-primary);
        border-radius: 0 2px 2px 0;
      }
      h2 {
        margin: 0;
        color: var(--el-text-color-primary);
        font-size: 14px;
        font-weight: 600;
        letter-spacing: -0.01em;
      }
    }
    .el-card__body { padding: 16px 20px; }
  }

  // --- Nested inner card (hover-shadow) ---
  .el-card.is-hover-shadow {
    border: none;
    box-shadow: none;
    background: transparent;
    margin-bottom: 0;
    .el-card__header {
      display: none;
    }
    .el-card__body {
      padding: 0;
    }
  }

  // --- Status cards ---
  .card {
    &--error {
      box-sizing: border-box;
      background-color: var(--el-color-danger-light-9);
      padding: 10px 14px;
      text-align: center;
      color: var(--el-color-danger);
      border: 1px solid var(--el-color-danger-light-7);
      border-radius: 8px;
      width: 100%;
      margin-bottom: 12px;
      font-size: 0.9em;
    }
    &--empty {
      box-sizing: border-box;
      background-color: var(--el-fill-color-light);
      padding: 10px 14px;
      text-align: center;
      color: var(--el-text-color-placeholder);
      border: 1px dashed var(--el-border-color);
      border-radius: 8px;
      width: 100%;
      margin-bottom: 12px;
      font-size: 0.9em;
    }
  }

  .el-breadcrumb { margin-bottom: 20px; }

  .el-input { margin-bottom: 2px; }

  .segment-rollout-percent input { text-align: right; }

  .jsoneditor {
    border-color: var(--el-border-color);
    border-radius: 8px;
    .jsoneditor-menu {
      background-color: var(--el-fill-color-light);
      border-bottom-color: var(--el-border-color);
    }
    .jsoneditor-poweredBy { display: none; }
  }

  .el-tag {
    margin: 2px;
    border: none;
    font-weight: 500;
    &--warning {
      color: #b8860b;
      background-color: #fef7e7;
    }
    &--primary {
      color: var(--el-color-primary);
      background-color: var(--el-color-primary-light-9);
    }
    &--info {
      color: var(--el-text-color-secondary);
      background-color: var(--el-fill-color-light);
    }
  }
.segment-rollout-percent .el-input-group__prepend {
  color: var(--el-text-color-placeholder);
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.02em;
}

  .el-tabs {
    .el-tabs__header { margin-bottom: 20px; }
    .el-tabs__item { font-size: 14px; font-weight: 500; }
  }

  .el-dialog {
    border-radius: 12px;
    .el-dialog__header { margin-right: 0; padding: 20px 24px 0; font-weight: 600; }
    .el-dialog__body { padding: 12px 24px; }
    .el-dialog__footer { padding: 8px 24px 20px; }
  }
  .el-button--small { font-size: 12px; }
  .el-button.is-disabled {
    opacity: 0.45;
  }
  .el-button--primary.is-disabled {
    background-color: var(--el-button-disabled-bg-color, var(--el-fill-color));
    border-color: var(--el-button-disabled-border-color, var(--el-border-color-light));
  }

  // --- Reduced motion ---
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }

  // --- History / Snapshot page ---
  .snapshot-container {
    .el-card__body {
      padding: 16px 20px;
    }
    .diff-snapshot-id-change {
      display: flex;
      align-items: center;
      gap: 6px;
      .el-tag {
        color: var(--el-color-primary);
        background-color: var(--el-color-primary-light-9);
        border: none;
        font-weight: 500;
      }
    }
    .diff {
      margin: 0;
      font-size: 12px;
      line-height: 1.6;
      overflow-x: auto;
      del {
        background-color: #f7b3b3;
        text-decoration: none;
      }
      ins {
        background-color: #b6ddc6;
        text-decoration: none;
      }
    }
    .compact {
      font-size: 12px;
      color: var(--el-text-color-secondary);
    }
  }
}
</style>
