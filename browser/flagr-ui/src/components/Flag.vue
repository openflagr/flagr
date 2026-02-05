<template>
  <el-row>
    <el-col
      :span="20"
      :offset="2"
    >
      <div class="container flag-container">
        <el-dialog
          v-model="dialogDeleteFlagVisible"
          title="Delete feature flag"
        >
          <span>Are you sure you want to delete this feature flag?</span>
          <template #footer>
            <span class="dialog-footer">
              <el-button @click="dialogDeleteFlagVisible = false">Cancel</el-button>
              <el-button
                type="primary"
                @click.prevent="deleteFlag"
              >Confirm</el-button>
            </span>
          </template>
        </el-dialog>

        <el-dialog
          v-model="dialogEditDistributionOpen"
          title="Edit distribution"
        >
          <div v-if="loaded && flag">
            <div
              v-for="variant in flag.variants"
              :key="'distribution-variant-' + variant.id"
            >
              <div>
                <el-checkbox
                  :model-value="!!newDistributions[variant.id]"
                  @change="(e) => selectVariant(e, variant)"
                />
                <el-tag
                  type="danger"
                  :disable-transitions="true"
                >
                  {{ variant.key }}
                </el-tag>
              </div>
              <el-slider
                v-if="!newDistributions[variant.id]"
                :model-value="0"
                :disabled="true"
                show-input
              />
              <div v-if="!!newDistributions[variant.id]">
                <el-slider
                  v-model="newDistributions[variant.id].percent"
                  :disabled="false"
                  show-input
                />
              </div>
            </div>
          </div>
          <el-button
            class="width--full"
            :disabled="!newDistributionIsValid"
            @click.prevent="() => saveDistribution(selectedSegment)"
          >
            Save
          </el-button>

          <el-alert
            v-if="!newDistributionIsValid"
            class="edit-distribution-alert"
            :title="
              'Percentages must add up to 100% (currently at ' +
                newDistributionPercentageSum +
                '%)'
            "
            type="error"
            show-icon
          />
        </el-dialog>

        <el-dialog
          v-model="dialogCreateSegmentOpen"
          title="Create segment"
        >
          <div>
            <p>
              <el-input
                v-model="newSegment.description"
                placeholder="Segment description"
              />
            </p>
            <p>
              <el-slider
                v-model="newSegment.rolloutPercent"
                show-input
              />
            </p>
            <el-button
              class="width--full"
              :disabled="!newSegment.description"
              @click.prevent="createSegment"
            >
              Create Segment
            </el-button>
          </div>
        </el-dialog>

        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ name: 'home' }">
            Home page
          </el-breadcrumb-item>
          <el-breadcrumb-item>Flag ID: {{ route.params.flagId }}</el-breadcrumb-item>
        </el-breadcrumb>

        <div v-if="loaded && flag">
          <el-tabs @tab-click="handleHistoryTabClick">
            <el-tab-pane label="Config">
              <el-card class="flag-config-card">
                <template #header>
                  <div class="el-card-header">
                    <div class="flex-row">
                      <div class="flex-row-left">
                        <h2>Flag</h2>
                      </div>
                      <div
                        v-if="flag"
                        class="flex-row-right"
                      >
                        <el-tooltip
                          content="Enable/Disable Flag"
                          placement="top"
                          effect="light"
                        >
                          <el-switch
                            v-model="flag.enabled"
                            active-color="#13ce66"
                            inactive-color="#ff4949"
                            :active-value="true"
                            :inactive-value="false"
                            @change="setFlagEnabled"
                          />
                        </el-tooltip>
                      </div>
                    </div>
                  </div>
                </template>
                <el-card
                  shadow="hover"
                  :class="toggleInnerConfigCard"
                >
                  <div class="flex-row id-row">
                    <div class="flex-row-left">
                      <el-tag
                        type="primary"
                        :disable-transitions="true"
                      >
                        Flag ID: {{ route.params.flagId }}
                      </el-tag>
                    </div>
                    <div class="flex-row-right">
                      <el-button
                        size="small"
                        @click="putFlag(flag)"
                      >
                        Save Flag
                      </el-button>
                    </div>
                  </div>
                  <el-row
                    class="flag-content"
                    align="middle"
                  >
                    <el-col :span="17">
                      <el-row>
                        <el-col :span="24">
                          <el-input
                            v-model="flag.key"
                            size="small"
                            placeholder="Key"
                          >
                            <template #prepend>
                              Flag Key
                            </template>
                          </el-input>
                        </el-col>
                      </el-row>
                    </el-col>
                    <el-col
                      style="text-align: right;"
                      :span="5"
                    >
                      <div>
                        <el-switch
                          v-model="flag.dataRecordsEnabled"
                          size="small"
                          active-color="#74E5E0"
                          :active-value="true"
                          :inactive-value="false"
                        />
                      </div>
                    </el-col>
                    <el-col :span="2">
                      <div class="data-records-label">
                        Data Records
                        <el-tooltip
                          content="Controls whether to log to data pipeline, e.g. Kafka, Kinesis, Pubsub"
                          placement="top-end"
                          effect="light"
                        >
                          <el-icon><InfoFilled /></el-icon>
                        </el-tooltip>
                      </div>
                    </el-col>
                  </el-row>
                  <el-row
                    class="flag-content"
                    align="middle"
                  >
                    <el-col :span="17">
                      <el-row>
                        <el-col :span="24">
                          <el-input
                            v-model="flag.description"
                            size="small"
                            placeholder="Description"
                          >
                            <template #prepend>
                              Flag Description
                            </template>
                          </el-input>
                        </el-col>
                      </el-row>
                    </el-col>
                    <el-col
                      style="text-align: right;"
                      :span="5"
                    >
                      <div>
                        <el-select
                          v-show="!!flag.dataRecordsEnabled"
                          v-model="flag.entityType"
                          size="small"
                          filterable
                          :allow-create="allowCreateEntityType"
                          default-first-option
                          placeholder="<null>"
                        >
                          <el-option
                            v-for="item in entityTypes"
                            :key="item.value"
                            :label="item.label"
                            :value="item.value"
                          />
                        </el-select>
                      </div>
                    </el-col>
                    <el-col :span="2">
                      <div
                        v-show="!!flag.dataRecordsEnabled"
                        class="data-records-label"
                      >
                        Entity Type
                        <el-tooltip
                          content="Overrides the entityType in data records logging"
                          placement="top-end"
                          effect="light"
                        >
                          <el-icon><InfoFilled /></el-icon>
                        </el-tooltip>
                      </div>
                    </el-col>
                  </el-row>
                  <div style="margin: 10px;">
                    <h5>
                      <span style="margin-right: 10px;">Flag Notes</span>
                      <el-button
                        round
                        size="small"
                        @click="toggleShowMdEditor"
                      >
                        <el-icon v-if="!showMdEditor">
                          <Edit />
                        </el-icon>
                        <el-icon v-else>
                          <View />
                        </el-icon>
                        {{ !showMdEditor ? "edit" : "view" }}
                      </el-button>
                    </h5>
                  </div>
                  <div>
                    <markdown-editor
                      v-model:markdown="flag.notes"
                      :show-editor="showMdEditor"
                      @save="putFlag(flag)"
                    />
                  </div>
                  <div style="margin: 10px;">
                    <h5>
                      <span style="margin-right: 10px;">Tags</span>
                    </h5>
                  </div>
                  <div>
                    <div class="tags-container-inner">
                      <el-tag
                        v-for="tag in flag.tags"
                        :key="tag.id"
                        closable
                        @close="deleteTag(tag)"
                      >
                        {{ tag.value }}
                      </el-tag>
                      <div
                        v-if="tagInputVisible"
                        class="tag-key-input"
                      >
                        <el-autocomplete
                          ref="saveTagInput"
                          v-model="newTag.value"
                          size="small"
                          style="width: 100%"
                          :trigger-on-focus="false"
                          :fetch-suggestions="queryTags"
                          @select="createTag"
                          @keyup.enter="createTag"
                          @keyup.esc="cancelCreateTag"
                        />
                      </div>
                      <el-button
                        v-else
                        class="button-new-tag"
                        size="small"
                        @click="showTagInput"
                      >
                        + New Tag
                      </el-button>
                    </div>
                  </div>
                </el-card>
              </el-card>

              <el-card class="variants-container">
                <template #header>
                  <div class="clearfix">
                    <h2>Variants</h2>
                  </div>
                </template>
                <div
                  v-if="flag.variants.length"
                  class="variants-container-inner"
                >
                  <div
                    v-for="variant in flag.variants"
                    :key="variant.id"
                  >
                    <el-card shadow="hover">
                      <el-form
                        label-position="left"
                        label-width="100px"
                      >
                        <div class="flex-row id-row">
                          <el-tag
                            type="primary"
                            :disable-transitions="true"
                          >
                            Variant ID:
                            <b>{{ variant.id }}</b>
                          </el-tag>
                          <el-input
                            v-model="variant.key"
                            class="variant-key-input"
                            size="small"
                            placeholder="Key"
                          >
                            <template #prepend>
                              Key
                            </template>
                          </el-input>
                          <div class="flex-row-right save-remove-variant-row">
                            <el-button
                              size="small"
                              @click="putVariant(variant)"
                            >
                              Save Variant
                            </el-button>
                            <el-button
                              size="small"
                              @click="deleteVariant(variant)"
                            >
                              <el-icon><Delete /></el-icon>
                            </el-button>
                          </div>
                        </div>
                        <el-collapse class="flex-row">
                          <el-collapse-item
                            title="Variant attachment"
                            class="variant-attachment-collapsable-title"
                          >
                            <p
                              class="variant-attachment-title"
                            >
                              You can add JSON in key/value pairs format.
                            </p>
                            <JsonEditorVue
                              v-model="variant.attachment"
                              :mode="'text'"
                              :main-menu-bar="false"
                              :navigation-bar="false"
                              :status-bar="false"
                              class="variant-attachment-content"
                              @change="(content, prev, { contentErrors }) => handleAttachmentChange(variant, content, contentErrors)"
                            />
                          </el-collapse-item>
                        </el-collapse>
                      </el-form>
                    </el-card>
                  </div>
                </div>
                <div
                  v-else
                  class="card--error"
                >
                  No variants created for this feature flag yet
                </div>
                <div class="variants-input">
                  <div class="flex-row equal-width constraints-inputs-container">
                    <div>
                      <el-input
                        v-model="newVariant.key"
                        placeholder="Variant Key"
                      />
                    </div>
                  </div>
                  <el-button
                    class="width--full"
                    :disabled="!newVariant.key"
                    @click.prevent="createVariant"
                  >
                    Create Variant
                  </el-button>
                </div>
              </el-card>

              <el-card class="segments-container">
                <template #header>
                  <div class="el-card-header">
                    <div class="flex-row">
                      <div class="flex-row-left">
                        <h2>Segments</h2>
                      </div>
                      <div class="flex-row-right">
                        <el-tooltip
                          content="You can drag and drop segments to reorder"
                          placement="top"
                          effect="light"
                        >
                          <el-button @click="putSegmentsReorder(flag.segments)">
                            Reorder
                          </el-button>
                        </el-tooltip>
                        <el-button @click="dialogCreateSegmentOpen = true">
                          New Segment
                        </el-button>
                      </div>
                    </div>
                  </div>
                </template>
                <div
                  v-if="flag.segments.length"
                  class="segments-container-inner"
                >
                  <draggable
                    v-model="flag.segments"
                    item-key="id"
                    @start="drag = true"
                    @end="drag = false"
                  >
                    <template #item="{ element: segment }">
                      <el-card
                        shadow="hover"
                        class="segment grabbable"
                      >
                        <div class="flex-row id-row">
                          <div class="flex-row-left">
                            <el-tag
                              type="primary"
                              :disable-transitions="true"
                            >
                              Segment ID:
                              <b>{{ segment.id }}</b>
                            </el-tag>
                          </div>
                          <div class="flex-row-right">
                            <el-button
                              size="small"
                              @click="putSegment(segment)"
                            >
                              Save Segment Setting
                            </el-button>
                            <el-button
                              size="small"
                              @click="deleteSegment(segment)"
                            >
                              <el-icon><Delete /></el-icon>
                            </el-button>
                          </div>
                        </div>
                        <el-row
                          :gutter="10"
                          class="id-row"
                        >
                          <el-col :span="15">
                            <el-input
                              v-model="segment.description"
                              size="small"
                              placeholder="Description"
                            >
                              <template #prepend>
                                Description
                              </template>
                            </el-input>
                          </el-col>
                          <el-col :span="9">
                            <el-input
                              v-model="segment.rolloutPercent"
                              class="segment-rollout-percent"
                              size="small"
                              placeholder="0"
                              :min="0"
                              :max="100"
                            >
                              <template #prepend>
                                Rollout
                              </template>
                              <template #append>
                                %
                              </template>
                            </el-input>
                          </el-col>
                        </el-row>
                        <el-row>
                          <el-col :span="24">
                            <h5>Constraints (match ALL of them)</h5>
                            <div class="constraints">
                              <div
                                v-if="segment.constraints.length"
                                class="constraints-inner"
                              >
                                <div
                                  v-for="constraint in segment.constraints"
                                  :key="constraint.id"
                                >
                                  <el-row
                                    :gutter="3"
                                    class="segment-constraint"
                                  >
                                    <el-col :span="20">
                                      <el-input
                                        v-model="constraint.property"
                                        size="small"
                                        placeholder="Property"
                                      >
                                        <template #prepend>
                                          Property
                                        </template>
                                      </el-input>
                                    </el-col>
                                    <el-col :span="4">
                                      <el-select
                                        v-model="constraint.operator"
                                        class="width--full"
                                        size="small"
                                        placeholder="operator"
                                      >
                                        <el-option
                                          v-for="item in operatorOptions"
                                          :key="item.value"
                                          :label="item.label"
                                          :value="item.value"
                                        />
                                      </el-select>
                                    </el-col>
                                    <el-col :span="20">
                                      <el-input
                                        v-model="constraint.value"
                                        size="small"
                                      >
                                        <template #prepend>
                                          Value&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
                                        </template>
                                      </el-input>
                                    </el-col>
                                    <el-col :span="2">
                                      <el-button
                                        type="success"
                                        plain
                                        class="width--full"
                                        size="small"
                                        @click="
                                          putConstraint(segment, constraint)
                                        "
                                      >
                                        Save
                                      </el-button>
                                    </el-col>
                                    <el-col :span="2">
                                      <el-button
                                        type="danger"
                                        plain
                                        class="width--full"
                                        size="small"
                                        @click="
                                          deleteConstraint(segment, constraint)
                                        "
                                      >
                                        <el-icon><Delete /></el-icon>
                                      </el-button>
                                    </el-col>
                                  </el-row>
                                </div>
                              </div>
                              <div
                                v-else
                                class="card--empty"
                              >
                                <span>No constraints (ALL will pass)</span>
                              </div>
                              <div>
                                <el-row :gutter="3">
                                  <el-col :span="5">
                                    <el-input
                                      v-model="segment.newConstraint.property"
                                      size="small"
                                      placeholder="Property"
                                    />
                                  </el-col>
                                  <el-col :span="4">
                                    <el-select
                                      v-model="segment.newConstraint.operator"
                                      size="small"
                                      placeholder="operator"
                                    >
                                      <el-option
                                        v-for="item in operatorOptions"
                                        :key="item.value"
                                        :label="item.label"
                                        :value="item.value"
                                      />
                                    </el-select>
                                  </el-col>
                                  <el-col :span="11">
                                    <el-input
                                      v-model="segment.newConstraint.value"
                                      size="small"
                                    />
                                  </el-col>
                                  <el-col :span="4">
                                    <el-button
                                      class="width--full"
                                      size="small"
                                      type="primary"
                                      plain
                                      :disabled="
                                        !segment.newConstraint.property ||
                                          !segment.newConstraint.value
                                      "
                                      @click.prevent="
                                        () => createConstraint(segment)
                                      "
                                    >
                                      Add Constraint
                                    </el-button>
                                  </el-col>
                                </el-row>
                              </div>
                            </div>
                          </el-col>
                          <el-col
                            :span="24"
                            class="segment-distributions"
                          >
                            <h5>
                              <span>Distribution</span>
                              <el-button
                                round
                                size="small"
                                @click="editDistribution(segment)"
                              >
                                <el-icon><Edit /></el-icon> edit
                              </el-button>
                            </h5>
                            <el-row
                              v-if="segment.distributions.length"
                              :gutter="20"
                            >
                              <el-col
                                v-for="distribution in segment.distributions"
                                :key="distribution.id"
                                :span="6"
                              >
                                <el-card
                                  shadow="never"
                                  class="distribution-card"
                                >
                                  <div>
                                    <span size="small">
                                      {{
                                        distribution.variantKey
                                      }}
                                    </span>
                                  </div>
                                  <el-progress
                                    type="circle"
                                    color="#74E5E0"
                                    :width="70"
                                    :percentage="distribution.percent"
                                  />
                                </el-card>
                              </el-col>
                            </el-row>

                            <div
                              v-else
                              class="card--error"
                            >
                              No distribution yet
                            </div>
                          </el-col>
                        </el-row>
                      </el-card>
                    </template>
                  </draggable>
                </div>
                <div
                  v-else
                  class="card--error"
                >
                  No segments created for this feature flag yet
                </div>
              </el-card>
              <debug-console :flag="flag" />
              <el-card>
                <template #header>
                  <div class="el-card-header">
                    <h2>Flag Settings</h2>
                  </div>
                </template>
                <el-button
                  type="danger"
                  plain
                  @click="dialogDeleteFlagVisible = true"
                >
                  <el-icon><Delete /></el-icon>
                  Delete Flag
                </el-button>
              </el-card>
              <spinner v-if="!loaded" />
            </el-tab-pane>

            <el-tab-pane label="History">
              <flag-history
                v-if="historyLoaded"
                :flag-id="parseInt(route.params.flagId, 10)"
              />
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
    </el-col>
  </el-row>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import clone from "lodash.clone";
