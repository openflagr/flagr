<template>
  <div class="container flag-container">
    <el-dialog
      title="Delete feature flag"
      :visible.sync="dialogDeleteFlagVisible"
      size="tiny">
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
            <el-tag type="danger">
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
      <el-breadcrumb-item>Flag {{ $route.params.flagId }}</el-breadcrumb-item>
    </el-breadcrumb>

    <div v-if="loaded && flag">
      <el-card>
        <div slot="header" class="el-card-header">
          <div class="flex-row">
            <div class="flex-row-left">
              <h2>Flag ID: {{ $route.params.flagId }}</h2>
            </div>
            <div class="flex-row-right" v-if="flag">
              <el-switch
                v-model="flag.enabled"
                on-color="#13ce66"
                off-color="#ff4949"
                @change="setFlagEnabled"
                :on-value="true"
                :off-value="false">
              </el-switch>
            </div>
          </div>
        </div>
        <div class="flag-description">
          <div>
            <el-input
              placeholder="Key"
              v-model="flag.description">
              <template slot="prepend">Flag Description</template>
            </el-input>
          </div>
        </div>
      </el-card>

      <el-card class="variants-container">
        <div slot="header" class="clearfix">
          <h2>Variants</h2>
        </div>
        <div class="variants-container-inner" v-if="flag.variants.length">
          <div v-for="variant in flag.variants" :key="variant.id">
            <el-card>
              <el-form :label-position="right" label-width="100px">
                <el-form-item label="Key">
                  <el-input
                    placeholder="Key"
                    v-model="variant.key">
                  </el-input>
                </el-form-item>
                <el-form-item label="Attachment">
                  <el-input
                    placeholder="{}"
                    v-model="variant.attachment">
                  </el-input>
                </el-form-item>
              </el-form>
            </el-card>
          </div>
        </div>
        <div class="card--empty" v-else>
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
              <el-button @click="dialogCreateSegmentOpen = true">
                <span class="el-icon-edit"></span>
                Create
              </el-button>
            </div>
          </div>
        </div>
        <div class="segments-container-inner" v-if="flag.segments.length">
          <el-card
            v-for="segment in flag.segments"
            :key="segment.id"
            class="segment">
            <div class="flex-row">
              <div class="flex-row-left">
                <el-tag>{{ segment.id }}</el-tag> {{ segment.description }}
              </div>
              <div class="flex-row-right">
                {{ segment.rolloutPercent || 0 }}%
              </div>
            </div>
            <hr>
            <div class="flex-row equal-width align-items-top">
              <div class="segment-contraints">
                <h4>Constraints</h4>
                <div class="constraints">
                  <ol class="constraints-inner" v-if="segment.constraints.length">
                    <li
                      class="flex-row"
                      v-for="constraint in segment.constraints"
                      :key="constraint.id">
                      <div class="flex-row-left">
                        <el-tag type="gray">{{ constraint.property }}</el-tag>
                        <el-tag type="primary">{{ operatorValueToLabelMap[constraint.operator] }}</el-tag>
                        <el-tag type="gray">{{ constraint.value }}</el-tag>
                      </div>
                      <div class="flex-row-right">
                        <el-button @click="deleteConstraint(segment, constraint)">
                          <span class="el-icon-delete2"/>
                        </el-button>
                      </div>
                    </li>
                  </ol>
                  <div class="card--empty" v-else>
                    <span>No constraints for this segment yet</span>
                  </div>
                  <div>
                    <div class="flex-row equal-width constraints-inputs-container">
                      <div>
                        <el-input
                          placeholder="Property"
                          v-model="segment.newConstraint.property">  
                        </el-input>
                      </div>
                      <div>
                        <el-select v-model="segment.newConstraint.operator" placeholder="operator">
                          <el-option
                            v-for="item in operatorOptions"
                            :key="item.value"
                            :label="item.label"
                            :value="item.value">
                          </el-option>
                        </el-select>
                      </div>
                      <div>
                        <el-input
                          placeholder="Value"
                          v-model="segment.newConstraint.value">  
                        </el-input>
                      </div>
                    </div>
                    <el-button
                      class="width--full"
                      :disabled="!segment.newConstraint.property || !segment.newConstraint.value"
                      @click.prevent="() => createConstraint(segment)">
                      Create Constraint
                    </el-button>
                  </div>
                </div>
              </div>
              <div class="segment-distributions">
                <h4>Distribution</h4>
                <ul class="segment-distributions-inner" v-if="segment.distributions.length">
                  <li v-for="distribution in segment.distributions" :key="distribution.id">
                    <el-tag type="danger">{{ distribution.variantKey }}</el-tag>
                    <span>{{ distribution.percent }} %</span>
                  </li>
                </ul>
                <div class="card--empty" v-else>
                  No distribution yet
                </div>
                <div>
                  <el-button class="width--full" @click="editDistribution(segment)">
                    <span class="el-icon-edit"></span>
                    Edit distribution
                  </el-button>
                </div>
              </div>
            </div>
          </el-card>
        </div>
        <div class="card--empty" v-else>
          No segments created for this feature flag yet
        </div>
      </el-card>
      <el-card>
        <div slot="header" class="el-card-header">
          <h2>Flag Settings</h2>
        </div>
        <el-button @click="dialogDeleteFlagVisible = true" type="danger">
          <span class="el-icon-delete2"></span>
          Delete Flag
        </el-button>
      </el-card>
    </div>
    <spinner v-if="!loaded"></spinner>
  </div>
