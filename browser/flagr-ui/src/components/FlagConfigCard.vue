<template>
  <el-card class="flag-config-card">
    <template #header>
      <div class="el-card-header">
        <div class="flex-row">
          <div class="flex-row-left flag-config-title">
            <h2>Flag</h2>
            <span class="flag-id ui-id-badge">#{{ flag.id }}</span>
            <copy-link-button
              v-if="flag.id != null"
              :url="flagShareUrl"
              aria-label="Copy flag URL"
              tooltip="Copy flag URL"
              test-id="copy-flag-url-btn"
            />
          </div>
          <div class="flex-row-right">
            <el-tooltip
              :content="SAVE_DIRTY_TOOLTIP"
              placement="top"
              effect="light"
              :disabled="!flagDirty"
            >
              <el-button
                size="small"
                :plain="!flagDirty"
                :type="saveButtonType(flagDirty)"
                data-testid="save-flag-btn"
                @click="handleSaveFlag"
              >
                {{ saveButtonLabel(flagDirty) }}
              </el-button>
            </el-tooltip>
            <el-switch
              :model-value="flag.enabled"
              :active-value="true"
              :inactive-value="false"
              inline-prompt
              active-text="Live"
              inactive-text="Off"
              data-testid="flag-enable-switch"
              @update:model-value="$emit('toggle-enabled', $event)"
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
          <label class="flag-label ui-field-label">Flag Key</label>
          <el-input
            size="small"
            placeholder="Key"
            :model-value="flag.key"
            data-testid="flag-key-input"
            @update:model-value="onUpdateFlag({ key: $event })"
          />
        </div>

        <!-- Description -->
        <div class="flag-field-block">
          <label class="flag-label ui-field-label">Description</label>
          <el-input
            size="small"
            placeholder="Description"
            :model-value="flag.description"
            data-testid="flag-desc-input"
            @update:model-value="onUpdateFlag({ description: $event })"
          />
        </div>

        <!-- Data Records + Entity Type in a compact row -->
        <div class="flag-compact-row">
          <div class="flag-field-block flag-field-narrow">
            <label class="flag-label ui-field-label">Data Records</label>
            <div class="flag-inline-row">
              <el-switch
                size="small"
                :model-value="flag.dataRecordsEnabled"
                :active-value="true"
                :inactive-value="false"
                data-testid="data-records-switch"
                @update:model-value="onUpdateFlag({ dataRecordsEnabled: $event })"
              />
              <el-tooltip
                content="When enabled, evaluation and exposure events are sent to the data pipeline (e.g. Kafka, Kinesis, Pub/Sub)"
                placement="top"
                effect="light"
              >
                <el-icon style="color: var(--el-text-color-placeholder);">
                  <InfoFilled />
                </el-icon>
              </el-tooltip>
            </div>
          </div>
          <div
            v-show="!!flag.dataRecordsEnabled"
            class="flag-field-block"
          >
            <label class="flag-label ui-field-label">Entity Type</label>
            <el-select
              :model-value="flag.entityType"
              size="small"
              filterable
              :allow-create="allowCreateEntityType"
              default-first-option
              placeholder="Entity Type"
              style="width: 100%;"
              @update:model-value="onUpdateFlag({ entityType: $event })"
            >
              <el-option
                v-for="item in entityTypes"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </div>
        </div>

        <!-- Tags -->
        <div class="flag-field-block flag-tags-block">
          <label class="flag-label ui-field-label">Tags</label>
          <div class="flag-tags-row">
            <el-tag
              v-for="tag in flag.tags"
              :key="tag.id"
              closable
              size="small"
              effect="plain"
              disable-transitions
              :style="{ backgroundColor: tagColor(tag.value), borderColor: 'transparent' }"
              @close="$emit('delete-tag', tag)"
            >
              {{ tag.value }}
            </el-tag>
            <el-autocomplete
              v-if="tagInputVisible"
              ref="saveTagInput"
              v-model="newTagValue"
              size="small"
              :trigger-on-focus="false"
              :fetch-suggestions="queryTags"
              style="width: 150px;"
              data-testid="new-tag-input"
              @select="() => $emit('create-tag', { value: newTagValue })"
              @keyup.enter="() => $emit('create-tag', { value: newTagValue })"
              @keyup.esc="cancelCreateTag"
            />
            <el-button
              v-else
              size="small"
              link
              type="primary"
              @click="showTagInput"
            >
              + Tag
            </el-button>
          </div>
        </div>
      </div>

      <!-- Right column: Notes -->
      <div class="flag-right">
        <div class="flag-notes-panel ui-surface-inset">
          <div class="flag-section-header">
            <label class="flag-label ui-field-label">Notes</label>
            <el-button
              size="small"
              link
              type="primary"
              @click="$emit('toggle-notes')"
            >
              <el-icon v-if="!showMdEditor">
                <Edit />
              </el-icon>
              <el-icon v-else>
                <View />
              </el-icon>
              {{ showMdEditor ? "view" : "edit" }}
            </el-button>
          </div>
          <markdown-editor
            v-if="showMdEditor || flag.notes"
            :show-editor="showMdEditor"
            :markdown="flag.notes"
            @update:markdown="onUpdateFlag({ notes: $event })"
            @save="handleSaveFlag"
          />
        </div>
      </div>
    </div>
  </el-card>