import draggable from "vuedraggable";
import Axios from "axios";
import JsonEditorVue from "json-editor-vue";
import { ElMessage } from "element-plus";

import constants from "@/constants";
import helpers from "@/helpers/helpers";
import Spinner from "@/components/Spinner";
import DebugConsole from "@/components/DebugConsole";
import FlagHistory from "@/components/FlagHistory";
import MarkdownEditor from "@/components/MarkdownEditor.vue";
import { operators } from "@/operators.json";

const { sum, pluck, handleErr } = helpers;
const { API_URL, FLAGR_UI_POSSIBLE_ENTITY_TYPES } = constants;

const DEFAULT_SEGMENT = {
  description: "",
  rolloutPercent: 50
};

const DEFAULT_CONSTRAINT = {
  operator: "EQ",
  property: "",
  value: ""
};

const DEFAULT_VARIANT = {
  key: ""
};

const DEFAULT_TAG = {
  value: ""
};

const DEFAULT_DISTRIBUTION = {
  bitmap: "",
  variantID: 0,
  variantKey: "",
  percent: 0
};

function processSegment(segment) {
  segment.newConstraint = clone(DEFAULT_CONSTRAINT);
}

function processVariant(variant) {
  if (typeof variant.attachment === "string") {
    variant.attachment = JSON.parse(variant.attachment);
  }
}

