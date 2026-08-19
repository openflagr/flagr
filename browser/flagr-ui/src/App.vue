<template>
  <div id="app">
    <header class="navbar">
      <div class="navbar-inner">
        <router-link
          :to="{ name: 'home' }"
          class="navbar-brand"
        >
          <h3 class="logo">
            Flagr <span class="version">v{{ version }}</span>
          </h3>
        </router-link>
        <nav class="navbar-nav">
          <a
            href="https://openflagr.github.io/flagr/api_docs"
            target="_blank"
          >API</a>
          <a
            href="https://openflagr.github.io/flagr"
            target="_blank"
          >Docs</a>
        </nav>
      </div>
    </header>
    <div class="router-view-container">
      <el-alert
        v-if="evalOnlyMode"
        class="readonly-banner"
        type="warning"
        :closable="false"
        show-icon
        title="Read-only (GitOps) mode"
        description="Flags are managed via the JSON source. Browsing and evaluation are available; editing is disabled."
        data-testid="readonly-banner"
      />
      <router-view />
    </div>
  </div>
</template>

<script lang="ts">
import pkg from '../package.json'
import { evalOnlyMode } from '@/helpers/serverMode'

const version = pkg.version || '1.0.0'

// The server mode is resolved in main.ts before mount, so the first paint
// already knows whether to render the read-only banner.
export default {
  name: 'App',
  setup() {
    return { evalOnlyMode }
  },
  data: () => ({ version }),
}
</script>

<style lang="scss">
@use './styles/app.scss';

.readonly-banner {
  margin-top: var(--space-sm);
}
</style>
