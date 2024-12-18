<template>
  <el-row>
    <el-col :span="20" :offset="2">
      <div class="container flag-container">
        <el-dialog title="Delete feature flag" v-model="dialogDeleteFlagVisible">
          <span>Are you sure you want to delete this feature flag?</span>
          <template v-slot:footer>
            <div class="dialog-footer">
              <el-button @click="dialogDeleteFlagVisible = false">Cancel</el-button>
              <el-button type="primary" @click.prevent="deleteFlag">Confirm</el-button>
            </div>
          </template>
        </el-dialog>

        <el-dialog title="Edit distribution" v-model="dialogEditDistributionOpen">
          <div v-if="loaded && flag">
            <div v-for="variant in flag.variants" :key="'distribution-variant-' + variant.id">
              <div>
                <el-checkbox
                  @change="(e) => selectVariant(e, variant)"
                  :checked="!!newDistributions[variant.id]"
                ></el-checkbox>
                <el-tag type="danger" :disable-transitions="true">{{ variant.key }}</el-tag>
              </div>
              <el-slider
                v-if="!newDistributions[variant.id]"
                :value="0"
                :disabled="true"
                show-input
              ></el-slider>
              <div v-if="!!newDistributions[variant.id]">
                <el-slider
                  v-model="newDistributions[variant.id].percent"
                  :disabled="false"
                  show-input
                ></el-slider>
              </div>
            </div>
          </div>
          <el-button
            class="width--full"
            :disabled="!newDistributionIsValid"
            @click.prevent="() => saveDistribution(selectedSegment)"
          >Save</el-button>

          <el-alert
            class="edit-distribution-alert"
            v-if="!newDistributionIsValid"
            :title="
              'Percentages must add up to 100% (currently at ' +
              newDistributionPercentageSum +
              '%)'
            "
            type="error"
            show-icon
          ></el-alert>
        </el-dialog>

        <el-dialog :title="'Create ' + currentTerms.segment" v-model="dialogCreateSegmentOpen">
          <div>
            <p>
              <el-input :placeholder="currentTerms.segment + ' description'" v-model="newSegment.description"></el-input>
            </p>
            <p>
              <el-slider v-model="newSegment.rolloutPercent" :disabled="!isModeAB" show-input></el-slider>
            </p>
            <el-button
              class="width--full"
              :disabled="!newSegment.description"
              @click.prevent="createSegment"
            >Create {{ currentTerms.segment }}</el-button>
          </div>
        </el-dialog>

        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ name: 'home' }">Home page</el-breadcrumb-item>
          <el-breadcrumb-item>{{currentTerms.flag}} ID: {{ $route.params.flagId }}</el-breadcrumb-item>
        </el-breadcrumb>
        <div v-if="flag">
          <el-tabs>
            <el-tab-pane label="Config">
              <el-card class="flag-config-card">
                <template v-slot:header>
                  <div class="el-card-header">
                    <div class="flex-row">
                      <div class="flex-row-left">
                        <h2>{{currentTerms.flag}}</h2>
                      </div>
                      <div class="flex-row-right" v-if="flag">
                        <el-tooltip :content="'Enable/Disable ' + currentTerms.flag" placement="top" effect="light">
                          <el-switch
                            v-model="flag.enabled"
                            active-color="#13ce66"
                            inactive-color="#ff4949"
                            @change="setFlagEnabled"
                            :active-value="true"
                            :inactive-value="false"
                          ></el-switch>
                        </el-tooltip>
                      </div>
                    </div>
                  </div>
                </template>
                <el-card shadow="hover" :class="toggleInnerConfigCard">
                  <div class="flex-row id-row">
                    <div class="flex-row-left">
                      <el-tag
                        type="primary"
                        :disable-transitions="true"
                      >{{currentTerms.flag}} ID: {{ $route.params.flagId }}</el-tag>
                    </div>
                    <div class="flex-row-right">
                      <el-button size="small" @click="putFlag(flag)">Save {{currentTerms.flag}}</el-button>
                    </div>
                  </div>
                  <el-row class="flag-content" type="flex" align="middle">
                    <el-col :span="17">
                      <el-row>
                        <el-col :span="24">
                          <el-input size="small" placeholder="Key" v-model="flag.key">
                            <template #prepend>{{currentTerms.flag}} Key</template>
                          </el-input>
                        </el-col>
                      </el-row>
                    </el-col>
                    <el-col style="text-align: right;" :span="5">
                      <div>
                        <el-switch
                          size="small"
                          v-model="flag.dataRecordsEnabled"
                          active-color="#74E5E0"
                          :active-value="true"
                          :inactive-value="false"
                        ></el-switch>
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
                          <ElIcon>
                            <InfoFilled />
                          </ElIcon>
                        </el-tooltip>
                      </div>
                    </el-col>
                  </el-row>
                  <el-row class="flag-content" type="flex" align="middle">
                    <el-col :span="17">
                      <el-row>
                        <el-col :span="24">
                          <el-input
                            size="small"
                            placeholder="Description"
                            v-model="flag.description"
                          >
                            <template #prepend>{{currentTerms.flag}} Description</template>
                          </el-input>
                        </el-col>
                      </el-row>
                    </el-col>
                    <el-col style="text-align: right;" :span="5">
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
                          ></el-option>
                        </el-select>
                      </div>
                    </el-col>
                    <el-col :span="2">
                      <div v-show="!!flag.dataRecordsEnabled" class="data-records-label">
                        Entity Type
                        <el-tooltip
                          content="Overrides the entityType in data records logging"
                          placement="top-end"
                          effect="light"
                        >
                          <span class="el-icon-info" />
                        </el-tooltip>
                      </div>
                    </el-col>
                  </el-row>
                  <el-row style="margin: 10px;">
                    <h5>
                      <span style="margin-right: 10px;">{{currentTerms.flag}} Notes</span>
                      <el-button round size="small" @click="toggleShowMdEditor">
                       <ElIcon v-if="!this.showMdEditor">
                         <Edit />
                       </ElIcon>
                       <ElIcon v-else>
                         <View />
                       </ElIcon>
                        {{ !this.showMdEditor ? "edit" : "view" }}
                      </el-button>
                    </h5>
                  </el-row>
                  <el-row v-if="loaded">
                    <MarkdownEditor
                      :showEditor="showMdEditor"
                      v-model:markdown="flag.notes"
                      @save="putFlag(flag)"
                    />
                  </el-row>
                  <el-row style="margin: 10px;">
                    <h5>
                      <span style="margin-right: 10px;">Tags</span>
                    </h5>
                  </el-row>
                  <el-row>
                    <div class="tags-container-inner">
                      <el-tag
                        v-for="tag in flag.tags"
                        :key="tag.id"
                        closable
                        :type="warning"
                        @close="deleteTag(tag)"
                      >{{tag.value}}</el-tag>
                      <el-autocomplete
                        class="tag-key-input"
                        v-if="tagInputVisible"
                        v-model="newTag.value"
                        ref="saveTagInput"
                        size="mini"
                        :trigger-on-focus="false"
                        :fetch-suggestions="queryTags"
                        @select="createTag"
                        @keyup.enter.native="createTag"
                        @keyup.esc.native="cancelCreateTag"
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

              <el-card class="variants-container" v-if="isModeAB">
                <template v-slot:header>
                  <div class="clearfix">
                    <h2>Variants</h2>
                  </div>
                </template>
                <div class="variants-container-inner" v-if="flag.variants.length">
                  <div v-for="variant in flag.variants" :key="variant.id">
                    <el-card shadow="hover">
                      <el-form label-position="left" label-width="100px">
                        <div class="flex-row id-row">
                          <el-tag type="primary" :disable-transitions="true">
                            Variant ID:
                            <b>{{ variant.id }}</b>
                          </el-tag>
                          <el-input
                            class="variant-key-input"
                            size="small"
                            placeholder="Key"
                            v-model="variant.key"
                          >
                            <template #prepend>Key</template>
                          </el-input>
                          <div class="flex-row-right save-remove-variant-row">
                            <el-button
                              size="small"
                              @click="putVariant(variant)"
                            >Save Variant</el-button>
                            <el-button @click="deleteVariant(variant)" size="small">
                              <ElIcon>
                                <Delete />
                              </ElIcon>
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
                            >You can add JSON in key/value pairs format.</p>
                            <Vue3JsonEditor
                              v-model="variant.attachment"
                              :showBtns="false"
                              :mode="'code'"
                              v-on:has-error="variant.attachmentValid = false"
                              v-on:input="variant.attachmentValid = true"
                              class="variant-attachment-content"
                            ></Vue3JsonEditor>
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
                      <el-input placeholder="Variant Key" v-model="newVariant.key"></el-input>
                    </div>
                  </div>
                  <el-button
                    class="width--full"
                    :disabled="!newVariant.key"
                    @click.prevent="createVariant"
                  >Create Variant</el-button>
                </div>
              </el-card>

              <el-card class="segments-container">
                <template v-slot:header>
                  <div class="el-card-header">
                    <div class="flex-row">
                      <div class="flex-row-left">
                        <h2>{{ plural(currentTerms.segment) }}</h2>
                      </div>
                      <div class="flex-row-right">
                        <el-tooltip
                          content="You can drag and drop segments to reorder"
                          placement="top"
                          effect="light"
                        >
                          <el-button v-if="isModeAB" @click="putSegmentsReorder(flag.segments)">Reorder</el-button>
                        </el-tooltip>
                        <el-button v-if="isModeAB" @click="dialogCreateSegmentOpen = true">New {{ currentTerms.segment }}</el-button>
                      </div>
                    </div>
                  </div>
                </template>
                <div class="segments-container-inner" v-if="flag.segments.length">
                  <draggable v-model="flag.segments" @start="drag = true" @end="drag = false">
                    <transition-group>
                      <el-card
                        shadow="hover"
                        v-for="segment in flag.segments"
                        :key="segment.id"
                        class="segment grabbable"
                      >
                        <div class="flex-row id-row">
                          <div class="flex-row-left">
                            <el-tag type="primary" :disable-transitions="true">
                              {{ currentTerms.segment }} ID:
                              <b>{{ segment.id }}</b>
                            </el-tag>
                          </div>
                          <div class="flex-row-right">
                            <el-button
                              size="small"
                              @click="putSegment(segment)"
                            >Save {{ currentTerms.segment }} Setting</el-button>
                            <el-button @click="deleteSegment(segment)" size="small" v-if="isModeAB">
                              <ElIcon>
                                <Delete />
                              </ElIcon>
                            </el-button>
                          </div>
                        </div>
                        <el-row :gutter="10" class="id-row">
                          <el-col :span="15">
                            <el-input
                              size="small"
                              placeholder="Description"
                              v-model="segment.description"
                            >
                              <template v-slot:prepend>Description</template>
                            </el-input>
                          </el-col>
                          <el-col :span="9">
                            <el-input
                              class="segment-rollout-percent"
                              size="small"
                              placeholder="0"
                              v-model="segment.rolloutPercent"
                              :disabled="!isModeAB"
                              :min="0"
                              :max="100"
                            >
                              <template v-slot:prepend>Rollout</template>
                              <template v-slot:append>%</template>
                            </el-input>
                          </el-col>
                        </el-row>
                        <el-row>
                          <el-col :span="24">
                            <h5>{{plural(currentTerms.constraint)}} (match ALL of them)</h5>
                            <div class="constraints">
                              <div class="constraints-inner" v-if="segment.constraints.length">
                                <div v-for="constraint in segment.constraints" :key="constraint.id">
                                  <el-row :gutter="3" class="segment-constraint">
                                    <el-col :span="20">
                                      <el-input
                                        size="small"
                                        placeholder="Property"
                                        v-model="constraint.property"
                                      >
                                        <template v-slot:prepend>Property</template>
                                      </el-input>
                                      <!-- <el-select
                                        size="small"
                                        v-model="constraint.property"
                                        placeholder="Property"
                                        v-if="!isModeAB"
                                      >
                                        <el-option
                                          v-for="item in latchOptions"
                                          :key="item.value"
                                          :label="item.label"
                                          :value="item.value"
                                        ></el-option>
                                      </el-select> -->
                                    </el-col>
                                    <el-col :span="4">
                                      <el-select
                                        class="width--full"
                                        size="small"
                                        v-model="constraint.operator"
                                        placeholder="operator"
                                      >
                                        <el-option
                                          v-for="item in operatorOptions"
                                          :key="item.value"
                                          :label="item.label"
                                          :value="item.value"
                                        ></el-option>
                                      </el-select>
                                    </el-col>
                                    <el-col :span="20">
                                      <el-input size="small" v-model="constraint.value">
                                        <template v-slot:prepend>Value&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</template>
                                      </el-input>
                                    </el-col>
                                    <el-col :span="2">
                                      <el-button
                                        type="success"
                                        plain
                                        class="width--full"
                                        @click="
                                          putConstraint(segment, constraint)
                                        "
                                        size="small"
                                      >Save</el-button>
                                    </el-col>
                                    <el-col :span="2">
                                      <el-button
                                        type="danger"
                                        plain
                                        class="width--full"
                                        @click="
                                          deleteConstraint(segment, constraint)
                                        "
                                        size="small"
                                      >
                                      <ElIcon>
                                        <Delete />
                                      </ElIcon>
                                      </el-button>
                                    </el-col>
                                  </el-row>
                                </div>
                              </div>
                              <div class="card--empty" v-else>
                                <span>No constraints (ALL will pass)</span>
                              </div>
                              <div>
                                <el-row :gutter="3">
                                  <el-col :span="5">
                                    <el-input
                                      size="small"
                                      placeholder="Property"
                                      v-model="segment.newConstraint.property"
                                    ></el-input>
                                  </el-col>
                                  <el-col :span="4">
                                    <el-select
                                      size="small"
                                      v-model="segment.newConstraint.operator"
                                      placeholder="operator"
                                    >
                                      <el-option
                                        v-for="item in operatorOptions"
                                        :key="item.value"
                                        :label="item.label"
                                        :value="item.value"
                                      ></el-option>
                                    </el-select>
                                  </el-col>
                                  <el-col :span="11">
                                    <el-input size="small" v-model="segment.newConstraint.value"></el-input>
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
                                    >Add {{ currentTerms.constraint }}</el-button>
                                  </el-col>
                                </el-row>
                              </div>
                            </div>
                          </el-col>
                          <el-col :span="24" class="segment-distributions" v-if="isModeAB">
                            <h5>
                              <span>Distribution</span>
                              <el-button round size="small" @click="editDistribution(segment)">
                                <ElIcon>
                                  <Edit />
                                </ElIcon> edit
                              </el-button>
                            </h5>
                            <el-row type="flex" v-if="segment.distributions.length" :gutter="20">
                              <el-col
                                v-for="distribution in segment.distributions"
                                :key="distribution.id"
                                :span="6"
                              >
                                <el-card shadow="never" class="distribution-card">
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
                                  ></el-progress>
                                </el-card>
                              </el-col>
                            </el-row>

                            <div class="card--error" v-else>No distribution yet</div>
                          </el-col>
                        </el-row>
                      </el-card>
                    </transition-group>
                  </draggable>
                </div>
                <div class="card--error" v-else>No segments created for this feature flag yet</div>
              </el-card>
              <!-- sdd -->
              <debug-console v-if="loaded" :flag="flag"></debug-console>
              <el-card>
                <template v-slot:header>
                  <div class="el-card-header">
                    <h2>{{currentTerms.flag}} Settings</h2>
                  </div>
                </template>
                <el-button @click="dialogDeleteFlagVisible = true" type="danger" plain>
                  <ElIcon>
                    <Delete />
                  </ElIcon>
                  Delete {{currentTerms.flag}}
                </el-button>
              </el-card>
              <spinner v-if="!loaded"></spinner>
            </el-tab-pane>

            <el-tab-pane label="History">
              <flag-history :flag-id="parseInt($route.params.flagId, 10)"></flag-history>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
    </el-col>
  </el-row>
