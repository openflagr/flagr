import DefaultTheme from "vitepress/theme";
import type { Theme } from "vitepress";
import "./custom.css";
import Mermaid from "./Mermaid.vue";

const theme: Theme = {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    app.component("Mermaid", Mermaid);
  },
};

export default theme;