function handleAttachmentChange(variant, content, contentErrors) {
  variant.attachmentValid = !(contentErrors && contentErrors.parseError);
}

const route = useRoute();
const router = useRouter();

const loaded = ref(false);
const dialogDeleteFlagVisible = ref(false);
const dialogEditDistributionOpen = ref(false);
const dialogCreateSegmentOpen = ref(false);
const entityTypes = ref([]);
const allTags = ref([]);
const allowCreateEntityType = ref(true);
const tagInputVisible = ref(false);
const flag = ref({
  createdBy: "",
  dataRecordsEnabled: false,
  entityType: "",
  description: "",
  enabled: false,
  id: 0,
  key: "",
  tags: [],
  segments: [],
  updatedAt: "",
  variants: [],
  notes: ""
});
const newSegment = ref(clone(DEFAULT_SEGMENT));
const newVariant = ref(clone(DEFAULT_VARIANT));
const newTag = ref(clone(DEFAULT_TAG));
const selectedSegment = ref(null);
const newDistributions = reactive({});
const operatorOptions = operators;
const showMdEditor = ref(false);
const historyLoaded = ref(false);
const drag = ref(false);
const saveTagInput = ref(null);

const newDistributionPercentageSum = computed(() => {
  return sum(pluck(Object.values(newDistributions), "percent"));
});

