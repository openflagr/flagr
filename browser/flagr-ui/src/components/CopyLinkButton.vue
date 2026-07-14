<template>
  <el-tooltip
    :content="copied ? copiedTooltip : tooltip"
    placement="top"
    effect="light"
    :show-after="tooltipShowAfter"
    :hide-after="0"
  >
    <el-button
      class="copy-link-btn"
      :class="{ 'is-copied': copied }"
      :data-state="copied ? 'copied' : undefined"
      size="small"
      link
      type="primary"
      :aria-label="copied ? copiedAriaLabel : ariaLabel"
      :data-testid="testId"
      :disabled="disabled || !url"
      @click.stop.prevent="onCopy"
    >
      <el-icon>
        <Check v-if="copied" />
        <CopyDocument v-else />
      </el-icon>
      <span
        class="copy-link-btn__sr"
        aria-live="polite"
      >{{ copied ? 'Copied' : '' }}</span>
    </el-button>
  </el-tooltip>
</template>

<script lang="ts">
import { Check, CopyDocument } from '@element-plus/icons-vue'
import { COPY_FEEDBACK_MS, copyText } from '@/helpers/copyText'

const COPY_ERROR_MESSAGE = "Couldn't copy — select and copy manually"

export default {
  name: 'CopyLinkButton',
  components: { Check, CopyDocument },
  props: {
    /** Absolute URL to copy. */
    url: { type: String, required: true },
    /** Accessible name (and tooltip) in the default state. */
    ariaLabel: { type: String, default: 'Copy link' },
    tooltip: { type: String, default: 'Copy link' },
    copiedTooltip: { type: String, default: 'Copied' },
    copiedAriaLabel: { type: String, default: 'Copied' },
    testId: { type: String, default: 'copy-link-btn' },
    disabled: { type: Boolean, default: false },
    /** Hover tooltip delay (ms). Focus tooltips still open immediately via Element Plus. */
    tooltipShowAfter: { type: Number, default: 800 },
  },
  data() {
    return {
      copied: false,
      revertTimer: 0 as number | ReturnType<typeof setTimeout>,
    }
  },
  beforeUnmount() {
    this.clearRevertTimer()
  },
  methods: {
    clearRevertTimer() {
      if (this.revertTimer) {
        clearTimeout(this.revertTimer)
        this.revertTimer = 0
      }
    },
    async onCopy() {
      if (!this.url || this.disabled) return
      const ok = await copyText(this.url)
      if (!ok) {
        this.$message.error(COPY_ERROR_MESSAGE)
        return
      }
      this.copied = true
      this.clearRevertTimer()
      this.revertTimer = setTimeout(() => {
        this.copied = false
        this.revertTimer = 0
      }, COPY_FEEDBACK_MS)
    },
  },
}
</script>

<style lang="scss" scoped>
/* Hallmark · component: copy-link-button · genre: modern-minimal · theme: inherit-flagr
 * states: default · hover · focus · active · disabled · loading · error · success
 * contrast: pass (Element Plus primary + text tokens)
 */
.copy-link-btn {
  vertical-align: middle;
  padding: var(--space-3xs);
  min-width: 28px;
  min-height: 28px;
  color: var(--el-text-color-secondary);

  &:hover {
    color: var(--el-color-primary);
  }

  &:active {
    transform: translateY(1px);
  }

  &.is-copied {
    color: var(--el-color-success);
  }

  &:disabled {
    color: var(--el-text-color-placeholder);
    transform: none;
  }
}

.copy-link-btn__sr {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (prefers-reduced-motion: reduce) {
  .copy-link-btn:active {
    transform: none;
  }
}
</style>
