<template>
  <el-row>
    <el-col :span="20" :offset="2">
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
          <div>
            <p>
              <el-input placeholder="Segment description" v-model="newSegment.description" data-testid="new-segment-desc-input"></el-input>
            </p>
            <p>
              <el-slider v-model="newSegment.rolloutPercent" show-input></el-slider>
            </p>
            <el-button
              class="width--full"
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

              <el-card>
                <template #header>
                  <div class="el-card-header"><h2>Flag Settings</h2></div>
                </template>
                <el-button type="danger" plain @click="dialogDeleteFlagVisible = true" data-testid="delete-flag-btn">
                  <el-icon><Delete /></el-icon>
                  Delete Flag
                </el-button>
              </el-card>
            </el-tab-pane>

            <el-tab-pane label="History">
              <flag-history v-if="historyLoaded" :flag-id="parseInt($route.params.flagId, 10)"></flag-history>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
    </el-col>
  </el-row>
</template>

<script>
import Axios from "axios"
import clone from "lodash.clone"
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
const DEFAULT_CONSTRAINT = { operator: "EQ", property: "", value: "" }
const DEFAULT_TAG = { value: "" }

function processSegment(segment) {
  segment._newConstraint = clone(DEFAULT_CONSTRAINT)
}

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
      newSegment: clone(DEFAULT_SEGMENT),
      newTag: clone(DEFAULT_TAG),
      selectedSegment: null,
      distributionDraft: {},
      operatorOptions: operators,
      showMdEditor: false,
      historyLoaded: false
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
          this.flag.variants.push(response.data)
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
      Axios.put(`${API_URL}/flags/${this.flagId}/variants/${variant.id}`, variant).then(
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

    handleVariantAttachmentChange({ variant, valid }) {
      variant.attachmentValid = valid
    },

    // --- Segments ---
    createSegment() {
      Axios.post(`${API_URL}/flags/${this.flagId}/segments`, this.newSegment).then(
        response => {
          const segment = response.data
          processSegment(segment)
          segment.constraints = []
          this.newSegment = clone(DEFAULT_SEGMENT)
          this.flag.segments.push(segment)
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
            const idx = this.flag.segments.findIndex(el => el.id === segment.id)
            this.flag.segments.splice(idx, 1)
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
      const arr = this.flag.segments
      const temp = arr[index - 1]
      arr[index - 1] = arr[index]
      arr[index] = temp
    },
    moveSegmentDown(_element, index) {
      if (index >= this.flag.segments.length - 1) return
      const arr = this.flag.segments
      const temp = arr[index + 1]
      arr[index + 1] = arr[index]
      arr[index] = temp
    },


    handleUpdateSegmentField({ segment, field, value }) {
      segment[field] = value
    },

    // --- Constraints ---
    createConstraint(segment) {
      const c = segment._newConstraint
      c.property = c.property.trim()
      c.value = c.value.trim()
      Axios.post(
        `${API_URL}/flags/${this.flagId}/segments/${segment.id}/constraints`,
        c
      ).then(response => {
        segment.constraints.push(response.data)
        segment._newConstraint = clone(DEFAULT_CONSTRAINT)
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
          const idx = segment.constraints.findIndex(c => c.id === constraint.id)
          segment.constraints.splice(idx, 1)
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
      const label = tab.props?.label || tab.label
      if (label == "History" && !this.historyLoaded) {
        this.historyLoaded = true
      }
    },

    // --- Data fetching ---
    fetchFlag() {
      Axios.get(`${API_URL}/flags/${this.flagId}`).then(response => {
        const flag = response.data
        flag.segments.forEach(s => processSegment(s))
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
  margin: 10px 0 5px;
}

.grabbable {
  cursor: move;
  cursor: grab;
  cursor: -moz-grab;
  cursor: -webkit-grab;
}

.flag-inner-config-card {
  .el-card__body {
    padding-bottom: 0px;
  }
}

.segment {
  .highlightable {
    padding: 4px;
    &:hover {
      background-color: #ddd;
    }
  }
  .segment-constraint {
    margin-bottom: 12px;
    padding: 1px;
    background-color: #f6f6f6;
    border-radius: 5px;
  }
  .distribution-card {
    height: 110px;
    text-align: center;
    .el-card__body {
      padding: 3px 10px 10px 10px;
    }
    font-size: 0.9em;
  }
}

ol.constraints-inner {
  background-color: white;
  padding-left: 8px;
  padding-right: 8px;
  border-radius: 3px;
  border: 1px solid #ddd;
  li {
    padding: 3px 0;
    .el-tag {
      font-size: 0.7em;
    }
  }
}

.constraints-inputs-container {
  padding: 5px 0;
}

.variants-container-inner {
  .el-card {
    margin-bottom: 1em;
  }
  .el-input-group__prepend {
    width: 2em;
  }
}

.segment-description-rollout {
  margin-top: 10px;
}

.edit-distribution-button {
  margin-top: 5px;
}

.edit-distribution-alert {
  margin-top: 10px;
}

.el-form-item {
  margin-bottom: 5px;
}

.id-row {
  margin-bottom: 8px;
}

.flag-config-card {
  .flag-content {
    margin-top: 8px;
    margin-bottom: -8px;
    .el-input-group__prepend {
      width: 8em;
    }
  }
  .data-records-label {
    margin-left: 3px;
    margin-bottom: 5px;
    margin-top: 6px;
    font-size: 0.65em;
    white-space: nowrap;
    vertical-align: middle;
    display: inline-flex;
    align-items: center;
    gap: 2px;
  }
}

.variant-attachment-collapsable-title {
  margin: 0;
  font-size: 13px;
  color: #909399;
  width: 100%;
}

.variant-attachment-title {
  margin: 0;
  font-size: 13px;
  color: #909399;
}

.variant-key-input {
  margin-left: 10px;
  width: 50%;
}

.save-remove-variant-row {
  padding-bottom: 5px;
}

.tag-key-input {
  margin: 2.5px;
  width: 20%;
}

.tags-container-inner {
  margin-bottom: 10px;
}

.button-new-tag {
  margin: 2.5px;
}
</style>
