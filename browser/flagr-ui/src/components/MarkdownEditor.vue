<template>
  <div id="editor" v-if="showEditor || markdown">
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/4.0.0/github-markdown.min.css"
    />
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/KaTeX/0.11.1/katex.min.css"
    />

    <el-row :gutter="10">
      <el-col :span="12" v-if="showEditor">
        <el-input
          type="textarea"
          :rows="12"
          placeholder="Please input"
          v-model="input"
          @change="syncMarkdown"
        ></el-input>
      </el-col>
      <el-col :span="showEditor ? 12 : 24">
        <div class="markdown-body" v-html="compiledMarkdown"></div>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import MarkdownIt from "markdown-it";
import mk from "@iktakahiro/markdown-it-katex";
import xss from "xss";

let md = MarkdownIt("commonmark");
md.use(mk);

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
      return md.render(xss(this.input));
    },
  },
  methods: {
    syncMarkdown(md) {
      this.$emit("update:markdown", md);
    },
  },
  watch: {},
  mounted() {},
};
</script>

<style lang="less" scoped>
.markdown-body {
  background-color: #f6f8fa;
  padding: 0.5rem;
}
</style>

