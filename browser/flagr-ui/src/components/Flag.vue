<template>
  <el-row>
    <el-col :span="14" :offset="5">
      <div class="container flag-container">
        <el-dialog
          title="Delete feature flag"
          :visible.sync="dialogDeleteFlagVisible"
        >
          <span>Are you sure you want to delete this feature flag?</span>
          <span slot="footer" class="dialog-footer">
            <el-button @click="dialogDeleteFlagVisible = false">Cancel</el-button>
            <el-button type="primary" @click.prevent="deleteFlag">Confirm</el-button>
          </span>
        </el-dialog>

        <el-dialog title="Edit distribution" :visible.sync="dialogEditDistributionOpen">
          <ul class="edit-distribution-choose-variants" v-if="loaded && flag">
            <li
              v-for="variant in flag.variants"
              :key="'distribution-variant-' + variant.id">
              <div>
                <el-checkbox
                  @change="(e) => selectVariant(e, variant)"
                  :checked="!!newDistributions[variant.id]"
                >
                </el-checkbox>
                <el-tag type="danger" :disable-transitions="true">
                  {{ variant.key }}
                </el-tag>
              </div>
              <el-slider
                v-if="!newDistributions[variant.id]"
                :value="0"
                :disabled="true"
                show-input>
              </el-slider>
              <div v-if="!!newDistributions[variant.id]">
                <el-slider
                  v-model="newDistributions[variant.id].percent"
                  :disabled="false"
                  show-input>
                </el-slider>
              </div>
            </li>
          </ul>
          <el-button
            class="width--full"
            :disabled="!newDistributionIsValid"
            @click.prevent="() => saveDistribution(selectedSegment)">
            Save
          </el-button>

          <el-alert
            class="edit-distribution-alert"
            v-if="!newDistributionIsValid"
            :title="'Percentages must add up to 100% (currently at ' + newDistributionPercentageSum + '%)'"
            type="error"
            show-icon>
          </el-alert>
        </el-dialog>

        <el-dialog title="Create segment" :visible.sync="dialogCreateSegmentOpen">
          <div>
            <p>
              <el-input
                placeholder="Segment description"
                v-model="newSegment.description">
              </el-input>
            </p>
            <p>
              <el-slider
                v-model="newSegment.rolloutPercent"
                show-input>
              </el-slider>
            </p>
            <el-button
              class="width--full"
              :disabled="!newSegment.description"
              @click.prevent="createSegment">
              Create Segment
            </el-button>
          </div>
        </el-dialog>

        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{name: 'home'}">Home page</el-breadcrumb-item>
          <el-breadcrumb-item>Flag ID: {{ $route.params.flagId }}</el-breadcrumb-item>
        </el-breadcrumb>

        <div v-if="loaded && flag">
          <el-tabs>
            <el-tab-pane label="Config">
              <el-card>
                <div slot="header" class="el-card-header">
                  <div class="flex-row">
                    <div class="flex-row-left">
                      <h2>Flag: {{ flag.name }}</h2>
                    </div>
                    <div class="flex-row-right" v-if="flag">
                      <el-tooltip content="Enable/Disable Flag" placement="top">
                        <el-switch
                          v-model="flag.enabled"
                          active-color="#13ce66"
                          inactive-color="#ff4949"
                          @change="setFlagEnabled"
                          :active-value="true"
                          :inactive-value="false">
                        </el-switch>
                      </el-tooltip>
                    </div>
                  </div>
                </div>
                <el-card>
                  <div class="flex-row id-row">
                    <div class="flex-row-left">
                      <el-tag type="primary" :disable-transitions="true">
                        Flag ID: {{ $route.params.flagId }}
                      </el-tag>
                    </div>
                    <div class="flex-row-right">
                      <el-button size="small" @click="putFlag(flag)">
                        Save Flag
                      </el-button>
                    </div>
                  </div>
                  <el-row :gutter="10" class="flag-content">
                    <el-col :span="17">
                      <el-input
                        size="small"
                        placeholder="Name"
                        v-model="flag.name"
                        :disabled="true">
                        <template slot="prepend">Flag Name</template>
                      </el-input>
                    </el-col>
                  </el-row>
                  <el-row :gutter="10" class="flag-content">
                    <el-col :span="17">
                      <el-input
                        v-model="flag.description"
                        size="small">
                        <template slot="prepend">Flag Description</template>
                      </el-input>
                    </el-col>
                    <el-col :span="7" style="text-align: right">
                      <el-tooltip content="Controls whether to log to data pipeline, e.g. Kafka" placement="top">
                        <el-switch
                          size="small"
                          v-model="flag.dataRecordsEnabled"
                          active-color="#74E5E0"
                          :active-value="true"
                          :inactive-value="false">
                        </el-switch>
                      </el-tooltip>
                      <span size="small">Data Records</span>
                    </el-col>
                  </el-row>
                </el-card>
              </el-card>

              <el-card class="variants-container">
                <div slot="header" class="clearfix">
                  <h2>Variants</h2>
                </div>
                <div class="variants-container-inner" v-if="flag.variants.length">
                  <div v-for="variant in flag.variants" :key="variant.id">
                    <el-card>
                      <el-form label-position="left" label-width="100px">
                        <div class="flex-row id-row">
                          <div class="flex-row-left">
                            <el-tag type="primary" :disable-transitions="true"> Variant ID: <b>{{ variant.id }}</b> </el-tag>
                          </div>
                          <div class="flex-row-right">
                            <el-button slot="append" size="small" @click="putVariant(variant)">
                              Save Variant
                            </el-button>
                            <el-button @click="deleteVariant(variant)" size="small">
                              <span class="el-icon-delete"/>
                            </el-button>
                          </div>
                        </div>
                        <el-input
                          placeholder="Key"
                          v-model="variant.key">
                          <template slot="prepend">Key&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</template>
                        </el-input>
                        <el-input
                          placeholder="{}"
                          v-model="variant.attachmentStr">
                          <template slot="prepend">Attachment </template>
                        </el-input>
                      </el-form>
                    </el-card>
                  </div>
                </div>
                <div class="card--error" v-else>
                  No variants created for this feature flag yet
                </div>
                <div class="variants-input">
                  <div class="flex-row equal-width constraints-inputs-container">
                    <div>
                      <el-input
                        placeholder="Variant Key"
                        v-model="newVariant.key">
                      </el-input>
                    </div>
                  </div>
                  <el-button
                    class="width--full"
                    :disabled="!newVariant.key"
                    @click.prevent="createVariant">
                    Create Variant
                  </el-button>
                </div>
              </el-card>

              <el-card class="segments-container">
                <div slot="header" class="el-card-header">
                  <div class="flex-row">
                    <div class="flex-row-left">
                      <h2>Segments</h2>
                    </div>
                    <div class="flex-row-right">
                      <el-tooltip content="You can drag and drop segments to reorder" placement="top">
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
                <div class="segments-container-inner" v-if="flag.segments.length">
                  <draggable v-model="flag.segments" @start="drag=true" @end="drag=false">
                    <transition-group>
                      <el-card
                        v-for="segment in flag.segments"
                        :key="segment.id"
                        class="segment grabbable">
                        <div class="flex-row id-row">
                          <div class="flex-row-left">
                            <el-tag type="primary" :disable-transitions="true">Segment ID: <b>{{ segment.id }}</b></el-tag>
                          </div>
                          <div class="flex-row-right">
                            <el-button slot="append" size="small" @click="putSegment(segment)">
                              Save Segment
                            </el-button>
                            <el-button @click="deleteSegment(segment)" size="small">
                              <span class="el-icon-delete"/>
                            </el-button>
                          </div>
                        </div>
                        <el-row :gutter="20" class="id-row">
                          <el-col :span="15">
                            <el-input
                              size="small"
                              placeholder="Description"
                              v-model="segment.description">
                              <template slot="prepend">Description</template>
                            </el-input>
                          </el-col>
                          <el-col :span="9">
                            <el-input
                              class="segment-rollout-percent"
                              size="small"
                              placeholder="0"
                              v-model="segment.rolloutPercent"
                              :min="0"
                              :max="100"
                              >
                              <template slot="prepend">Rollout</template>
                              <template slot="append">%</template>
                            </el-input>
                          </el-col>
                        </el-row>
                        <el-row>
                          <el-col :span="24">
                            <h4>Constraints (match ALL of them)</h4>
                            <div class="constraints">
                              <div class="constraints-inner" v-if="segment.constraints.length">
                                <div
                                  v-for="constraint in segment.constraints"
                                  :key="constraint.id">
                                  <el-row :gutter="3" class="segment-constraint">
                                    <el-col :span="20">
                                      <el-input
                                        size="small"
                                        placeholder="Property"
                                        v-model="constraint.property">
                                        <template slot="prepend">Property</template>
                                      </el-input>
                                    </el-col>
                                    <el-col :span="4">
                                      <el-select size="small" v-model="constraint.operator" placeholder="operator">
                                        <el-option
                                          v-for="item in operatorOptions"
                                          :key="item.value"
                                          :label="item.label"
                                          :value="item.value">
                                        </el-option>
                                      </el-select>
                                    </el-col>
                                    <el-col :span="20">
                                      <el-input
                                        size="small"
                                        placeholder='Value, e.g. "CA", ["CA", "NY"]'
                                        v-model="constraint.value">
                                        <template slot="prepend">Value&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</template>
                                      </el-input>
                                    </el-col>
                                    <el-col :span="2">
                                      <el-button class="width--full" @click="putConstraint(segment, constraint)" size="small">
                                        <span class="el-icon-check"/>
                                      </el-button>
                                    </el-col>
                                    <el-col :span="2">
                                      <el-button class="width--full" @click="deleteConstraint(segment, constraint)" size="small">
                                        <span class="el-icon-delete"/>
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
                                      v-model="segment.newConstraint.property">
                                    </el-input>
                                  </el-col>
                                  <el-col :span="4">
                                    <el-select size="small" v-model="segment.newConstraint.operator" placeholder="operator">
                                      <el-option
                                        v-for="item in operatorOptions"
                                        :key="item.value"
                                        :label="item.label"
                                        :value="item.value">
                                      </el-option>
                                    </el-select>
                                  </el-col>
                                  <el-col :span="11">
                                    <el-input
                                      size="small"
                                      placeholder='Value, e.g. "CA", ["CA", "NY"]'
                                      v-model="segment.newConstraint.value">
                                    </el-input>
                                  </el-col>
                                  <el-col :span="4">
                                    <el-button
                                      class="width--full"
                                      size="small"
                                      :disabled="!segment.newConstraint.property || !segment.newConstraint.value"
                                      @click.prevent="() => createConstraint(segment)">
                                      <span class="el-icon-plus"/>
                                    </el-button>
                                  </el-col>
                                </el-row>
                              </div>
                            </div>
                          </el-col>
                          <el-col :span="24" class="segment-distributions">
                            <h4>Distribution</h4>
                            <ul class="segment-distributions-inner" v-if="segment.distributions.length">
                              <li v-for="distribution in segment.distributions" :key="distribution.id">
                                <el-tag type="gray" :disable-transitions="true">{{ distribution.variantKey }}</el-tag>
                                <span size="small">{{ distribution.percent }}%</span>
                              </li>
                            </ul>
                            <div class="card--error" v-else>
                              No distribution yet
                            </div>
                            <div class="edit-distribution-button">
                              <el-button class="width--full" @click="editDistribution(segment)">
                                <span class="el-icon-edit"></span>
                                Edit distribution
                              </el-button>
                            </div>
                          </el-col>
                        </el-row>
                      </el-card>
                    </transition-group>
                  </draggable>
                </div>
                <div class="card--error" v-else>
                  No segments created for this feature flag yet
                </div>
              </el-card>
              <debug-console :flag-id="parseInt($route.params.flagId, 10)"></debug-console>
              <el-card>
                <div slot="header" class="el-card-header">
                  <h2>Flag Settings</h2>
                </div>
                <el-button @click="dialogDeleteFlagVisible = true">
                  <span class="el-icon-delete"></span>
                  Delete Flag
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
import clone from 'lodash.clone'
import draggable from 'vuedraggable'

