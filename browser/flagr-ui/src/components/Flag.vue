<template>
  <div class="container flag-container">
        <el-dialog title="Delete feature flag" v-model="dialogDeleteFlagVisible">
          <span>Are you sure you want to delete this feature flag?</span>
          <template #footer>
            <span class="dialog-footer">
              <el-button @click="dialogDeleteFlagVisible = false">Cancel</el-button>
              <el-button type="primary" @click.prevent="deleteFlag">Confirm</el-button>
            </span>
          </template>
        </el-dialog>

        <el-dialog title="Create segment" v-model="dialogCreateSegmentOpen">
          <div class="create-segment-dialog">
            <el-input placeholder="Segment description" v-model="newSegment.description" data-testid="new-segment-desc-input" />
            <div class="create-segment-slider">
              <label class="create-segment-label">Rollout %</label>
              <el-slider v-model="newSegment.rolloutPercent" show-input :max="100" />
            </div>
            <el-button
              class="width--full"
              type="primary"
              :disabled="!newSegment.description"
              @click.prevent="createSegment"
              data-testid="create-segment-btn"
            >Create Segment</el-button>
          </div>
        </el-dialog>

        <distribution-dialog
          :visible="dialogEditDistributionOpen"
          :flag="flag"
          :initial-distributions="distributionDraft"
          @update:visible="dialogEditDistributionOpen = $event"
          @save="handleSaveDistribution"
        />

        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ name: 'home' }">Home page</el-breadcrumb-item>
          <el-breadcrumb-item>Flag ID: {{ $route.params.flagId }}</el-breadcrumb-item>
        </el-breadcrumb>

        <div v-if="loaded && flag">
          <el-tabs @tab-click="handleHistoryTabClick">
            <el-tab-pane label="Config">
              <flag-config-card
                :flag="flag"
                :show-md-editor="showMdEditor"
                :entity-types="entityTypes"
                :allow-create-entity-type="allowCreateEntityType"
                :tag-input-visible="tagInputVisible"
                :all-tags="allTags"
                @toggle-enabled="handleToggleEnabled"
                @save-flag="putFlag"
                @update-flag="handleUpdateFlag"
                @toggle-notes="showMdEditor = !showMdEditor"
                @delete-tag="deleteTag"
                @create-tag="handleCreateTag"
                @cancel-create-tag="handleCancelCreateTag"
                @show-tag-input="handleShowTagInput"
              />

              <variants-section
                :variants="flag.variants"
                @create-variant="handleCreateVariant"
                @update-variant-key="handleUpdateVariantKey"
                @save-variant="putVariant"
                @delete-variant="deleteVariant"
                @attachment-change="handleVariantAttachmentChange"
              />

              <segments-section
                :segments="flag.segments"
                :operator-options="operatorOptions"
                @reorder="handleReorderSegments"
                @move-up="moveSegmentUp"
                @move-down="moveSegmentDown"
                @new-segment="dialogCreateSegmentOpen = true"
                @save-segment="putSegment"
                @delete-segment="deleteSegment"
                @update-segment-field="handleUpdateSegmentField"
                @create-constraint="createConstraint"
                @save-constraint="putConstraint"
                @delete-constraint="deleteConstraint"
                @update-constraint-field="handleUpdateConstraintField"
                @edit-distribution="handleEditDistribution"
              />

              <debug-console :flag="flag" />

              <div style="margin-top: 8px;">
                <el-button type="danger" plain size="small" @click="dialogDeleteFlagVisible = true" data-testid="delete-flag-btn">
                  <el-icon><Delete /></el-icon>
                  Delete Flag
                </el-button>
              </div>
            </el-tab-pane>

            <el-tab-pane label="History" name="history">
              <flag-history v-if="historyLoaded" :key="historyKey" :flag-id="parseInt($route.params.flagId, 10)"></flag-history>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
</template>

<script>
import Axios from "axios"
import { Delete } from "@element-plus/icons-vue"

