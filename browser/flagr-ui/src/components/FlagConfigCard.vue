<template>
  <el-card class="flag-config-card">
    <template #header>
      <div class="el-card-header">
        <div class="flex-row">
          <div class="flex-row-left">
            <h2>Flag</h2>
            <span class="flag-id">#{{ flag.id }}</span>
          </div>
          <div class="flex-row-right">
            <el-button size="small" plain @click="$emit('save-flag')" data-testid="save-flag-btn">Save</el-button>
            <el-switch
              :model-value="flag.enabled"
              @update:model-value="$emit('toggle-enabled', $event)"
              :active-value="true"
              :inactive-value="false"
              inline-prompt
              active-text="Live"
              inactive-text="Off"
              data-testid="flag-enable-switch"
            />
          </div>
        </div>
      </div>
    </template>

    <div class="flag-grid">
      <!-- Left column: main fields -->
      <div class="flag-left">
        <!-- Key -->
        <div class="flag-field-block">
          <label class="flag-label">Flag Key</label>
          <el-input size="small" placeholder="Key" :model-value="flag.key" @update:model-value="$emit('update-flag', { key: $event })" data-testid="flag-key-input" />
        </div>

        <!-- Description -->
        <div class="flag-field-block">
          <label class="flag-label">Description</label>
          <el-input size="small" placeholder="Description" :model-value="flag.description" @update:model-value="$emit('update-flag', { description: $event })" data-testid="flag-desc-input" />
        </div>

        <!-- Data Records + Entity Type in a compact row -->
        <div class="flag-compact-row">
          <div class="flag-field-block flag-field-narrow">
            <label class="flag-label">Data Records</label>
            <div class="flag-inline-row">
              <el-switch size="small" :model-value="flag.dataRecordsEnabled" :active-value="true" :inactive-value="false"
                @update:model-value="$emit('update-flag', { dataRecordsEnabled: $event })" data-testid="data-records-switch" />
              <el-tooltip content="Controls whether to log to data pipeline, e.g. Kafka, Kinesis, Pubsub" placement="top" effect="light">
                <el-icon style="color: var(--el-text-color-placeholder);"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </div>
          <div class="flag-field-block" v-show="!!flag.dataRecordsEnabled">
            <label class="flag-label">Entity Type</label>
            <el-select :model-value="flag.entityType" @update:model-value="$emit('update-flag', { entityType: $event })"
              size="small" filterable :allow-create="allowCreateEntityType" default-first-option placeholder="Entity Type" style="width: 100%;">
              <el-option v-for="item in entityTypes" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
          </div>
        </div>

        <!-- Tags -->
        <div class="flag-field-block flag-tags-block">
          <label class="flag-label">Tags</label>
          <div class="flag-tags-row">
            <el-tag v-for="tag in flag.tags" :key="tag.id" closable size="small" type="success" @close="$emit('delete-tag', tag)">{{ tag.value }}</el-tag>
            <el-autocomplete v-if="tagInputVisible" v-model="newTagValue" ref="saveTagInput" size="small"
              :trigger-on-focus="false" :fetch-suggestions="queryTags"
              @select="() => $emit('create-tag', { value: newTagValue })"
              @keyup.enter="() => $emit('create-tag', { value: newTagValue })"
              @keyup.esc="cancelCreateTag" style="width: 150px;" data-testid="new-tag-input" />
            <el-button v-else size="small" link type="primary" @click="showTagInput">+ Tag</el-button>
          </div>
        </div>
      </div>

      <!-- Right column: Notes -->
      <div class="flag-right">
        <div class="flag-notes-panel">
          <div class="flag-section-header">
            <label class="flag-label">Notes</label>
            <el-button size="small" link type="primary" @click="$emit('toggle-notes')">
              <el-icon v-if="!showMdEditor"><Edit /></el-icon>
              <el-icon v-else><View /></el-icon>
              {{ showMdEditor ? "view" : "edit" }}
            </el-button>
          </div>
          <markdown-editor
            v-if="showMdEditor || flag.notes"
            :show-editor="showMdEditor"
            :markdown="flag.notes"
            @update:markdown="$emit('update-flag', { notes: $event })"
            @save="$emit('save-flag')"
          />
        </div>
      </div>
    </div>
  </el-card>
</template>

<script>
import { defineAsyncComponent } from "vue"
import { InfoFilled, Edit, View } from "@element-plus/icons-vue"

export default {
  name: "flag-config-card",
  components: {
    MarkdownEditor: defineAsyncComponent(() => import("@/components/MarkdownEditor.vue")),
    InfoFilled, Edit, View
  },
  props: {
    flag: { type: Object, required: true },
    showMdEditor: Boolean,
    entityTypes: { type: Array, default: () => [] },
    allowCreateEntityType: { type: Boolean, default: true },
    tagInputVisible: { type: Boolean, default: false },
    allTags: { type: Array, default: () => [] }
  },
  emits: [
    "toggle-enabled", "save-flag", "update-flag", "toggle-notes",
    "delete-tag", "create-tag", "cancel-create-tag", "show-tag-input"
  ],
  data() {
    return { newTagValue: "" }
  },
  methods: {
    queryTags(queryString, cb) {
      cb(this.allTags.filter(tag => tag.value.toLowerCase().includes(queryString.toLowerCase())))
    },
    cancelCreateTag() { this.newTagValue = ""; this.$emit("cancel-create-tag") },
    showTagInput() {
      this.$emit("show-tag-input")
      this.$nextTick(() => { if (this.$refs.saveTagInput) this.$refs.saveTagInput.focus() })
    }
  }
}
</script>

<style lang="less" scoped>
.flag-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  padding: 4px 0;
}

.flag-left {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.flag-right {
  min-width: 0;
}

.flag-notes-panel {
  background: var(--el-fill-color-light);
  border-radius: 8px;
  padding: 10px 12px;
  min-height: 120px;
}

.flag-id {
  font-size: 10px;
  font-weight: 600;
  color: var(--el-text-color-placeholder);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  letter-spacing: 0.02em;
  margin-left: 8px;
}

.flag-field-block {
  // each field block in the left column
}

.flag-label {
  display: block;
  font-size: 11px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
  margin-bottom: 3px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.flag-compact-row {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.flag-field-narrow {
  flex: 0 0 auto;
  min-width: 160px;
}

.flag-inline-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.flag-section-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}

.flag-tags-block {
  margin-top: 2px;
}

.flag-tags-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
}
</style>