</template>

<script>
import constants from '@/constants'
import helpers from '@/helpers/helpers'
import fetchHelpers from '@/helpers/fetch'
import Spinner from '@/components/Spinner'
import clone from 'lodash.clone'
import { Button, Dialog, Slider, Checkbox, Tag, Breadcrumb, BreadcrumbItem, Switch } from 'element-ui'
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
  getJson,
  postJson,
  putJson
} = fetchHelpers

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

export default {
  name: 'flag',
  components: {
    spinner: Spinner,
    'el-button': Button,
    'el-dialog': Dialog,
    'el-slider': Slider,
    'el-checkbox': Checkbox,
    'el-tag': Tag,
    'el-breadcrumb': Breadcrumb,
    'el-breadcrumb-item': BreadcrumbItem,
    'el-switch': Switch
  },
  data () {
    return {
      loaded: false,
      dialogDeleteFlagVisible: false,
      dialogEditDistributionOpen: false,
      dialogCreateSegmentOpen: false,
      flag: null,
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
      fetch(`${API_URL}/flags/${flagId}`, {method: 'delete'})
        .then(() => {
          this.$router.replace({name: 'home'})
          this.$message(`You deleted flag ${flagId}`)
        })
    },
    setFlagEnabled (checked) {
      const flagId = this.$route.params.flagId
      putJson(`${API_URL}/flags/${flagId}/enabled`, {enabled: checked})
        .then(() => {
          const checkedStr = checked ? 'on' : 'off'
          this.$message(`You turned ${checkedStr} this feature flag`)
        })
    },
    selectVariant ($event, variant) {
      const checked = $event.target.checked

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

      putJson(`${API_URL}/flags/${flagId}/segments/${segment.id}/distributions`, {distributions})
        .then(distributions => {
          this.selectedSegment.distributions = distributions
          this.dialogEditDistributionOpen = false
        })
    },
    createVariant () {
      const flagId = this.$route.params.flagId
      postJson(`${API_URL}/flags/${flagId}/variants`, this.newVariant)
        .then(variant => {
          this.$message('You created a new variant')
          this.newVariant = clone(DEFAULT_VARIANT)
          this.flag.variants.push(variant)
        })
    },
    createConstraint (segment) {
      const {flagId} = this.$route.params
      postJson(`${API_URL}/flags/${flagId}/segments/${segment.id}/constraints`, segment.newConstraint)
        .then(constraint => {
          segment.constraints.push(constraint)
          segment.newConstraint = clone(DEFAULT_CONSTRAINT)
          this.$message('You created a new constraint')
        })
    },
    deleteConstraint (segment, constraint) {
      const {flagId} = this.$route.params

      if (!confirm('Are you sure you want to delete this constraint?')) {
        return
      }

      fetch(
        `${API_URL}/flags/${flagId}/segments/${segment.id}/constraints/${constraint.id}`,
        {method: 'delete'}
      ).then(() => {
        const index = segment.constraints.findIndex(constraint => constraint.id === constraint.id)
        segment.constraints.splice(index, 1)
      })
    },
    createSegment () {
      const flagId = this.$route.params.flagId
      postJson(`${API_URL}/flags/${flagId}/segments`, this.newSegment)
        .then(segment => {
          processSegment(segment)
          segment.constraints = []
          this.newSegment = clone(DEFAULT_SEGMENT)
          this.flag.segments.push(segment)
          this.$message('You created a new segment')
          this.dialogCreateSegmentOpen = false
        })
    }
  },
  created () {
    const flagId = this.$route.params.flagId

    getJson(`${API_URL}/flags/${flagId}`).then(flag => {
      flag.segments.forEach(segment => processSegment(segment))
      this.flag = flag
      this.loaded = true
    })
  }
}
</script>

<style lang="less" scoped>
h4 {
  padding: 0;
  margin: 10px 0;
}

h2 {
  margin: -0.2em;
  color: white;
}

.flag-container {
  width: 700px;
}

.el-breadcrumb {
  margin-bottom: 2em;
}

.segment {
  cursor: pointer;
  .highlightable {
    padding: 4px;
    &:hover {
      background-color: #ddd;
    }
  }
}

ol.constraints-inner {
  background-color: white;
  padding: 5px;
  padding-left: 30px;
  border-radius: 3px;
  border: 1px solid #ddd;
  li {
    padding: 8px 0;
    .el-tag {
      font-size: 1em;
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

.segment-distributions {
  padding-left: 8px;
}

.edit-distribution-choose-variants {
  padding: 10px 0;
}

.edit-distribution-alert {
  margin-top: 10px;
} 

.el-card {
  margin-bottom: 1em;
}
</style>
