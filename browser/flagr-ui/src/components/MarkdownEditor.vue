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

<script setup>
import { ref, computed } from "vue";
import MarkdownIt from "markdown-it";
import mk from "@iktakahiro/markdown-it-katex";
import xss from "xss";

let md = MarkdownIt("commonmark");
md.use(mk);

const props = defineProps({
  showEditor: Boolean,
  markdown: {
    type: String,
    default: "",
  },
});

const emit = defineEmits(["update:markdown"]);

const input = ref(props.markdown);

const compiledMarkdown = computed(() => {
  return md.render(xss(input.value));
});

function syncMarkdown(val) {
  emit("update:markdown", val);
}
</script>

<style lang="less" scoped>
.markdown-body {
  background-color: #f6f8fa;
  padding: 0.5rem;
}
</style>