import constants from "@/constants"
import helpers from "@/helpers/helpers"
import DebugConsole from "@/components/DebugConsole"
import FlagHistory from "@/components/FlagHistory"
import DistributionDialog from "@/components/DistributionDialog"
import FlagConfigCard from "@/components/FlagConfigCard"
import VariantsSection from "@/components/VariantsSection"
import SegmentsSection from "@/components/SegmentsSection"
import operatorsData from "@/operators.json"

const operators = operatorsData.operators
const { pluck, handleErr } = helpers
const { API_URL, FLAGR_UI_POSSIBLE_ENTITY_TYPES } = constants

const DEFAULT_SEGMENT = { description: "", rolloutPercent: 50 }
const DEFAULT_TAG = { value: "" }

function processVariant(variant) {
  if (typeof variant.attachment === "string") {
    variant.attachment = JSON.parse(variant.attachment)
  }
}

export default {
  name: "flag",
  components: {
    DebugConsole,
    FlagHistory,
    DistributionDialog,
    FlagConfigCard,
    VariantsSection,
    SegmentsSection,
    Delete
  },
  data() {
    return {
      loaded: false,
      dialogDeleteFlagVisible: false,
      dialogEditDistributionOpen: false,
      dialogCreateSegmentOpen: false,
      entityTypes: [],
      allTags: [],
      allowCreateEntityType: true,
      tagInputVisible: false,
      flag: {},
      newSegment: { ...DEFAULT_SEGMENT },
      newTag: { ...DEFAULT_TAG },
      selectedSegment: null,
      distributionDraft: {},
      operatorOptions: operators,
      showMdEditor: false,
      historyLoaded: false,
      historyKey: 0
    }
  },
  computed: {
    flagId() {
      return this.$route.params.flagId
    }
  },
  methods: {
    // --- Flag CRUD ---
    deleteFlag() {
      const id = this.flagId
      Axios.delete(`${API_URL}/flags/${id}`).then(() => {
        this.$router.replace({ name: "home" })
        this.$message.success(`You deleted flag ${id}`)
      }, handleErr.bind(this))
    },

    putFlag() {
      const f = this.flag
      Axios.put(`${API_URL}/flags/${this.flagId}`, {
        description: f.description,
        dataRecordsEnabled: f.dataRecordsEnabled,
        key: f.key || "",
        entityType: f.entityType || "",
        notes: f.notes || ""
      }).then(() => {
        this.$message.success("Flag updated")
      }, handleErr.bind(this))
    },

    handleToggleEnabled(checked) {
      Axios.put(`${API_URL}/flags/${this.flagId}/enabled`, {
        enabled: checked
      }).then(() => {
        this.flag.enabled = checked
        this.$message.success(`You turned ${checked ? "on" : "off"} this feature flag`)
      }, handleErr.bind(this))
    },

    handleUpdateFlag(patch) {
      Object.assign(this.flag, patch)
    },

    // --- Tags ---
    handleCreateTag({ value }) {
      this.newTag.value = value
      Axios.post(`${API_URL}/flags/${this.flagId}/tags`, { value }).then(
        response => {
          const tag = response.data
          this.newTag = clone(DEFAULT_TAG)
          if (!this.flag.tags.map(t => t.value).includes(tag.value)) {
            this.flag.tags.push(tag)
            this.$message.success("new tag created")
          }
          this.tagInputVisible = false
          this.loadAllTags()
        },
        handleErr.bind(this)
      )
    },

    handleCancelCreateTag() {
      this.newTag = clone(DEFAULT_TAG)
      this.tagInputVisible = false
    },

    handleShowTagInput() {
      this.tagInputVisible = true
    },

    deleteTag(tag) {
      this.$confirm(`Are you sure you want to delete tag #${tag.value}`, "Warning", {
        confirmButtonText: "OK", cancelButtonText: "Cancel", type: "warning"
      }).then(() => {
        Axios.delete(`${API_URL}/flags/${this.flagId}/tags/${tag.id}`).then(
          () => {
            this.$message.success("tag deleted")
            this.fetchFlag()
            this.loadAllTags()
          },
          handleErr.bind(this)
        )
      }).catch(() => {})
    },

    loadAllTags() {
      Axios.get(`${API_URL}/tags`).then(response => {
        this.allTags = response.data
      }, handleErr.bind(this))
    },

    // --- Variants ---
    handleCreateVariant({ key }) {
      Axios.post(`${API_URL}/flags/${this.flagId}/variants`, { key }).then(
        response => {
          this.flag.variants = [...this.flag.variants, response.data]
          this.$message.success("new variant created")
        },
        handleErr.bind(this)
      )
    },


    handleUpdateVariantKey({ variant, key }) {
      variant.key = key
    },


    putVariant(variant) {
      if (variant.attachmentValid === false) {
        this.$message.error("variant attachment is not valid")
        return
      }
      Axios.put(`${API_URL}/flags/${this.flagId}/variants/${variant.id}`, { key: variant.key, attachment: variant.attachment }).then(
        () => this.$message.success("variant updated"),
        handleErr.bind(this)
      )
    },

    deleteVariant(variant) {
      if (this.flag.segments.some(s =>
        s.distributions.some(d => d.variantID === variant.id)
      )) {
        this.$message.warning(
          "This variant is being used by a segment distribution. Please remove the segment or edit the distribution in order to remove this variant."
        )
        return
      }
      this.$confirm(
        `Are you sure you want to delete variant #${variant.id} [${variant.key}]`,
        "Warning",
        { confirmButtonText: "OK", cancelButtonText: "Cancel", type: "warning" }
      ).then(() => {
        Axios.delete(`${API_URL}/flags/${this.flagId}/variants/${variant.id}`).then(
          () => {
            this.$message.success("variant deleted")
            this.fetchFlag()
          },
          handleErr.bind(this)
        )
      }).catch(() => {})
    },


    // --- Segments ---
    createSegment() {
      Axios.post(`${API_URL}/flags/${this.flagId}/segments`, this.newSegment).then(
        response => {
          const segment = response.data
          segment.constraints = []
          this.newSegment = clone(DEFAULT_SEGMENT)
          this.flag.segments = [...this.flag.segments, segment]
          this.dialogCreateSegmentOpen = false
          this.$message.success("new segment created")
        },
        handleErr.bind(this)
      )
    },

    putSegment(segment) {
      Axios.put(`${API_URL}/flags/${this.flagId}/segments/${segment.id}`, {
        description: segment.description,
        rolloutPercent: parseInt(segment.rolloutPercent, 10)
      }).then(() => {
        this.$message.success("segment updated")
      }, handleErr.bind(this))
    },

    deleteSegment(segment) {
      this.$confirm("Are you sure you want to delete this segment?", "Warning", {
        confirmButtonText: "OK", cancelButtonText: "Cancel", type: "warning"
      }).then(() => {
        Axios.delete(`${API_URL}/flags/${this.flagId}/segments/${segment.id}`).then(
          () => {
            this.flag.segments = this.flag.segments.filter(el => el.id !== segment.id)
            this.$message.success("segment deleted")
          },
          handleErr.bind(this)
        )
      }).catch(() => {})
    },

    handleReorderSegments(segments) {
      Axios.put(`${API_URL}/flags/${this.flagId}/segments/reorder`, {
        segmentIDs: pluck(segments, "id")
      }).then(() => {
        this.$message.success("segment reordered")
      }, handleErr.bind(this))
    },

    moveSegmentUp(_element, index) {
      if (index <= 0) return
      const arr = [...this.flag.segments]
      const temp = arr[index - 1]
      arr[index - 1] = arr[index]
      arr[index] = temp
      this.flag.segments = arr
    },
    moveSegmentDown(_element, index) {
      if (index >= this.flag.segments.length - 1) return
      const arr = [...this.flag.segments]
      const temp = arr[index + 1]
      arr[index + 1] = arr[index]
      arr[index] = temp
      this.flag.segments = arr
    },


    handleUpdateSegmentField({ segment, field, value }) {
      segment[field] = value
    },

    // --- Constraints ---
    createConstraint({ segment, constraint }) {
      const c = { ...constraint }
      c.property = c.property.trim()
      c.value = c.value.trim()
      Axios.post(
        `${API_URL}/flags/${this.flagId}/segments/${segment.id}/constraints`,
        c
      ).then(response => {
        segment.constraints = [...segment.constraints, response.data]
        this.$message.success("new constraint created")
      }, handleErr.bind(this))
    },

    putConstraint({ segment, constraint }) {
      constraint.property = constraint.property.trim()
      constraint.value = constraint.value.trim()
      Axios.put(
        `${API_URL}/flags/${this.flagId}/segments/${segment.id}/constraints/${constraint.id}`,
        constraint
      ).then(() => {
        this.$message.success("constraint updated")
      }, handleErr.bind(this))
    },

    deleteConstraint({ segment, constraint }) {
      this.$confirm("Are you sure you want to delete this constraint?", "Warning", {
        confirmButtonText: "OK", cancelButtonText: "Cancel", type: "warning"
      }).then(() => {
        Axios.delete(
          `${API_URL}/flags/${this.flagId}/segments/${segment.id}/constraints/${constraint.id}`
        ).then(() => {
          segment.constraints = segment.constraints.filter(c => c.id !== constraint.id)
          this.$message.success("constraint deleted")
        }, handleErr.bind(this))
      }).catch(() => {})
    },

    handleUpdateConstraintField({ constraint, field, value }) {
      constraint[field] = value
    },

    // --- Distributions ---
    handleEditDistribution(segment) {
      this.selectedSegment = segment
      this.distributionDraft = {}
      segment.distributions.forEach(d => {
        this.distributionDraft[d.variantID] = clone(d)
      })
      this.dialogEditDistributionOpen = true
    },

    handleSaveDistribution(draft) {
      const distributions = Object.values(draft)
        .filter(d => d.percent !== 0)
        .map(d => {
          const dist = clone(d)
          delete dist.id
          return dist
        })
      Axios.put(
        `${API_URL}/flags/${this.flagId}/segments/${this.selectedSegment.id}/distributions`,
        { distributions }
      ).then(response => {
        this.selectedSegment.distributions = response.data
        this.dialogEditDistributionOpen = false
        this.$message.success("distributions updated")
      }, handleErr.bind(this))
    },

    // --- Other ---
    handleHistoryTabClick(tab) {
      if (tab.props?.name === 'history') {
        this.historyLoaded = true
        this.historyKey++
      }
    },

    // --- Data fetching ---
    fetchFlag() {
      Axios.get(`${API_URL}/flags/${this.flagId}`).then(response => {
        const flag = response.data
        flag.variants.forEach(v => processVariant(v))
        this.flag = flag
        this.loaded = true
      }, handleErr.bind(this))
      this.fetchEntityTypes()
    },

    fetchEntityTypes() {
      const prepareEntityTypes = (entityTypes) => {
        const arr = entityTypes.map(key => ({
          label: key === "" ? "<null>" : key,
          value: key
        }))
        if (entityTypes.indexOf("") === -1) {
          arr.unshift({ label: "<null>", value: "" })
        }
        return arr
      }

      if (FLAGR_UI_POSSIBLE_ENTITY_TYPES && FLAGR_UI_POSSIBLE_ENTITY_TYPES != "null") {
        this.entityTypes = prepareEntityTypes(FLAGR_UI_POSSIBLE_ENTITY_TYPES.split(","))
        this.allowCreateEntityType = false
        return
      }
      Axios.get(`${API_URL}/flags/entity_types`).then(response => {
        this.entityTypes = prepareEntityTypes(response.data)
      }, handleErr.bind(this))
    }
  },
  mounted() {
    this.fetchFlag()
    this.loadAllTags()
  }
}
</script>

<style lang="less">
h5 {
  padding: 0;
  margin: 8px 0 4px;
  font-size: 13px;
}

.create-segment-dialog {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.create-segment-slider {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.create-segment-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--el-text-color-secondary);
}


.grabbable {
  cursor: move;
  cursor: grab;
  cursor: -moz-grab;
  cursor: -webkit-grab;
}
</style>
