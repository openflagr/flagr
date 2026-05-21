<template>
  <el-card class="flag-config-card">
    <template #header>
      <div class="el-card-header">
        <div class="flex-row">
          <div class="flex-row-left"><h2>Flag</h2></div>
          <div class="flex-row-right">
            <el-tooltip content="Enable/Disable Flag" placement="top" effect="light">
              <el-switch
                :model-value="flag.enabled"
                active-color="#13ce66"
                inactive-color="#ff4949"
                @update:model-value="$emit('toggle-enabled', $event)"
                :active-value="true"
                :inactive-value="false"
                data-testid="flag-enable-switch"
              ></el-switch>
            </el-tooltip>
          </div>
        </div>
      </div>
    </template>

    <el-card shadow="hover" :class="innerCardClass">
      <!-- Flag ID + Save -->
      <div class="flex-row id-row">
        <div class="flex-row-left">
          <el-tag type="primary">Flag ID: {{ flag.id }}</el-tag>
        </div>
        <div class="flex-row-right">
          <el-button size="small" @click="$emit('save-flag')" data-testid="save-flag-btn">Save Flag</el-button>
        </div>
      </div>

      <!-- Key -->
      <el-row class="flag-content" align="middle">
        <el-col :span="24">
          <el-input size="small" placeholder="Key" :model-value="flag.key" @update:model-value="$emit('update-flag', { key: $event })" data-testid="flag-key-input">
            <template #prepend>Flag Key</template>
          </el-input>
        </el-col>
      </el-row>

      <!-- Data Records + Entity Type -->
      <el-row class="flag-content" align="middle">
        <el-col :span="17">
          <el-input size="small" placeholder="Description" :model-value="flag.description" @update:model-value="$emit('update-flag', { description: $event })" data-testid="flag-desc-input">
            <template #prepend>Flag Description</template>
          </el-input>
        </el-col>
        <el-col style="text-align: right;" :span="5">
          <el-switch
            size="small"
            :model-value="flag.dataRecordsEnabled"
            active-color="#74E5E0"
            :active-value="true"
            :inactive-value="false"
            @update:model-value="$emit('update-flag', { dataRecordsEnabled: $event })"
            data-testid="data-records-switch"
          ></el-switch>
        </el-col>
        <el-col :span="2">
          <div class="data-records-label">
            Data Records
            <el-tooltip content="Controls whether to log to data pipeline, e.g. Kafka, Kinesis, Pubsub" placement="top-end" effect="light">
              <el-icon><InfoFilled /></el-icon>
            </el-tooltip>
          </div>
        </el-col>
      </el-row>

      <!-- Entity Type -->
      <el-row class="flag-content" align="middle">
        <el-col :span="17"></el-col>
        <el-col style="text-align: right;" :span="5">
          <el-select
            v-show="!!flag.dataRecordsEnabled"
            :model-value="flag.entityType"
            @update:model-value="$emit('update-flag', { entityType: $event })"
            size="small"
            filterable
            :allow-create="allowCreateEntityType"
            default-first-option
            placeholder="<null>"
          >
            <el-option v-for="item in entityTypes" :key="item.value" :label="item.label" :value="item.value"></el-option>
          </el-select>
        </el-col>
        <el-col :span="2">
          <div v-show="!!flag.dataRecordsEnabled" class="data-records-label">
            Entity Type
            <el-tooltip content="Overrides the entityType in data records logging" placement="top-end" effect="light">
              <el-icon><InfoFilled /></el-icon>
            </el-tooltip>
          </div>
        </el-col>
      </el-row>

      <!-- Notes -->
      <el-row style="margin: 10px;">
        <h5>
          <span style="margin-right: 10px;">Flag Notes</span>
          <el-button round size="small" @click="$emit('toggle-notes')">
            <el-icon v-if="!showMdEditor"><Edit /></el-icon>
            <el-icon v-else><View /></el-icon>
            {{ showMdEditor ? "view" : "edit" }}
          </el-button>
        </h5>
      </el-row>
      <el-row>
        <markdown-editor
          :show-editor="showMdEditor"
          :markdown="flag.notes"
          @update:markdown="$emit('update-flag', { notes: $event })"
          @save="$emit('save-flag')"
        ></markdown-editor>
      </el-row>

      <!-- Tags -->
      <el-row style="margin: 10px;">
        <h5><span style="margin-right: 10px;">Tags</span></h5>
      </el-row>
      <el-row>
        <div class="tags-container-inner">
          <el-tag
            v-for="tag in flag.tags"
            :key="tag.id"
            closable
            type="warning"
            @close="$emit('delete-tag', tag)"
          >{{ tag.value }}</el-tag>
          <el-autocomplete
            class="tag-key-input"
            v-if="tagInputVisible"
            v-model="newTagValue"
            ref="saveTagInput"
            size="small"
            :trigger-on-focus="false"
            :fetch-suggestions="queryTags"
            @select="() => $emit('create-tag', { value: newTagValue })"
            @keyup.enter="() => $emit('create-tag', { value: newTagValue })"
            @keyup.esc="cancelCreateTag"
            data-testid="new-tag-input"
          ></el-autocomplete>
          <el-button
            v-else
            class="button-new-tag"
            size="small"
            @click="showTagInput"
          >+ New Tag</el-button>
        </div>
      </el-row>
    </el-card>
  </el-card>
</template>

<script>
import MarkdownEditor from "@/components/MarkdownEditor.vue"
import { InfoFilled, Edit, View } from "@element-plus/icons-vue"

export default {
  name: "flag-config-card",
  components: { MarkdownEditor, InfoFilled, Edit, View },
  props: {
    flag: { type: Object, required: true },
    showMdEditor: Boolean,
    entityTypes: { type: Array, default: () => [] },
    allowCreateEntityType: { type: Boolean, default: true },
    tagInputVisible: { type: Boolean, default: false },
    allTags: { type: Array, default: () => [] }
  },
  emits: [
    "toggle-enabled",
    "save-flag",
    "update-flag",
    "toggle-notes",
    "delete-tag",
    "create-tag",
    "cancel-create-tag",
    "show-tag-input"
  ],
  data() {
    return {
      newTagValue: ""
    }
  },
  computed: {
    innerCardClass() {
      return !this.showMdEditor && !this.flag?.notes ? "flag-inner-config-card" : ""
    }
  },
  methods: {
    queryTags(queryString, cb) {
      const results = this.allTags.filter(tag =>
        tag.value.toLowerCase().includes(queryString.toLowerCase())
      )
      cb(results)
    },
    cancelCreateTag() {
      this.newTagValue = ""
      this.$emit("cancel-create-tag")
    },
    showTagInput() {
      this.$emit("show-tag-input")
      this.$nextTick(() => {
        if (this.$refs.saveTagInput) {
          this.$refs.saveTagInput.focus()
        }
      })
    }
  }
}
</script>
