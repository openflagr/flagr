<template>
  <div id="editor" v-if="showEditor || markdown">
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/4.0.0/github-markdown.min.css"
    />

    <div v-if="showEditor" class="me-editor-section">
      <el-input
        type="textarea"
        :rows="5"
        placeholder="Please input"
        v-model="input"
        @change="syncMarkdown"
      ></el-input>
    </div>
    <div class="markdown-body" v-html="compiledMarkdown"></div>
  </div>
</template>

<script lang="ts">
import MarkdownIt from "markdown-it";
import mk from "@vscode/markdown-it-katex";
import xss from "xss";
import "katex/dist/katex.min.css"

let md = MarkdownIt("commonmark");
md.use(mk, { output: "html" });

export default {
  name: "editor",
  props: {
    showEditor: Boolean,
    markdown: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      input: this.markdown,
    };
  },
  computed: {
    compiledMarkdown() {
      return xss(md.render(this.input));
    },
  },
  watch: {
    markdown(val) {
      this.input = val
    }
  },
  methods: {
    syncMarkdown(markdown: string) {
      this.$emit('update:markdown', markdown)
    },
  }
};
</script>

<style lang="scss" scoped>
.markdown-body {
  background-color: var(--el-fill-color-light);
  padding: 0.5rem;
}
</style>

