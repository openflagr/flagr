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

/** Stable text form used as the editor's source of truth. */
function toJsonText(value: unknown): string {
  return JSON.stringify(value, null, 2)
}

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
      // Last value we pushed to the parent, so the editor can ignore its own
      // echo instead of resetting the caret on every keystroke.
      emitted: toJsonText(this.modelValue),
    }
  },
  watch: {
    modelValue: {
      handler(value: unknown) {
        const text = toJsonText(value)
        if (text === this.emitted) return
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