const newDistributionIsValid = computed(() => {
  const percentageSum = sum(
    pluck(Object.values(newDistributions), "percent")
  );
  return percentageSum === 100;
});

const flagId = computed(() => {
  return route.params.flagId;
});

const toggleInnerConfigCard = computed(() => {
  if (!showMdEditor.value && !flag.value.notes) {
    return "flag-inner-config-card";
  } else {
    return "";
  }
});

function deleteFlag() {
  const id = flagId.value;
  Axios.delete(`${API_URL}/flags/${flagId.value}`).then(() => {
    router.replace({ name: "home" });
    ElMessage.success(`You deleted flag ${id}`);
  }, handleErr);
}

function putFlag(f) {
  Axios.put(`${API_URL}/flags/${flagId.value}`, {
    description: f.description,
    dataRecordsEnabled: f.dataRecordsEnabled,
    key: f.key || "",
    entityType: f.entityType || "",
    notes: f.notes || ""
  }).then(() => {
    ElMessage.success(`Flag updated`);
  }, handleErr);
}

function setFlagEnabled(checked) {
  Axios.put(`${API_URL}/flags/${flagId.value}/enabled`, {
    enabled: checked
  }).then(() => {
    const checkedStr = checked ? "on" : "off";
    ElMessage.success(`You turned ${checkedStr} this feature flag`);
  }, handleErr);
}

