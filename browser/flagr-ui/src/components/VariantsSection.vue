<template>
  <el-card class="variants-container">
    <template #header>
      <div class="el-card-header">
        <h2>Variants</h2>
      </div>
    </template>

    <div class="variants-container-inner" v-if="variants.length">
      <div v-for="variant in variants" :key="variant.id">
        <el-card shadow="hover">
          <el-form label-position="left" label-width="100px">
            <div class="flex-row id-row">
              <el-tag type="primary">
                Variant ID: <b>{{ variant.id }}</b>
              </el-tag>
              <el-input
                class="variant-key-input"
                size="small"
                placeholder="Key"
                :model-value="variant.key"
                @update:model-value="(v) => $emit('update-variant-key', { variant, key: v })"
                data-testid="variant-key-input"
              >
                <template #prepend>Key</template>
              </el-input>
              <div class="flex-row-right save-remove-variant-row">
                <el-button size="small" @click="$emit('save-variant', variant)" data-testid="save-variant-btn">Save Variant</el-button>
                <el-button @click="$emit('delete-variant', variant)" size="small" data-testid="delete-variant-btn">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>
            <el-collapse class="flex-row">
              <el-collapse-item title="Variant attachment" class="variant-attachment-collapsable-title">
                <p class="variant-attachment-title">You can add JSON in key/value pairs format.</p>
                <json-editor
                  :json="variant.attachment"
                  :main-menu-bar="false"
                  :navigation-bar="false"
                  :status-bar="false"
                  mode="text"
                  @change="onAttachmentChange(variant, true)"
                  @error="onAttachmentChange(variant, false)"
                  class="variant-attachment-content"
                />
              </el-collapse-item>
            </el-collapse>
          </el-form>
        </el-card>
      </div>
    </div>
    <div class="card--error" v-else>No variants created for this feature flag yet</div>

    <div class="variants-input">
      <div class="flex-row equal-width constraints-inputs-container">
        <div>
          <el-input placeholder="Variant Key" v-model="newKey" data-testid="new-variant-input"></el-input>
        </div>
      </div>
      <el-button
        class="width--full"
        :disabled="!newKey"
        @click.prevent="createVariant"
        data-testid="create-variant-btn"
      >Create Variant</el-button>
    </div>
  </el-card>
</template>

<script>
import JsonEditor from "vue3-ts-jsoneditor"
import { Delete } from "@element-plus/icons-vue"

export default {
  name: "variants-section",
  components: { JsonEditor, Delete },
  props: {
    variants: { type: Array, required: true }
  },
  emits: [
    "update-variant-key",
    "save-variant",
    "delete-variant",
    "create-variant",
    "attachment-change"
  ],
  data() {
    return {
      newKey: ""
    }
  },
  methods: {
    createVariant() {
      this.$emit("create-variant", { key: this.newKey })
      this.newKey = ""
    },
    onAttachmentChange(variant, valid) {
      this.$emit("attachment-change", { variant, valid })
    }
  }
}
</script>
