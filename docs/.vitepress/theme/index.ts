import DefaultTheme from "vitepress/theme";
import type { Theme } from "vitepress";
import "./custom.css";
import { useMermaid } from "./mermaid";

const theme: Theme = {
  extends: DefaultTheme,
  setup() {
    useMermaid();
  },
};

export default theme;
