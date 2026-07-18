<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import mermaid from "mermaid";

mermaid.initialize({ startOnLoad: false, securityLevel: "strict" });

const ZOOM_MIN = 0.5;
const ZOOM_MAX = 3;
const ZOOM_STEP = 0.25;

const props = defineProps<{ code: string }>();
const host = ref<HTMLElement | null>(null);
const dialogEl = ref<HTMLDialogElement | null>(null);
const svgHtml = ref("");
const zoom = ref(1);

/** Scale via width % so scroll area grows and every browser actually zooms. */
const canvasStyle = computed(() => ({
  width: `${zoom.value * 100}%`,
}));

onMounted(async () => {
  if (!host.value || !props.code.trim()) {
    return;
  }
  try {
    const id = `mmd-${Math.random().toString(36).slice(2)}`;
    const { svg } = await mermaid.render(id, props.code);
    svgHtml.value = svg;
    host.value.innerHTML = svg;
  } catch (err) {
    host.value.textContent = props.code;
    console.error("[mermaid]", err);
  }
});

function openViewer() {
  if (!svgHtml.value || !dialogEl.value) {
    return;
  }
  zoom.value = 1;
  dialogEl.value.showModal();
}

function closeViewer() {
  dialogEl.value?.close();
}

function onDialogClick(event: MouseEvent) {
  if (event.target === dialogEl.value) {
    closeViewer();
  }
}

function zoomBy(delta: number) {
  const next = Math.round((zoom.value + delta) * 100) / 100;
  zoom.value = Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, next));
}
</script>

<template>
  <div class="wrap">
    <div ref="host" class="diagram" />
    <button
      v-if="svgHtml"
      type="button"
      class="expand"
      @click="openViewer"
    >
      Enlarge
    </button>

    <dialog
      ref="dialogEl"
      class="viewer"
      aria-label="Enlarged diagram"
      @click="onDialogClick"
    >
      <div class="bar" @click.stop>
        <div class="zoom">
          <button
            type="button"
            class="btn"
            aria-label="Zoom out"
            :disabled="zoom <= ZOOM_MIN"
            @click="zoomBy(-ZOOM_STEP)"
          >
            −
          </button>
          <span class="zoom-label" aria-live="polite">
            {{ Math.round(zoom * 100) }}%
          </span>
          <button
            type="button"
            class="btn"
            aria-label="Zoom in"
            :disabled="zoom >= ZOOM_MAX"
            @click="zoomBy(ZOOM_STEP)"
          >
            +
          </button>
        </div>
        <button type="button" class="btn" @click="closeViewer">Close</button>
      </div>
      <div class="body">
        <div class="canvas" :style="canvasStyle" v-html="svgHtml" />
      </div>
    </dialog>
  </div>
</template>

<style scoped>
.wrap {
  position: relative;
  margin: 1rem 0;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.diagram {
  width: 100%;
  overflow-x: auto;
  text-align: center;
}

.diagram :deep(svg) {
  max-width: 100%;
  height: auto;
}

.expand {
  display: inline-flex;
  align-items: center;
  margin-top: 0.5rem;
  padding: 0.25rem 0.625rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  background: var(--vp-c-bg-soft);
  color: var(--vp-c-text-2);
  font-size: 0.8125rem;
  line-height: 1.4;
  cursor: pointer;
}

.expand:hover {
  color: var(--vp-c-text-1);
  border-color: var(--vp-c-brand-1);
}

.expand:focus-visible {
  outline: 2px solid var(--vp-c-brand-1);
  outline-offset: 2px;
}

.viewer {
  width: min(96vw, 1400px);
  height: min(92vh, 960px);
  max-width: 96vw;
  max-height: 92vh;
  margin: auto;
  padding: 0;
  border: 1px solid var(--vp-c-divider);
  border-radius: 10px;
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  box-shadow: var(--vp-shadow-3, 0 12px 40px rgba(0, 0, 0, 0.2));
}

.viewer::backdrop {
  background: rgba(0, 0, 0, 0.55);
}

.bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg-soft);
}

.zoom {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.zoom-label {
  min-width: 3.25rem;
  text-align: center;
  font-size: 0.8125rem;
  font-variant-numeric: tabular-nums;
  color: var(--vp-c-text-2);
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 2rem;
  min-height: 2rem;
  padding: 0.25rem 0.6rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  font-size: 0.875rem;
  line-height: 1;
  cursor: pointer;
}

.btn:hover:not(:disabled) {
  border-color: var(--vp-c-brand-1);
}

.btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.btn:focus-visible {
  outline: 2px solid var(--vp-c-brand-1);
  outline-offset: 2px;
}

.body {
  height: calc(100% - 3rem);
  overflow: auto;
  padding: 1rem;
  box-sizing: border-box;
}

.canvas {
  margin: 0 auto;
  text-align: center;
}

.canvas :deep(svg) {
  display: block;
  width: 100% !important;
  max-width: none !important;
  height: auto !important;
}
</style>