</template>

<script>
import clone from "lodash.clone";
import { VueDraggableNext } from 'vue-draggable-next'
import constants from "@/constants";
import helpers from "@/helpers/helpers";
import Spinner from "@/components/Spinner";
import DebugConsole from "@/components/DebugConsole";
import FlagHistory from "@/components/FlagHistory";
import MarkdownEditor from "@/components/MarkdownEditor.vue";
import { Vue3JsonEditor } from 'vue3-json-editor'
import { operators, operators_latch, latch_options } from "@/operators.json";
import { ElIcon } from 'element-plus';
import { Edit, View, Delete, InfoFilled } from '@element-plus/icons';
import { mapState } from 'vuex';
import { getAxiosFlagrInstance } from "../utils/apiUtil";
import { MODES } from "../constants";
import { plural, reduceOperatorsToLabel } from '../utils/util'

const { sum, pluck, handleErr } = helpers;

const { FLAGR_UI_POSSIBLE_ENTITY_TYPES } = constants;

const DEFAULT_SEGMENT = {
  description: "",
  rolloutPercent: 50
};

const DEFAULT_SEGMENT_LATCH = {
  description: "",
  rolloutPercent: 100
};

const DEFAULT_CONSTRAINT = {
  operator: "EQ",
  property: "",
  value: ""
};