function selectVariant($event, variant) {
  const checked = $event;
  if (checked) {
    const distribution = Object.assign(clone(DEFAULT_DISTRIBUTION), {
      variantKey: variant.key,
      variantID: variant.id
    });
    newDistributions[variant.id] = distribution;
  } else {
    delete newDistributions[variant.id];
  }
}

function editDistribution(segment) {
  selectedSegment.value = segment;

  // Clear all keys from reactive object
  Object.keys(newDistributions).forEach(key => delete newDistributions[key]);

  segment.distributions.forEach(distribution => {
    newDistributions[distribution.variantID] = clone(distribution);
  });

  dialogEditDistributionOpen.value = true;
}

function saveDistribution(segment) {
  const distributions = Object.values(newDistributions).filter(
    distribution => distribution.percent !== 0
  ).map(distribution => {
    let dist = clone(distribution)
    delete dist.id;
    return dist
  });

  Axios.put(
    `${API_URL}/flags/${flagId.value}/segments/${segment.id}/distributions`,
    { distributions }
  ).then(response => {
    let distributions = response.data;
    selectedSegment.value.distributions = distributions;
    dialogEditDistributionOpen.value = false;
    ElMessage.success("distributions updated");
  }, handleErr);
}

function createVariant() {
  Axios.post(
    `${API_URL}/flags/${flagId.value}/variants`,
    newVariant.value
  ).then(response => {
    let variant = response.data;
    newVariant.value = clone(DEFAULT_VARIANT);
    flag.value.variants.push(variant);
    ElMessage.success("new variant created");
  }, handleErr);
}