</template>

<script lang="ts">
import {
  SAVE_DIRTY_TOOLTIP,
  saveButtonLabel as fmtSaveLabel,
  saveButtonType as fmtSaveType,
} from '@/helpers/saveDirtyUi'
import { defineAsyncComponent, type PropType } from 'vue'
import { InfoFilled, Edit, View } from '@element-plus/icons-vue'
import { tagColor } from '@/helpers/tagColor'
import { flagUrl } from '@/helpers/shareLinks'
import CopyLinkButton from '@/components/CopyLinkButton.vue'
import type { FlagView, Tag } from '@/api/types'

interface EntityTypeOption {
  label: string
  value: string
}

export default {
  name: 'FlagConfigCard',
  components: {
    MarkdownEditor: defineAsyncComponent(() => import('@/components/MarkdownEditor.vue')),
    CopyLinkButton,
    InfoFilled,
    Edit,
    View,
  },
  props: {
    flag: { type: Object as PropType<FlagView>, required: true },
    showMdEditor: Boolean,
    entityTypes: { type: Array as PropType<EntityTypeOption[]>, default: () => [] },
    allowCreateEntityType: { type: Boolean, default: true },
    tagInputVisible: { type: Boolean, default: false },
    allTags: { type: Array as PropType<Tag[]>, default: () => [] },
  },
  emits: [
    'toggle-enabled',
    'save-flag',
    'update-flag',
    'toggle-notes',
    'delete-tag',
    'create-tag',
    'cancel-create-tag',
    'show-tag-input',
  ],
  data() {
    return {
      SAVE_DIRTY_TOOLTIP,
      newTagValue: '',
      flagDirty: false,
    }
  },
  computed: {
    flagShareUrl(): string {
      if (this.flag.id == null) return ''
      return flagUrl(this.flag.id, window.location)
    },
  },
  methods: {
    tagColor,
    saveButtonLabel(dirty: boolean) {
      return fmtSaveLabel(dirty)
    },
    saveButtonType(dirty: boolean) {
      return fmtSaveType(dirty)
    },
    markFlagDirty() {
      this.flagDirty = true
    },
    handleSaveFlag() {
      this.$emit('save-flag')
      this.flagDirty = false
    },
    onUpdateFlag(patch: Partial<FlagView>) {
      this.markFlagDirty()
      this.$emit('update-flag', patch)
    },
    queryTags(queryString: string, cb: (results: Tag[]) => void) {
      cb(
        this.allTags.filter((tag) =>
          tag.value.toLowerCase().includes(queryString.toLowerCase()),
        ),
      )
    },
    cancelCreateTag() {
      this.newTagValue = ''
      this.$emit('cancel-create-tag')
    },
    showTagInput() {
      this.$emit('show-tag-input')
      this.$nextTick(() => {
        const el = this.$refs.saveTagInput as { focus?: () => void } | undefined
        el?.focus?.()
      })
    },
  },
}
</script>

<style lang="scss" scoped>
.flag-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-md);
  padding: var(--space-3xs) 0;
}

.flag-left {
  display: flex;
  flex-direction: column;
  gap: var(--space-2xs);
}

.flag-right {
  min-width: 0;
}

.flag-notes-panel {
  min-height: 120px;
}

.flag-config-title {
  display: flex;
  align-items: center;
  gap: var(--space-3xs);
  min-width: 0;
}

.flag-id {
  margin-left: var(--space-2xs);
}


.flag-field-block {
  // each field block in the left column
}

.flag-compact-row {
  display: flex;
  gap: var(--space-xs);
  align-items: flex-start;
}

.flag-field-narrow {
  flex: 0 0 auto;
  min-width: 160px;
}

.flag-inline-row {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
}

.flag-section-header {
  display: flex;
  align-items: center;
  gap: var(--space-2xs);
  margin-bottom: var(--space-2xs);
}

.flag-tags-block {
  margin-top: 2px;
}

.flag-tags-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-3xs);
}
@media (max-width: 768px) {
  .flag-grid {
    grid-template-columns: 1fr;
    gap: var(--space-sm);
  }
  .flag-compact-row {
    flex-direction: column;
    gap: var(--space-2xs);
  }
  .flag-field-narrow {
    min-width: 0;
  }
}
</style>