const DEFAULT_CONSTRAINT_LATCH = {
  operator: "IN",
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

function processSegment(segment, isModeAB) {
  const defaultConstraint = isModeAB ? DEFAULT_CONSTRAINT : DEFAULT_CONSTRAINT_LATCH
  segment.newConstraint = clone(defaultConstraint);
}

function processVariant(variant) {
  if (typeof variant.attachment === "string") {
    variant.attachment = JSON.parse(variant.attachment);
  }
}

export default {
  name: "flag",
  components: {
    spinner: Spinner,
    debugConsole: DebugConsole,
    flagHistory: FlagHistory,
    draggable: VueDraggableNext,
    MarkdownEditor,
    Vue3JsonEditor,
    ElIcon,
    InfoFilled,
    Edit,
    Delete,
    View
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
      flag: {
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
      },
      newSegment: clone(DEFAULT_SEGMENT),
      newVariant: clone(DEFAULT_VARIANT),
      newTag: clone(DEFAULT_TAG),
      selectedSegment: null,
      newDistributions: {},
      operatorOptions: operators,
      latchOptions: latch_options,
      operatorValueToLabelMap: reduceOperatorsToLabel(operators),
      showMdEditor: false
    };
  },
  computed: {
    ...mapState({
      userDetails: (state) => state.userDetails, // Map userDetails from Vuex state
    }),
    isModeAB() {
      return this.$store.state.mode === MODES.ABMode;
    },
    currentMode() {
      return this.$store.state.mode;
    },
    currentTerms(){
      return this.$store.state.terms;
    },
    newDistributionPercentageSum() {
      return sum(pluck(Object.values(this.newDistributions), "percent"));
    },
    newDistributionIsValid() {
      const percentageSum = sum(
        pluck(Object.values(this.newDistributions), "percent")
      );
      return percentageSum === 100;
    },
    flagId() {
      return this.$route.params.flagId;
    },
    editViewIcon() {
      return {
        "el-icon-edit": !this.showMdEditor,
        "el-icon-view": this.showMdEditor
      };
    },
    toggleInnerConfigCard() {
      if (!this.showMdEditor && !this.flag.notes) {
        return "flag-inner-config-card";
      } else {
        return "";
      }
    }
  },
  methods: {
    deleteFlag() {
      const flagId = this.flagId;
      getAxiosFlagrInstance().delete(`/flags/${this.flagId}`).then(() => {
        this.$router.replace({ name: "home" });
        this.$message.success(`You deleted flag ${flagId}`);
      }, handleErr.bind(this));
    },
    putFlag(flag) {
      getAxiosFlagrInstance().put(`/flags/${this.flagId}`, {
        description: flag.description,
        dataRecordsEnabled: flag.dataRecordsEnabled,
        key: flag.key || "",
        entityType: flag.entityType || "",
        notes: flag.notes || ""
      }).then(() => {
        this.$message.success(this.currentTerms.flag + ` updated`);
      }, handleErr.bind(this));
    },
    setFlagEnabled(checked) {
      getAxiosFlagrInstance().put(`/flags/${this.flagId}/enabled`, {
        enabled: checked
      }).then(() => {
        const checkedStr = checked ? "on" : "off";
        this.$message.success(`You turned ${checkedStr} this feature flag`);
      }, handleErr.bind(this));
    },
    selectVariant($event, variant) {
      const checked = $event;
      if (checked) {
        const distribution = Object.assign(clone(DEFAULT_DISTRIBUTION), {
          variantKey: variant.key,
          variantID: variant.id
        });
        this.newDistributions[variant.id] = distribution;
      } else {
        delete this.newDistributions[variant.id]
        this.$delete(this.newDistributions, variant.id);
      }
    },
    editDistribution(segment) {
      this.selectedSegment = segment;

      this.newDistributions = {};

      segment.distributions.forEach(distribution => {
        this.newDistributions[distribution.variantID] = clone(distribution);
      });

      this.dialogEditDistributionOpen = true;
    },
    saveDistribution(segment) {
      const distributions = Object.values(this.newDistributions).filter(
        distribution => distribution.percent !== 0
      ).map(distribution => {
        let dist = clone(distribution)
        delete dist.id;
        return dist
      });

      getAxiosFlagrInstance().put(
        `/flags/${this.flagId}/segments/${segment.id}/distributions`,
        { distributions }
      ).then(response => {
        let distributions = response.data;
        this.selectedSegment.distributions = distributions;
        this.dialogEditDistributionOpen = false;
        this.$message.success("distributions updated");
      }, handleErr.bind(this));
    },
    createVariant() {
      getAxiosFlagrInstance().post(
        `/flags/${this.flagId}/variants`,
        this.newVariant
      ).then(response => {
        let variant = response.data;
        this.newVariant = clone(DEFAULT_VARIANT);
        this.flag.variants.push(variant);
        this.$message.success("new variant created");
      }, handleErr.bind(this));
    },
    deleteVariant(variant) {
      const isVariantInUse = this.flag.segments.some(segment =>
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

      getAxiosFlagrInstance().delete(
        `/flags/${this.flagId}/variants/${variant.id}`
      ).then(() => {
        this.$message.success("variant deleted");
        this.fetchFlag();
      }, handleErr.bind(this));
    },
    putVariant(variant) {
      if (variant.attachmentValid === false) {
        this.$message.error("variant attachment is not valid");
        return;
      }
      getAxiosFlagrInstance().put(
        `/flags/${this.flagId}/variants/${variant.id}`,
        variant
      ).then(() => {
        this.$message.success("variant updated");
      }, handleErr.bind(this));
    },
    createTag() {
      if((this.newTag.value.toLowerCase() === 'latch' && this.isModeAB) || (this.newTag.value.toLowerCase() === 'ab' && !this.isModeAB)) {
        this.$message.error("tag value cannot be "+this.newTag.value);
        return;
      }
      getAxiosFlagrInstance().post(`/flags/${this.flagId}/tags`, this.newTag).then(
        response => {
          let tag = response.data;
          this.newTag = clone(DEFAULT_TAG);
          if (!this.flag.tags.map(tag => tag.value).includes(tag.value)) {
            this.flag.tags.push(tag);
            this.$message.success("new tag created");
          }
          this.tagInputVisible = false;
          this.loadAllTags();
        },
        handleErr.bind(this)
      );
    },
    cancelCreateTag() {
      this.newTag = clone(DEFAULT_TAG);
      this.tagInputVisible = false;
    },
    queryTags(queryString, cb) {
      let results = this.allTags.filter(tag =>
        tag.value.toLowerCase().includes(queryString.toLowerCase())
      );
      cb(results);
    },
    plural(text) {
      // Use the utility function
      return plural(text);
    },
    loadAllTags() {
      getAxiosFlagrInstance().get(`/tags`).then(response => {
        let result = response.data;
        this.allTags = result;
      }, handleErr.bind(this));
    },
    showTagInput() {
      this.tagInputVisible = true;
      this.$nextTick(() => {
        this.$refs.saveTagInput.$refs.input.focus();
      });
    },
    deleteTag(tag) {
      if((tag.value.toLowerCase() === 'latch' && !this.isModeAB) || (tag.value.toLowerCase() === 'ab' && this.isModeAB)) {
        this.$message.error(tag.value +" tag cannot be deleted when in " + (this.isModeAB ? "AB": "Latch") + " mode");
        return;
      }
      if (!confirm(`Are you sure you want to delete tag #${tag.value}`)) {
        return;
      }

      getAxiosFlagrInstance().delete(`/flags/${this.flagId}/tags/${tag.id}`).then(
        () => {
          this.$message.success("tag deleted");
          this.fetchFlag();
          this.loadAllTags();
        },
        handleErr.bind(this)
      );
    },
    createConstraint(segment) {
      getAxiosFlagrInstance().post(
        `/flags/${this.flagId}/segments/${segment.id}/constraints`,
        segment.newConstraint
      ).then(response => {
        let constraint = response.data;
        segment.constraints.push(constraint);
        segment.newConstraint = clone(DEFAULT_CONSTRAINT);
        this.$message.success("new constraint created");
      }, handleErr.bind(this));
    },
    putConstraint(segment, constraint) {
      getAxiosFlagrInstance().put(
        `/flags/${this.flagId}/segments/${segment.id}/constraints/${constraint.id}`,
        constraint
      ).then(() => {
        this.$message.success("constraint updated");
      }, handleErr.bind(this));
    },
    deleteConstraint(segment, constraint) {
      if (!confirm("Are you sure you want to delete this constraint?")) {
        return;
      }

      getAxiosFlagrInstance().delete(
        `/flags/${this.flagId}/segments/${segment.id}/constraints/${constraint.id}`
      ).then(() => {
        const index = segment.constraints.findIndex(
          c => c.id === constraint.id
        );
        segment.constraints.splice(index, 1);
        this.$message.success("constraint deleted");
      }, handleErr.bind(this));
    },
    putSegment(segment) {
      getAxiosFlagrInstance().put(`/flags/${this.flagId}/segments/${segment.id}`, {
        description: segment.description,
        rolloutPercent: parseInt(segment.rolloutPercent, 10)
      }).then(() => {
        this.$message.success("segment updated");
      }, handleErr.bind(this));
    },
    putSegmentsReorder(segments) {
      getAxiosFlagrInstance().put(`/flags/${this.flagId}/segments/reorder`, {
        segmentIDs: pluck(segments, "id")
      }).then(() => {
        this.$message.success("segment reordered");
      }, handleErr.bind(this));
    },
    deleteSegment(segment) {
      if (!confirm("Are you sure you want to delete this segment?")) {
        return;
      }

      getAxiosFlagrInstance().delete(
        `/flags/${this.flagId}/segments/${segment.id}`
      ).then(() => {
        const index = this.flag.segments.findIndex(el => el.id === segment.id);
        this.flag.segments.splice(index, 1);
        this.$message.success("segment deleted");
      }, handleErr.bind(this));
    },
    createSegment() {
      getAxiosFlagrInstance().post(
        `/flags/${this.flagId}/segments`,
        this.newSegment
      ).then(response => {
        let segment = response.data;
        processSegment(segment, this.isModeAB);
        segment.constraints = [];
        this.newSegment = clone(DEFAULT_SEGMENT);
        this.flag.segments.push(segment);
        this.$message.success("new segment created");
        this.dialogCreateSegmentOpen = false;
      }, handleErr.bind(this));
    },
    fetchFlag() {
      getAxiosFlagrInstance().get(`/flags/${this.flagId}`).then(response => {
        let flag = response.data;
        flag.segments.forEach(segment => processSegment(segment));
        flag.variants.forEach(variant => processVariant(variant));
        this.flag = flag;
        this.loaded = true;
      }, handleErr.bind(this));
      this.fetchEntityTypes();
    },
    fetchEntityTypes() {
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
        let entityTypes = FLAGR_UI_POSSIBLE_ENTITY_TYPES.split(",");
        this.entityTypes = prepareEntityTypes(entityTypes);
        this.allowCreateEntityType = false;
        return;
      }

      getAxiosFlagrInstance().get(`/flags/entity_types`).then(response => {
        this.entityTypes = prepareEntityTypes(response.data);
      }, handleErr.bind(this));
    },
    toggleShowMdEditor() {
      this.showMdEditor = !this.showMdEditor;
    }
  },
  mounted() {
    this.newSegment = clone((() => {
      return this.isModeAB ? DEFAULT_SEGMENT : DEFAULT_SEGMENT_LATCH
    })());
    this.operatorOptions = this.isModeAB ? operators : operators_latch,
    this.operatorValueToLabelMap = reduceOperatorsToLabel(this.isModeAB ? operators : operators_latch),
    this.fetchFlag();
    this.loadAllTags();
  },
  watch: {
  }
};
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
  width: 100%;
}

.button-new-tag {
  margin: 2.5px;
}
</style>
