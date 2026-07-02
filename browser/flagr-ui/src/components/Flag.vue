<template>
  <div class="container flag-container">
    <el-dialog
      v-model="dialogDeleteFlagVisible"
      title="Delete Flag"
    >
      <span>Are you sure you want to delete this feature flag?</span>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogDeleteFlagVisible = false">Cancel</el-button>
          <el-button
            type="primary"
            @click.prevent="flagPage.deleteFlag(page)"
          >Confirm</el-button>
        </span>
      </template>
    </el-dialog>
    <el-dialog
      v-model="dialogDuplicateFlagVisible"
      title="Duplicate Flag"
    >
      <span>{{ flagPage.DUPLICATE_FLAG_CONFIRM_MESSAGE }}</span>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogDuplicateFlagVisible = false">Cancel</el-button>
          <el-button
            type="primary"
            data-testid="confirm-duplicate-flag-btn"
            :disabled="page.duplicateInFlight"
            @click.prevent="flagPage.duplicateFlag(page)"
          >Confirm</el-button>
        </span>
      </template>
    </el-dialog>

    <el-dialog
      v-model="dialogCreateSegmentOpen"
      title="Create segment"
    >
      <div class="create-segment-dialog">
        <el-input
          v-model="newSegment.description"
          placeholder="Segment description"
          data-testid="new-segment-desc-input"
        />
        <div class="create-segment-slider">
          <label class="create-segment-label">Rollout %</label>
          <el-slider
            v-model="newSegment.rolloutPercent"
            show-input
            :max="100"
          />
        </div>
        <el-button
          class="width--full"
          type="primary"
          :disabled="!newSegment.description"
          data-testid="create-segment-btn"
          @click.prevent="flagPage.createSegment(page)"
        >
          Create Segment
        </el-button>
      </div>
    </el-dialog>

    <distribution-dialog
      :visible="dialogEditDistributionOpen"
      :flag="flag"
      :initial-distributions="distributionDraft"
      @update:visible="dialogEditDistributionOpen = $event"
      @save="(d) => flagPage.handleSaveDistribution(page, d)"
    />

    <el-breadcrumb separator="/">
      <el-breadcrumb-item :to="{ name: 'home' }">
        Home page
      </el-breadcrumb-item>
      <el-breadcrumb-item>Flag ID: {{ $route.params.flagId }}</el-breadcrumb-item>
    </el-breadcrumb>

    <div v-if="loaded && flag">
      <el-tabs @tab-click="onHistoryTabClick">
        <el-tab-pane label="Config">
          <flag-config-card
            :flag="flag"
            :show-md-editor="showMdEditor"
            :entity-types="entityTypes"
            :allow-create-entity-type="allowCreateEntityType"
            :tag-input-visible="tagInputVisible"
            :all-tags="allTags"
            @toggle-enabled="(c) => flagPage.handleToggleEnabled(page, c)"
            @save-flag="flagPage.putFlag(page)"
            @update-flag="(p) => flagPage.handleUpdateFlag(page, p)"
            @toggle-notes="showMdEditor = !showMdEditor"
            @delete-tag="(tag) => flagPage.deleteTag(page, tag)"
            @create-tag="(p) => flagPage.handleCreateTag(page, p)"
            @cancel-create-tag="flagPage.handleCancelCreateTag(page)"
            @show-tag-input="flagPage.handleShowTagInput(page)"
          />

          <variants-section
            :variants="flag.variants"
            @create-variant="(p) => flagPage.handleCreateVariant(page, p)"
            @update-variant-key="(p) => flagPage.handleUpdateVariantKey(page, p)"
            @save-variant="(v) => flagPage.putVariant(page, v)"
            @delete-variant="(v) => flagPage.deleteVariant(page, v)"
            @attachment-change="(p) => flagPage.handleVariantAttachmentChange(page, p)"
          />

          <segments-section
            :segments="flag.segments ?? []"
            :operator-options="operatorOptions"
            @reorder="(s) => flagPage.handleReorderSegments(page, s)"
            @move-up="(el, i) => flagPage.moveSegmentUp(page, el, i)"
            @move-down="(el, i) => flagPage.moveSegmentDown(page, el, i)"
            @new-segment="dialogCreateSegmentOpen = true"
            @save-segment="(s) => flagPage.putSegment(page, s)"
            @delete-segment="(s) => flagPage.deleteSegment(page, s)"
            @update-segment-field="(p) => flagPage.handleUpdateSegmentField(page, p)"
            @create-constraint="(p) => flagPage.createConstraint(page, p)"
            @save-constraint="(p) => flagPage.putConstraint(page, p)"
            @delete-constraint="(p) => flagPage.deleteConstraint(page, p)"
            @update-constraint-field="(p) => flagPage.handleUpdateConstraintField(page, p)"
            @edit-distribution="(s) => flagPage.handleEditDistribution(page, s)"
          />

          <debug-console
            :eval-context="evalContext"
            :eval-result="evalResult"
            :eval-summary="evalSummary"
            :batch-eval-context="batchEvalContext"
            :batch-eval-result="batchEvalResult"
            @update:eval-context="evalContext = $event"
            @update:eval-result="evalResult = $event"
            @update:batch-eval-context="batchEvalContext = $event"
            @update:batch-eval-result="batchEvalResult = $event"
            @post-evaluation="(ctx) => flagPage.postEvaluation(page, ctx)"
            @post-evaluation-batch="(ctx) => flagPage.postEvaluationBatch(page, ctx)"
          />

          <el-card
            class="flag-management-card"
            style="margin-top: var(--space-xl);"
          >
            <template #header>
              <div class="el-card-header">
                <h2>Flag Management</h2>
              </div>
            </template>
            <div class="flag-management-body">
              <div class="flag-management-row">
                <p class="flag-management-text">
                  Create a copy of this flag with the same segments, variants, constraints, distributions, and tags. The new flag gets its own ID and key.
                </p>
                <el-button
                  type="primary"
                  plain
                  size="small"
                  data-testid="duplicate-flag-btn"
                  @click="dialogDuplicateFlagVisible = true"
                >
                  Duplicate Flag
                </el-button>
              </div>
              <div class="flag-management-row flag-management-row--delete">
                <p class="flag-management-text">
                  Deleting hides this flag from evaluation. Segments, variants, and distributions are kept; restore it from <strong>Deleted flags</strong> on the flags list.
                </p>
                <el-button
                  type="danger"
                  plain
                  size="small"
                  data-testid="delete-flag-btn"
                  @click="dialogDeleteFlagVisible = true"
                >
                  <el-icon><Delete /></el-icon>
                  Delete Flag
                </el-button>
              </div>
            </div>
          </el-card>
        </el-tab-pane>

        <el-tab-pane
          label="History"
          name="history"
        >
          <flag-history
            v-if="historyLoaded"
            :key="historyKey"
            :snapshots="flagSnapshots"
          />
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
import type { BatchEvalContext, BatchEvalResult, DistributionDraft, EvalContext, EvalResult, EvalSummary, FlagView, Segment, Tag } from '@/api/types'
import type { EntityTypeOption } from '@/helpers/flagModel'
import { castFlagPage } from '@/helpers/vuePageCast'
import { handleHistoryTabClick, mountFlagPage } from '@/pages/flagPage'
import * as flagPage from '@/pages/flagPage'
import { OPERATOR_UI_OPTIONS } from '@/helpers/constraintOperators'

