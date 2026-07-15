<template>
  <div
    v-if="visible"
    class="spinner"
    role="status"
    aria-label="Loading"
  />
</template>

<script lang="ts">
/** Delay before showing to avoid flash on fast loads (Hallmark microinteraction). */
const SPINNER_SHOW_DELAY_MS = 150

export default {
  name: 'Spinner',
  props: {
    delayMs: {
      type: Number,
      default: SPINNER_SHOW_DELAY_MS,
    },
  },
  data() {
    return {
      visible: false,
      showTimer: null as ReturnType<typeof setTimeout> | null,
    }
  },
  mounted() {
    if (this.delayMs <= 0) {
      this.visible = true
      return
    }
    this.showTimer = setTimeout(() => {
      this.visible = true
      this.showTimer = null
    }, this.delayMs)
  },
  beforeUnmount() {
    if (this.showTimer != null) {
      clearTimeout(this.showTimer)
      this.showTimer = null
    }
  },
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
