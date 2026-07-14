<template>
  <div id="app">
    <header class="navbar">
      <div class="navbar-inner">
        <router-link
          :to="{ name: 'home' }"
          class="navbar-brand"
        >
          <h3 class="logo">
            <svg
              class="logo-mark"
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 96 96"
              width="28"
              height="28"
              aria-hidden="true"
              focusable="false"
            >
              <path
                fill="currentColor"
                d="M28 12 L50 44 L50 50 L38 54 L50 58 L50 64 L28 84 L20 64 L40 54 L20 44 Z"
              />
              <path
                fill="currentColor"
                d="M50 28 H82 L94 42 H62 L50 46 Z"
              />
              <path
                fill="currentColor"
                d="M50 52 H72 L80 64 H56 L50 58 Z"
              />
            </svg>
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
      <router-view />
    </div>
  </div>
</template>

<script lang="ts">
import pkg from '../package.json'

const version = pkg.version || '1.0.0'

export default {
  name: 'App',
  data: () => ({ version }),
}
</script>

<style lang="scss">
@use './styles/app.scss';
</style>