import constants from '@/constants'
import helpers from '@/helpers/helpers'
import Spinner from '@/components/Spinner'
import DebugConsole from '@/components/DebugConsole'
import FlagHistory from '@/components/FlagHistory'
import {operators} from '@/../config/operators.json'

const OPERATOR_VALUE_TO_LABEL_MAP = operators.reduce((acc, el) => {
  acc[el.value] = el.label
  return acc
}, {})

const {
  sum,
  pluck
} = helpers

const {
  API_URL
} = constants

const DEFAULT_SEGMENT = {
  description: '',
  rolloutPercent: 50
}

const DEFAULT_CONSTRAINT = {
  operator: 'EQ',
  property: '',
  value: ''
}

const DEFAULT_VARIANT = {
  key: ''
}

const DEFAULT_DISTRIBUTION = {
  bitmap: '',
  variantID: 0,
  variantKey: '',
  percent: 0
}

function processSegment (segment) {
  segment.newConstraint = clone(DEFAULT_CONSTRAINT)
}

function processVariant (variant) {
  variant.attachmentStr = JSON.stringify(variant.attachment)
}

export default {
  name: 'flag',
  components: {
    spinner: Spinner,
    debugConsole: DebugConsole,
    flagHistory: FlagHistory,
    draggable: draggable
  },
  data () {
    return {
      loaded: false,
      dialogDeleteFlagVisible: false,
      dialogEditDistributionOpen: false,
      dialogCreateSegmentOpen: false,
      flag: {},
      flagId: this.$route.params.id,
      newSegment: clone(DEFAULT_SEGMENT),
      newVariant: clone(DEFAULT_VARIANT),
      selectedSegment: null,
      newDistributions: {},
      operatorOptions: operators,
      operatorValueToLabelMap: OPERATOR_VALUE_TO_LABEL_MAP
    }
  },
  computed: {
    newDistributionPercentageSum () {
      return sum(pluck(Object.values(this.newDistributions), 'percent'))
    },
    newDistributionIsValid () {
      const percentageSum = sum(pluck(Object.values(this.newDistributions), 'percent'))
      return percentageSum === 100
    }
  },
  methods: {
    deleteFlag () {
      const {flagId} = this.$route.params
      this.$http.delete(`${API_URL}/flags/${flagId}`)
        .then(() => {
          this.$router.replace({name: 'home'})
          this.$message.success(`You deleted flag ${flagId}`)
        }, err => {
          this.$message.error(err.body.message)
        })
    },
    putFlag (flag) {
      const flagId = this.$route.params.flagId
      this.$http.put(`${API_URL}/flags/${flagId}`, {
        description: flag.description,
        dataRecordsEnabled: flag.dataRecordsEnabled
      }).then(() => {
        this.$message.success(`You've updated flag`)
      }, err => {
        this.$message.error(err.body.message)
      })
    },
    setFlagEnabled (checked) {
      const flagId = this.$route.params.flagId
      this.$http.put(
        `${API_URL}/flags/${flagId}/enabled`,
        {enabled: checked}
      ).then(() => {
        const checkedStr = checked ? 'on' : 'off'
        this.$message.success(`You turned ${checkedStr} this feature flag`)
      }, err => {
        this.$message.error(err.body.message)
      })
    },
    selectVariant ($event, variant) {
      const checked = $event
      if (checked) {
        const distribution = Object.assign(clone(DEFAULT_DISTRIBUTION), {
          variantKey: variant.key,
          variantID: variant.id
        })
        this.$set(this.newDistributions, variant.id, distribution)
      } else {
        this.$delete(this.newDistributions, variant.id)
      }
    },
    editDistribution (segment) {
      this.selectedSegment = segment

      this.$set(this, 'newDistributions', {})

      segment.distributions.forEach(distribution => {
        this.$set(this.newDistributions, distribution.variantID, clone(distribution))
      })

      this.dialogEditDistributionOpen = true
    },
    saveDistribution (segment) {
      const flagId = this.$route.params.flagId

      const distributions = Object.values(this.newDistributions).filter(distribution => distribution.percent !== 0)

      this.$http.put(
        `${API_URL}/flags/${flagId}/segments/${segment.id}/distributions`,
        {distributions}
      ).then(response => {
        let distributions = response.body
        this.selectedSegment.distributions = distributions
        this.dialogEditDistributionOpen = false
        this.$message.success('distributions updated')
      }, err => {
        this.$message.error(err.body.message)
      })
    },
    createVariant () {
      const flagId = this.$route.params.flagId
      this.$http.post(
        `${API_URL}/flags/${flagId}/variants`,
        this.newVariant
      ).then(response => {
        let variant = response.body
        this.newVariant = clone(DEFAULT_VARIANT)
        this.flag.variants.push(variant)
        this.$message.success('new variant created')
      }, err => {
        this.$message.error(err.body.message)
      })
    },
    deleteVariant (variant) {
      const {flagId} = this.$route.params

      const isVariantInUse = this.flag.segments.some(segment => (
        segment.distributions.some(distribution => distribution.variantID === variant.id)
      ))

      if (isVariantInUse) {
        alert('This variant is being used by a segment distribution. Please remove the segment or edit the distribution in order to remove this variant.')
        return
      }

      if (!confirm(`Are you sure you want to delete variant #${variant.id} [${variant.key}]`)) {
        return
      }

      this.$http.delete(
        `${API_URL}/flags/${flagId}/variants/${variant.id}`
      ).then(() => {
        this.$message.success('variant deleted')
        this.fetchFlag()
      }, err => {
        this.$message.error(err.body.message)
      })
    },
    putVariant (variant) {
      const flagId = this.$route.params.flagId
      variant.attachment = JSON.parse(variant.attachmentStr)
      this.$http.put(
        `${API_URL}/flags/${flagId}/variants/${variant.id}`,
        variant
      ).then(() => {
        this.$message.success('variant updated')
      }, err => {
        this.$message.error(err.body.message)
      })
    },
    createConstraint (segment) {
      const {flagId} = this.$route.params
      this.$http.post(
        `${API_URL}/flags/${flagId}/segments/${segment.id}/constraints`,
        segment.newConstraint
      ).then(response => {
        let constraint = response.body
        segment.constraints.push(constraint)
        segment.newConstraint = clone(DEFAULT_CONSTRAINT)
        this.$message.success('new constraint created')
      }, err => {
        this.$message.error(err.body.message)
      })
    },
    putConstraint (segment, constraint) {
      const {flagId} = this.$route.params

      this.$http.put(
        `${API_URL}/flags/${flagId}/segments/${segment.id}/constraints/${constraint.id}`,
        constraint
      ).then(() => {
        this.$message.success('constraint updated')
      }, err => {
        this.$message.error(err.body.message)
      })
    },
    deleteConstraint (segment, constraint) {
      const {flagId} = this.$route.params

      if (!confirm('Are you sure you want to delete this constraint?')) {
        return
      }

      this.$http.delete(
        `${API_URL}/flags/${flagId}/segments/${segment.id}/constraints/${constraint.id}`
      ).then(() => {
        const index = segment.constraints.findIndex(constraint => constraint.id === constraint.id)
        segment.constraints.splice(index, 1)
        this.$message.success('constraint deleted')
      }, err => {
        this.$message.error(err.body.message)
      })
    },
    putSegment (segment) {
      const flagId = this.$route.params.flagId
      this.$http.put(
        `${API_URL}/flags/${flagId}/segments/${segment.id}`,
        {
          description: segment.description,
          rolloutPercent: parseInt(segment.rolloutPercent, 10)
        }
      ).then(() => {
        this.$message.success('segment updated')
      }, err => {
        this.$message.error(err.body.message)
      })
    },
    putSegmentsReorder (segments) {
      const flagId = this.$route.params.flagId
      this.$http.put(
        `${API_URL}/flags/${flagId}/segments/reorder`,
        { segmentIDs: pluck(segments, 'id') }
      ).then(() => {
        this.$message.success('segment reordered')
      }, err => {
        this.$message.error(err.body.message)
      })
    },
    deleteSegment (segment) {
      const {flagId} = this.$route.params

      if (!confirm('Are you sure you want to delete this segment?')) {
        return
      }

      this.$http.delete(
        `${API_URL}/flags/${flagId}/segments/${segment.id}`
      ).then(() => {
        const index = this.flag.segments.findIndex(el => el.id === segment.id)
        this.flag.segments.splice(index, 1)
        this.$message.success('segment deleted')
      }, err => {
        this.$message.error(err.body.message)
      })
    },
    createSegment () {
      const flagId = this.$route.params.flagId
      this.$http.post(
        `${API_URL}/flags/${flagId}/segments`,
        this.newSegment
      ).then(response => {
        let segment = response.body
        processSegment(segment)
        segment.constraints = []
        this.newSegment = clone(DEFAULT_SEGMENT)
        this.flag.segments.push(segment)
        this.$message.success('new segment created')
        this.dialogCreateSegmentOpen = false
      }, err => {
        this.$message.error(err.body.message)
      })
    },
    fetchFlag () {
      const flagId = this.$route.params.flagId
      this.$http.get(`${API_URL}/flags/${flagId}`).then(response => {
        let flag = response.body
        flag.segments.forEach(segment => processSegment(segment))
        flag.variants.forEach(variant => processVariant(variant))
        this.flag = flag
        this.loaded = true
      }, err => {
        this.$message.error(err.body.message)
      })
    }
  },
  mounted () {
    this.fetchFlag()
  }
}
</script>

<style lang="less" scoped>
h4 {
  padding: 0;
  margin: 10px 0;
}

.grabbable {
    cursor: move; /* fallback if grab cursor is unsupported */
    cursor: grab;
    cursor: -moz-grab;
    cursor: -webkit-grab;
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
}

.segment-description-rollout {
  margin-top: 10px;
}

.edit-distribution-button {
  margin-top: 5px;
}

.edit-distribution-choose-variants {
  padding: 10px 0;
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

.flag-content{
  margin-top: 8px;
}
</style>
