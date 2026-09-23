<template>
  <json-editor
    :json-string="text"
    :main-menu-bar="false"
    :navigation-bar="false"
    :status-bar="false"
    mode="text"
    class="json-text-editor"
    @update:json-string="onText"
  />
</template>

<script lang="ts">
import type { PropType } from 'vue'
import JsonEditor from 'vue3-ts-jsoneditor'
import { editorTextFromValue, toJsonText } from '@/helpers/evaluation'

/**
 * Text-mode JSON editor bound to the `json-string` (text) prop. Driving the
 * editor by value (`:json`) while in `mode="text"` makes vue3-ts-jsoneditor
 * re-set the document on every prop change — including the echo of the user's
 * own keystrokes — which resets the caret. Binding text and ignoring that echo
 * keeps the caret where the user is typing.
 */
export default {
  name: 'JsonTextEditor',
  components: { JsonEditor },
  props: {
    modelValue: { type: Object as PropType<unknown>, required: true },
    /** Narrows the raw editor text back to the value's type; returns null when invalid. */
    parser: {
      type: Function as PropType<(text: string) => unknown>,
      required: true,
    },
  },
  emits: ['update:modelValue'],
  data() {
    return {
      text: toJsonText(this.modelValue),
      // Last text we pushed to the parent, so the editor can ignore its own
      // echo instead of resetting the caret on every keystroke.
      emitted: toJsonText(this.modelValue),
    }
  },
  watch: {
    modelValue: {
      handler(value: unknown) {
        const text = editorTextFromValue(value, this.emitted)
        if (text === null) return
        this.emitted = text
        this.text = text
      },
      deep: true,
    },
  },
  methods: {
    onText(text: string) {
      this.text = text
      const parsed = this.parser(text)
      if (parsed == null) return
      this.emitted = toJsonText(parsed)
      this.$emit('update:modelValue', parsed)
    },
  },
}
</script>

<style scoped>
.json-text-editor {
  width: 100%;
}
</style>
