<template>
  <div class="container flag-container">
        <el-dialog title="Delete feature flag" v-model="dialogDeleteFlagVisible">
          <span>Are you sure you want to delete this feature flag?</span>
          <template #footer>
            <span class="dialog-footer">
              <el-button @click="dialogDeleteFlagVisible = false">Cancel</el-button>
              <el-button type="primary" @click.prevent="flagPage.deleteFlag(pageVm)">Confirm</el-button>
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
              @click.prevent="flagPage.createSegment(pageVm)"
              data-testid="create-segment-btn"
            >Create Segment</el-button>
          </div>
        </el-dialog>

        <distribution-dialog
          :visible="dialogEditDistributionOpen"
          :flag="flag"
          :initial-distributions="distributionDraft"
          @update:visible="dialogEditDistributionOpen = $event"
          @save="(d) => flagPage.handleSaveDistribution(pageVm, d)"
        />

        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ name: 'home' }">Home page</el-breadcrumb-item>
          <el-breadcrumb-item>Flag ID: {{ $route.params.flagId }}</el-breadcrumb-item>
        </el-breadcrumb>

        <div v-if="loaded && flag">
          <el-tabs @tab-click="onHistoryTab">
            <el-tab-pane label="Config">
              <flag-config-card
                :flag="flag"
                :show-md-editor="showMdEditor"
                :entity-types="entityTypes"
                :allow-create-entity-type="allowCreateEntityType"
                :tag-input-visible="tagInputVisible"
                :all-tags="allTags"
                @toggle-enabled="(c) => flagPage.handleToggleEnabled(pageVm, c)"
                @save-flag="flagPage.putFlag(pageVm)"
                @update-flag="(p) => flagPage.handleUpdateFlag(pageVm, p)"
                @toggle-notes="showMdEditor = !showMdEditor"
                @delete-tag="(tag) => flagPage.deleteTag(pageVm, tag)"
                @create-tag="(p) => flagPage.handleCreateTag(pageVm, p)"
                @cancel-create-tag="flagPage.handleCancelCreateTag(pageVm)"
                @show-tag-input="flagPage.handleShowTagInput(pageVm)"
              />

              <variants-section
                :variants="flag.variants"
                @create-variant="(p) => flagPage.handleCreateVariant(pageVm, p)"
                @update-variant-key="handleUpdateVariantKey"
                @save-variant="(v) => flagPage.putVariant(pageVm, v)"
                @delete-variant="(v) => flagPage.deleteVariant(pageVm, v)"
                @attachment-change="handleVariantAttachmentChange"
              />

              <segments-section
                :segments="flag.segments ?? []"
                :operator-options="operatorOptions"
                @reorder="(s) => flagPage.handleReorderSegments(pageVm, s)"
                @move-up="(el, i) => flagPage.moveSegmentUp(pageVm, el, i)"
                @move-down="(el, i) => flagPage.moveSegmentDown(pageVm, el, i)"
                @new-segment="dialogCreateSegmentOpen = true"
                @save-segment="(s) => flagPage.putSegment(pageVm, s)"
                @delete-segment="(s) => flagPage.deleteSegment(pageVm, s)"
                @update-segment-field="handleUpdateSegmentField"
                @create-constraint="(p) => flagPage.createConstraint(pageVm, p)"
                @save-constraint="(p) => flagPage.putConstraint(pageVm, p)"
                @delete-constraint="(p) => flagPage.deleteConstraint(pageVm, p)"
                @update-constraint-field="handleUpdateConstraintField"
                @edit-distribution="(s) => flagPage.handleEditDistribution(pageVm, s)"
              />

              <debug-console :flag="flag" />

              <el-card class="danger-zone-card is-card-danger" style="margin-top: var(--space-xl);">
                <template #header>
                  <div class="el-card-header">
                    <h2>Danger Zone</h2>
                  </div>
                </template>
                <div class="danger-zone-body">
                  <p class="danger-zone-text">
                    Deleting a flag hides it from active evaluation. Its segments, variants, and distributions stay intact and come back when you restore the flag from the Deleted Flags section on the flags list page.
                  </p>
                  <el-button type="danger" plain size="small" @click="dialogDeleteFlagVisible = true" data-testid="delete-flag-btn">
                    <el-icon><Delete /></el-icon>
                    Delete Flag
                  </el-button>
                </div>
              </el-card>
            </el-tab-pane>

            <el-tab-pane label="History" name="history">
              <flag-history v-if="historyLoaded" :key="historyKey" :flag-id="flagId"></flag-history>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
