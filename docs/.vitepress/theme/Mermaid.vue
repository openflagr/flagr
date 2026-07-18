<script setup lang="ts">
import { onMounted, ref } from "vue";
import mermaid from "mermaid";

mermaid.initialize({ startOnLoad: false, securityLevel: "strict" });

const props = defineProps<{ code: string }>();
const host = ref<HTMLElement | null>(null);

onMounted(async () => {
  if (!host.value || !props.code.trim()) {
    return;
  }
  try {
    const id = `mmd-${Math.random().toString(36).slice(2)}`;
    const { svg } = await mermaid.render(id, props.code);
    host.value.innerHTML = svg;
  } catch (err) {
    host.value.textContent = props.code;
    console.error("[mermaid]", err);
  }
});
</script>

<template>
  <div ref="host" class="vp-mermaid" />
</template>