function defaultEvalContext(): EvalContext {
  return {
    entityID: 'a1234',
    entityType: 'report',
    entityContext: { hello: 'world' },
    enableDebug: true,
  }
}

function defaultBatchEvalContext(): BatchEvalContext {
  return {
    entities: [
      { entityID: 'a1234', entityType: 'report', entityContext: { hello: 'world' } },
      { entityID: 'a5678', entityType: 'report', entityContext: { hello: 'world' } },
    ],
    enableDebug: true,
    flagIDs: [],
  }
}

export default {
  name: 'Flag',
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
      loaded: false,
      flagId: '',
      flagPageLoadGen: 0,
      duplicateInFlight: false,
      dialogDeleteFlagVisible: false,
      dialogDuplicateFlagVisible: false,
      dialogEditDistributionOpen: false,
      dialogCreateSegmentOpen: false,
      flagPage,
      entityTypes: [] as EntityTypeOption[],
      allTags: [] as Tag[],
      allowCreateEntityType: true,
      tagInputVisible: false,
      flag: { description: '', tags: [], variants: [], segments: [] } as FlagView,
      newSegment: { ...flagPage.DEFAULT_SEGMENT },
      newTag: { ...flagPage.DEFAULT_TAG },
      selectedSegment: null as Segment | null,
      distributionDraft: {} as Record<string, DistributionDraft>,
      operatorOptions: OPERATOR_UI_OPTIONS,
      showMdEditor: false,
      historyLoaded: false,
      historyKey: 0,
      flagSnapshots: [],
      evalContext: defaultEvalContext(),
      evalResult: {} as EvalResult,
      evalSummary: null as EvalSummary | null,
      batchEvalContext: defaultBatchEvalContext(),
      batchEvalResult: { evaluationResults: [] } as BatchEvalResult,
    }
  },
  computed: {
    page() {
      return castFlagPage(this)
    },
  },

  watch: {
    // Initial load and flag switches: mountFlagPage → syncEvalContextFromFlag (not mounted-only).
    '$route.params.flagId': {
      immediate: true,
      handler(id: string | string[] | undefined) {
        this.flagId = String(id ?? '')
        if (this.flagId) {
          mountFlagPage(this.page)
        }
      },
    },
  },
  methods: {
    onHistoryTabClick(tab: { props?: { name?: string } }) {
      handleHistoryTabClick(this.page, tab)
    },
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


.flag-management-card {
  margin-bottom: 0;
  .flag-management-body {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }
  .flag-management-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-md);
    flex-wrap: wrap;
  }
  .flag-management-row--delete {
    padding-top: var(--space-md);
    border-top: 1px solid var(--el-border-color-lighter);
  }
  .flag-management-text {
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