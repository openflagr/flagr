<script setup lang="ts">
import { onMounted, ref } from "vue";
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
  const next = Math.round((zoom.value + delta) / ZOOM_STEP) * ZOOM_STEP;
  zoom.value = Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, next));
}
</script>

<template>
  <div class="vp-mermaid-wrap">
    <div ref="host" class="vp-mermaid" />
    <button
      v-if="svgHtml"
      type="button"
      class="vp-mermaid-expand"
      @click="openViewer"
    >
      Enlarge
    </button>

    <dialog
      ref="dialogEl"
      class="vp-mermaid-dialog"
      aria-label="Enlarged diagram"
      @click="onDialogClick"
    >
      <div class="vp-mermaid-dialog-bar">
        <div class="vp-mermaid-dialog-zoom">
          <button
            type="button"
            class="vp-mermaid-dialog-btn"
            aria-label="Zoom out"
            :disabled="zoom <= ZOOM_MIN"
            @click="zoomBy(-ZOOM_STEP)"
          >
            −
          </button>
          <span class="vp-mermaid-dialog-zoom-label" aria-live="polite">
            {{ Math.round(zoom * 100) }}%
          </span>
          <button
            type="button"
            class="vp-mermaid-dialog-btn"
            aria-label="Zoom in"
            :disabled="zoom >= ZOOM_MAX"
            @click="zoomBy(ZOOM_STEP)"
          >
            +
          </button>
        </div>
        <button
          type="button"
          class="vp-mermaid-dialog-btn"
          @click="closeViewer"
        >
          Close
        </button>
      </div>
      <div class="vp-mermaid-dialog-body">
        <div
          class="vp-mermaid-dialog-canvas"
          :style="{ zoom }"
          v-html="svgHtml"
        />
      </div>
    </dialog>
  </div>
</template>
