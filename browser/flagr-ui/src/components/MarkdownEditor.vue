<template>
    <div id="editor" v-if="hideMdBox">
        <mavon-editor
          language="en"
          v-bind="editorSettings"
          :value="markdown"
          @save="save"
          @change="syncMarkdown"
        ></mavon-editor>
    </div>
</template>

<script>

import { mavonEditor } from 'mavon-editor'
import 'mavon-editor/dist/css/index.css'

export default {
  name: 'editor',
  props: {
    showEditor: Boolean,
    markdown: {
      type: String,
      default: ''
    }
  },
  components: {
    mavonEditor
  },
  data () {
    return {
      editorSettings: {
        toolbarsFlag: false,
        subfield: false,
        defaultOpen: 'preview'
      }
    }
  },
  computed: {
    hideMdBox: function () {
      return this.markdown.length > 0 || this.showEditor
    }
  },
  methods: {
    toggleEditor (show) {
      this.editorSettings.toolbarsFlag = show
      this.editorSettings.subfield = show
    },
    syncMarkdown (md, _) {
      this.$emit('update:markdown', md)
    },
    save () {
      this.$emit('save')
    }
  },
  watch: {
    showEditor: function () {
      this.toggleEditor(this.showEditor)
    }
  },
  mounted () {
    this.toggleEditor(this.showEditor)
  }
}
</script>
