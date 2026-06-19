<template>
  <el-card class="variants-container is-card-secondary">
    <template #header>
      <div class="el-card-header"><h2>Variants</h2></div>
    </template>

    <div v-if="variants.length" class="variants-row">
      <div v-for="variant in variants" :key="variant.id" class="variant-item">
        <div class="variant-row">
          <span class="variant-id">#{{ variant.id }}</span>
          <el-input
            size="small"
            placeholder="Variant Key"
            :model-value="variant.key"
            @update:model-value="(v) => $emit('update-variant-key', { variant, key: v })"
            data-testid="variant-key-input"
            class="variant-key-field"
          />
          <div class="variant-actions">
            <el-button size="small" plain @click="$emit('save-variant', variant)" data-testid="save-variant-btn">Save</el-button>
            <el-button size="small" plain @click="$emit('delete-variant', variant)" data-testid="delete-variant-btn"><el-icon><Delete /></el-icon></el-button>
          </div>
        </div>
        <el-collapse>
          <el-collapse-item title="Variant attachment" class="variant-attachment-collapsable-title">
            <p class="variant-attachment-title">JSON in key/value pairs format.</p>
            <json-editor
              :json="variant.attachment"
              :main-menu-bar="false" :navigation-bar="false" :status-bar="false"
              mode="text"
              @update:json="onAttachmentChange(variant, $event, true)"
              @update:jsonString="onAttachmentTextChange(variant, $event)"
              @error="onAttachmentChange(variant, null, false)"
            />
          </el-collapse-item>
        </el-collapse>
      </div>
    </div>
    <div class="card--cue" v-else>
      <p class="card--cue-title">No variants yet</p>
      <p class="card--cue-body">Variants are the outcomes a flag can return — e.g. <code>on</code>/<code>off</code>, <code>control</code>/<code>test</code>. Add one to start defining what entities receive.</p>
    </div>

    <div class="variant-add-row">
      <el-input size="small" placeholder="New Variant Key" v-model="newKey" data-testid="new-variant-input" />
      <el-button type="primary" size="small" :disabled="!newKey" @click.prevent="createVariant" data-testid="create-variant-btn">Create Variant</el-button>
    </div>
  </el-card>
</template>

<script>
import JsonEditor from "vue3-ts-jsoneditor"
import { Delete } from "@element-plus/icons-vue"

export default {
  name: "variants-section",
  components: { JsonEditor, Delete },
  props: { variants: { type: Array, required: true } },
  emits: ["update-variant-key", "save-variant", "delete-variant", "create-variant", "attachment-change"],
  data() { return { newKey: "" } },
  methods: {
    createVariant() { this.$emit("create-variant", { key: this.newKey }); this.newKey = "" },
    onAttachmentChange(variant, val, valid) {
      if (val !== null) variant.attachment = val;
      variant.attachmentValid = valid;
      this.$emit("attachment-change", { variant, valid });
    },
    onAttachmentTextChange(variant, text) {
      try {
        const v = JSON.parse(text);
        variant.attachment = v;
        variant.attachmentValid = true;
        this.$emit("attachment-change", { variant, valid: true });
      } catch(e) {
        variant.attachmentValid = false;
        this.$emit("attachment-change", { variant, valid: false });
      }
    }
  }
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
  background: #fff;
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  padding: var(--space-2xs) var(--space-xs);
}
.variant-row {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
}
.variant-id {
  font-size: 10px;
  font-weight: 600;
  color: var(--el-text-color-placeholder);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  letter-spacing: 0.02em;
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
.variant-attachment-collapsable-title {
  margin: 0;
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}
.variant-attachment-title {
  margin: 0 0 var(--space-3xs);
  font-size: 11px;
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
