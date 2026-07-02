<template>
  <el-card class="variants-container is-card-secondary">
    <template #header>
      <div class="el-card-header">
        <h2>Variants</h2>
      </div>
    </template>

    <div
      v-if="variants.length"
      class="variants-row"
    >
      <div
        v-for="variant in variants"
        :key="variant.id"
        class="variant-item ui-surface-inset"
      >
        <div class="variant-row">
          <span class="variant-id ui-id-badge">#{{ variant.id }}</span>
          <el-input
            size="small"
            placeholder="Variant Key"
            :model-value="variant.key"
            data-testid="variant-key-input"
            class="variant-key-field"
            @update:model-value="onVariantKeyInput(variant, $event)"
          />
          <div class="variant-actions">
            <el-tooltip
              :content="SAVE_DIRTY_TOOLTIP"
              placement="top"
              effect="light"
              :disabled="!isVariantDirty(variant)"
            >
              <el-button
                size="small"
                :plain="!isVariantDirty(variant)"
                :type="saveButtonType(isVariantDirty(variant))"
                data-testid="save-variant-btn"
                @click="handleSaveVariant(variant)"
              >
                {{ saveButtonLabel(isVariantDirty(variant)) }}
              </el-button>
            </el-tooltip>
            <el-button
              size="small"
              plain
              data-testid="delete-variant-btn"
              @click="$emit('delete-variant', variant)"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
        <el-collapse class="variant-attachment-collapse">
          <el-collapse-item
            title="Variant attachment"
            class="variant-attachment-collapsable-title"
          >
            <p class="variant-attachment-title">
              JSON in key/value pairs format.
            </p>
            <json-editor
              :json="variant.attachment"
              :main-menu-bar="false"
              :navigation-bar="false"
              :status-bar="false"
              mode="text"
              @update:json="onAttachmentChange(variant, $event, true)"
              @update:json-string="onAttachmentTextChange(variant, $event)"
              @error="onAttachmentChange(variant, null, false)"
            />
          </el-collapse-item>
        </el-collapse>
      </div>
    </div>
    <div
      v-else
      class="card--cue"
    >
      <p class="card--cue-title">
        No variants yet
      </p>
      <p class="card--cue-body">
        Variants are the outcomes a flag can return — e.g. <code>on</code>/<code>off</code>, <code>control</code>/<code>test</code>. Add one to start defining what entities receive.
      </p>
    </div>

    <div class="variant-add-row">
      <el-input
        v-model="newKey"
        size="small"
        placeholder="New Variant Key"
        data-testid="new-variant-input"
      />
      <el-button
        type="primary"
        size="small"
        :disabled="!newKey"
        data-testid="create-variant-btn"
        @click.prevent="createVariant"
      >
        Create Variant
      </el-button>
    </div>
  </el-card>
</template>


<script lang="ts">
import {
  SAVE_DIRTY_TOOLTIP,
  saveButtonLabel as fmtSaveLabel,
  saveButtonType as fmtSaveType,
} from '@/helpers/saveDirtyUi'
import JsonEditor from 'vue3-ts-jsoneditor'
import { Delete } from '@element-plus/icons-vue'
import type { PropType } from 'vue'
import type { Variant, VariantAttachment } from '@/api/types'

export default {
  name: 'VariantsSection',
  components: { JsonEditor, Delete },
  props: { variants: { type: Array as PropType<Variant[]>, required: true } },
  emits: [
    'update-variant-key',
    'save-variant',
    'delete-variant',
    'create-variant',
    'attachment-change',
  ],
  data() {
    return {
      SAVE_DIRTY_TOOLTIP,
      newKey: '',
      variantDirtyIds: {} as Record<number, boolean>,
    }
  },
  methods: {
    saveButtonLabel(dirty: boolean) {
      return fmtSaveLabel(dirty)
    },
    saveButtonType(dirty: boolean) {
      return fmtSaveType(dirty)
    },
    isVariantDirty(variant: Variant): boolean {
      return variant.id != null && !!this.variantDirtyIds[variant.id]
    },
    markVariantDirty(variant: Variant): void {
      if (variant.id != null) this.variantDirtyIds[variant.id] = true
    },
    clearVariantDirty(variant: Variant): void {
      if (variant.id != null) delete this.variantDirtyIds[variant.id]
    },
    handleSaveVariant(variant: Variant): void {
      this.$emit('save-variant', variant)
      this.clearVariantDirty(variant)
    },
    createVariant() {
      this.$emit('create-variant', { key: this.newKey })
      this.newKey = ''
    },
    onVariantKeyInput(variant: Variant, key: string) {
      this.markVariantDirty(variant)
      this.$emit('update-variant-key', { variant, key })
    },
    onAttachmentChange(variant: Variant, val: unknown, valid: boolean) {
      this.markVariantDirty(variant)
      if (valid && val !== null && typeof val === 'object' && !Array.isArray(val)) {
        variant.attachment = val as VariantAttachment
      }
      variant.attachmentValid = valid
      this.$emit('attachment-change', { variant, valid })
    },
    onAttachmentTextChange(variant: Variant, text: string) {
      this.markVariantDirty(variant)
      try {
        const v = JSON.parse(text) as VariantAttachment
        variant.attachment = v
        variant.attachmentValid = true
        this.$emit('attachment-change', { variant, valid: true })
      } catch {
        variant.attachmentValid = false
        this.$emit('attachment-change', { variant, valid: false })
      }
    },
  },
}
</script>

<style lang="scss" scoped>
.variants-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2xs);
}
.variant-item {
  flex: 1;
  min-width: 260px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-light);
  border-radius: var(--surface-inset-radius);
  padding: var(--surface-inset-padding);
}
.variant-row {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
}
.variant-key-field { flex: 1; }
.variant-actions {
  display: flex;
  gap: var(--space-3xs);
  flex-shrink: 0;
}
.variant-add-row {
  display: flex;
  gap: var(--space-2xs);
  margin-top: var(--space-2xs);
  > *:first-child { flex: 1; }
}
.variant-attachment-collapse {
  margin-top: var(--space-3xs);
  border: none;

  :deep(.el-collapse-item__header) {
    height: auto;
    min-height: 32px;
    line-height: 1.4;
    font-size: var(--font-size-caption);
    font-weight: var(--font-weight-semibold);
    letter-spacing: var(--letter-spacing-ui);
    color: var(--el-text-color-secondary);
    font-family: inherit;
    background: transparent;
    border-bottom: none;
  }

  :deep(.el-collapse-item__arrow) {
    color: var(--el-text-color-placeholder);
  }

  :deep(.el-collapse-item__wrap),
  :deep(.el-collapse-item__content) {
    border-bottom: none;
    padding-bottom: 0;
  }
}

.variant-attachment-title {
  margin: 0 0 var(--space-3xs);
  font-size: var(--font-size-caption);
  font-weight: var(--font-weight-normal);
  line-height: var(--line-height-tight);
  color: var(--el-text-color-placeholder);
}
@media (max-width: 640px) {
  .variant-row {
    flex-wrap: wrap;
  }
  .variant-key-field {
    flex: 1 1 100%;
    order: -1;
  }
  .variant-actions {
    margin-left: auto;
  }
}
</style>