function deleteVariant(variant) {
  const isVariantInUse = flag.value.segments.some(segment =>
    segment.distributions.some(
      distribution => distribution.variantID === variant.id
    )
  );

  if (isVariantInUse) {
    alert(
      "This variant is being used by a segment distribution. Please remove the segment or edit the distribution in order to remove this variant."
    );
    return;
  }

  if (
    !confirm(
      `Are you sure you want to delete variant #${variant.id} [${variant.key}]`
    )
  ) {
    return;
  }

  Axios.delete(
    `${API_URL}/flags/${flagId.value}/variants/${variant.id}`
  ).then(() => {
    ElMessage.success("variant deleted");
    fetchFlag();
  }, handleErr);
}

function putVariant(variant) {
  if (variant.attachmentValid === false) {
    ElMessage.error("variant attachment is not valid");
    return;
  }

  // Prepare payload - parse attachment if it's a string (from text mode editor)
  let payload = { ...variant };
  if (typeof payload.attachment === "string") {
    try {
      payload.attachment = JSON.parse(payload.attachment);
    } catch {
      ElMessage.error("variant attachment is not valid JSON");
      return;
    }
  }

  Axios.put(
    `${API_URL}/flags/${flagId.value}/variants/${variant.id}`,
    payload
  ).then(() => {
    ElMessage.success("variant updated");
  }, handleErr);
}

