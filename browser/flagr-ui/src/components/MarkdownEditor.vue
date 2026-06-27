<template>
  <div
    v-if="showEditor || markdown"
    id="editor"
  >
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/4.0.0/github-markdown.min.css"
    >

    <div
      v-if="showEditor"
      class="me-editor-section"
    >
      <el-input
        v-model="input"
        type="textarea"
        :rows="5"
        placeholder="Please input"
        @change="syncMarkdown"
      />
    </div>
    <!-- eslint-disable vue/no-v-html -- sanitized via xss() in compiledMarkdown -->
    <div
      class="markdown-body"
      v-html="compiledMarkdown"
    />
    <!-- eslint-enable vue/no-v-html -->
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
  name: "Editor",
  props: {
    showEditor: Boolean,
    markdown: {
      type: String,
      default: '',
    },
  },
  emits: ['update:markdown', 'save'],
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

