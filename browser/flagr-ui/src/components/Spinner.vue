<template>
  <div class="spinner" />
</template>

<script lang="ts">
export default {
  name: 'Spinner',
}
</script>

<style scoped>
.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid var(--el-border-color-lighter);
  border-top-color: var(--el-color-primary);
  border-radius: 50%;
  margin: 80px auto;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