function createTag() {
  Axios.post(`${API_URL}/flags/${flagId.value}/tags`, newTag.value).then(
    response => {
      let tag = response.data;
      newTag.value = clone(DEFAULT_TAG);
      if (!flag.value.tags.map(tag => tag.value).includes(tag.value)) {
        flag.value.tags.push(tag);
        ElMessage.success("new tag created");
      }
      tagInputVisible.value = false;
      loadAllTags();
    },
    handleErr
  );
}

function cancelCreateTag() {
  newTag.value = clone(DEFAULT_TAG);
  tagInputVisible.value = false;
}

function queryTags(queryString, cb) {
  let results = allTags.value.filter(tag =>
    tag.value.toLowerCase().includes(queryString.toLowerCase())
  );
  cb(results);
}

function loadAllTags() {
  Axios.get(`${API_URL}/tags`).then(response => {
    let result = response.data;
    allTags.value = result;
  }, handleErr);
}

function showTagInput() {
  tagInputVisible.value = true;
  nextTick(() => {
    if (saveTagInput.value && saveTagInput.value.focus) {
      saveTagInput.value.focus();
    }
  });
}

function deleteTag(tag) {
  if (!confirm(`Are you sure you want to delete tag #${tag.value}`)) {
    return;
  }

  Axios.delete(`${API_URL}/flags/${flagId.value}/tags/${tag.id}`).then(
    () => {
      ElMessage.success("tag deleted");
      fetchFlag();
      loadAllTags();
    },
    handleErr
  );
}

function createConstraint(segment) {
  segment.newConstraint.property = segment.newConstraint.property.trim();
  segment.newConstraint.value = segment.newConstraint.value.trim();
  Axios.post(
    `${API_URL}/flags/${flagId.value}/segments/${segment.id}/constraints`,
    segment.newConstraint
  ).then(response => {
    let constraint = response.data;
    segment.constraints.push(constraint);
    segment.newConstraint = clone(DEFAULT_CONSTRAINT);
    ElMessage.success("new constraint created");
  }, handleErr);
}

function putConstraint(segment, constraint) {
  constraint.property = constraint.property.trim();
  constraint.value = constraint.value.trim();
  Axios.put(
    `${API_URL}/flags/${flagId.value}/segments/${segment.id}/constraints/${constraint.id}`,
    constraint
  ).then(() => {
    ElMessage.success("constraint updated");
  }, handleErr);
}