</template>

<script lang="ts">
import { Delete } from '@element-plus/icons-vue'
import DebugConsole from '@/components/DebugConsole.vue'
import DistributionDialog from '@/components/DistributionDialog.vue'
import FlagConfigCard from '@/components/FlagConfigCard.vue'
import FlagHistory from '@/components/FlagHistory.vue'
import SegmentsSection from '@/components/SegmentsSection.vue'
import VariantsSection from '@/components/VariantsSection.vue'
import * as flagPage from '@/pages/flagPage'
import {
  handleHistoryTabClick,
  handleUpdateConstraintField,
  handleUpdateSegmentField,
  handleUpdateVariantKey,
  handleVariantAttachmentChange,
  mountFlagPage,
  type FlagPageVm,
} from '@/pages/flagPage'
import operatorsData from '@/operators.json'
import type { DistributionDraft, FlagView, Segment, Tag } from '@/api/types'

const operators = operatorsData.operators

export default {
  name: 'flag',
  components: {
    DebugConsole,
    FlagHistory,
    DistributionDialog,
    FlagConfigCard,
    VariantsSection,
    SegmentsSection,
    Delete,
  },
  data() {
    return {
      flagPage,
      loaded: false,
      dialogDeleteFlagVisible: false,
      dialogEditDistributionOpen: false,
      dialogCreateSegmentOpen: false,
      entityTypes: [] as FlagPageVm['entityTypes'],
      allTags: [] as Tag[],
      allowCreateEntityType: true,
      tagInputVisible: false,
      flag: { description: '', variants: [], segments: [] } as FlagView,
      newSegment: { ...flagPage.DEFAULT_SEGMENT },
      newTag: { ...flagPage.DEFAULT_TAG },
      selectedSegment: null as Segment | null,
      distributionDraft: {} as Record<string, DistributionDraft>,
      operatorOptions: operators,
      showMdEditor: false,
      historyLoaded: false,
      historyKey: 0,
    }
  },
  computed: {
    flagId(): string {
      return String(this.$route.params.flagId)
    },
    pageVm(): FlagPageVm {
      return this as unknown as FlagPageVm
    },
  },
  methods: {
    onHistoryTab(tab: { props?: { name?: string } }) {
      handleHistoryTabClick(this.pageVm, tab)
    },
    handleUpdateVariantKey,
    handleVariantAttachmentChange,
    handleUpdateSegmentField,
    handleUpdateConstraintField,
  },
  mounted() {
    mountFlagPage(this.pageVm)
  },
}
</script>

<style lang="scss">
h5 {
  padding: 0;
  margin: var(--space-2xs) 0 var(--space-3xs);
  font-size: 13px;
}

.create-segment-dialog {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}
.create-segment-slider {
  display: flex;
  flex-direction: column;
  gap: var(--space-2xs);
}
.create-segment-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--el-text-color-secondary);
}

.danger-zone-card {
  margin-bottom: 0;
  .danger-zone-body {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-md);
    flex-wrap: wrap;
  }
  .danger-zone-text {
    margin: 0;
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
    flex: 1;
    min-width: 200px;
  }
}

.grabbable {
  cursor: move;
  cursor: grab;
  cursor: -moz-grab;
  cursor: -webkit-grab;
}
</style>