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

<style lang="scss">
@use './styles/app.scss';
</style>