function deleteConstraint(segment, constraint) {
  if (!confirm("Are you sure you want to delete this constraint?")) {
    return;
  }

  Axios.delete(
    `${API_URL}/flags/${flagId.value}/segments/${segment.id}/constraints/${constraint.id}`
  ).then(() => {
    const index = segment.constraints.findIndex(
      c => c.id === constraint.id
    );
    segment.constraints.splice(index, 1);
    ElMessage.success("constraint deleted");
  }, handleErr);
}

function putSegment(segment) {
  Axios.put(`${API_URL}/flags/${flagId.value}/segments/${segment.id}`, {
    description: segment.description,
    rolloutPercent: parseInt(segment.rolloutPercent, 10)
  }).then(() => {
    ElMessage.success("segment updated");
  }, handleErr);
}

function putSegmentsReorder(segments) {
  Axios.put(`${API_URL}/flags/${flagId.value}/segments/reorder`, {
    segmentIDs: pluck(segments, "id")
  }).then(() => {
    ElMessage.success("segment reordered");
  }, handleErr);
}

function deleteSegment(segment) {
  if (!confirm("Are you sure you want to delete this segment?")) {
    return;
  }

  Axios.delete(
    `${API_URL}/flags/${flagId.value}/segments/${segment.id}`
  ).then(() => {
    const index = flag.value.segments.findIndex(el => el.id === segment.id);
    flag.value.segments.splice(index, 1);
    ElMessage.success("segment deleted");
  }, handleErr);
}

function createSegment() {
  Axios.post(
    `${API_URL}/flags/${flagId.value}/segments`,
    newSegment.value
  ).then(response => {
    let segment = response.data;
    processSegment(segment);
    segment.constraints = [];
    newSegment.value = clone(DEFAULT_SEGMENT);
    flag.value.segments.push(segment);
    ElMessage.success("new segment created");
    dialogCreateSegmentOpen.value = false;
  }, handleErr);
}

function fetchFlag() {
  Axios.get(`${API_URL}/flags/${flagId.value}`).then(response => {
    let f = response.data;
    f.segments.forEach(segment => processSegment(segment));
    f.variants.forEach(variant => processVariant(variant));
    flag.value = f;
    loaded.value = true;
  }, handleErr);
  fetchEntityTypes();
}

function fetchEntityTypes() {
  function prepareEntityTypes(entityTypes) {
    let arr = entityTypes.map(key => {
      let label = key === "" ? "<null>" : key;
      return { label: label, value: key };
    });
    if (entityTypes.indexOf("") === -1) {
      arr.unshift({ label: "<null>", value: "" });
    }
    return arr;
  }

  if (
    FLAGR_UI_POSSIBLE_ENTITY_TYPES &&
    FLAGR_UI_POSSIBLE_ENTITY_TYPES != "null"
  ) {
    let types = FLAGR_UI_POSSIBLE_ENTITY_TYPES.split(",");
    entityTypes.value = prepareEntityTypes(types);
    allowCreateEntityType.value = false;
    return;
  }

  Axios.get(`${API_URL}/flags/entity_types`).then(response => {
    entityTypes.value = prepareEntityTypes(response.data);
  }, handleErr);
}

function toggleShowMdEditor() {
  showMdEditor.value = !showMdEditor.value;
}

function handleHistoryTabClick(tab) {
  if (tab.props.label == "History" && !historyLoaded.value) {
    historyLoaded.value = true;
  }
}

onMounted(() => {
  fetchFlag();
  loadAllTags();
});
</script>

<style lang="less">
h5 {
  padding: 0;
  margin: 10px 0 5px;
}

.grabbable {
  cursor: move; /* fallback if grab cursor is unsupported */
  cursor: grab;
  cursor: -moz-grab;
  cursor: -webkit-grab;
}

.segments-container-inner .segment {
  transition: transform 0.3s;
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
