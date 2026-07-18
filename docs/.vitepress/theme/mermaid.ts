import mermaid from "mermaid";
import { nextTick, onMounted, watch } from "vue";
import { useRoute } from "vitepress";

let mermaidReady = false;

function ensureMermaid() {
  if (mermaidReady) {
    return;
  }
  mermaid.initialize({
    startOnLoad: false,
    securityLevel: "strict",
  });
  mermaidReady = true;
}

/** Render authored ```mermaid``` fences after each page navigation. */
export function useMermaid() {
  const route = useRoute();

  const render = async () => {
    await nextTick();
    if (typeof document === "undefined") {
      return;
    }
    // Skip nodes mermaid already replaced (svg children).
    const nodes = Array.from(
      document.querySelectorAll<HTMLElement>("div.mermaid"),
    ).filter((el) => !el.querySelector("svg"));
    if (!nodes.length) {
      return;
    }
    ensureMermaid();
    await mermaid.run({ nodes });
  };

  onMounted(() => {
    void render();
  });
  watch(
    () => route.path,
    () => {
      void render();
    },
  );
}
