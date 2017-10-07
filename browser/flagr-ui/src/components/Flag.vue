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
          v-for="variant in flag.variants">
          <div>
            <el-checkbox
              @change="(e) => selectVariant(e, variant)"
              :checked="selectedVariants[variant.id]"
            >  
            </el-checkbox>
            <el-tag type="danger">
              {{ variant.key }}
            </el-tag>
          </div>
          <el-slider
            :disabled="!selectedVariants[variant.id]"
            v-model="distributionPercentages[variant.id]"
            show-input>
          </el-slider>
        </li>
      </ul>
      <el-button
        class="width--full"
        @click.prevent="() => saveDistribution(selectedSegment)">
        Save
      </el-button>
    </el-dialog>

    <el-breadcrumb separator="/">
      <el-breadcrumb-item :to="{name: 'home'}">Home page</el-breadcrumb-item>
      <el-breadcrumb-item>Flag {{ $route.params.flagId }}</el-breadcrumb-item>
    </el-breadcrumb>

    <div v-if="loaded && flag">
      <div class="flex-row">
        <div class="flex-row-left">
          <h2>Flag #{{ $route.params.flagId }}</h2>
        </div>
        <div class="flex-row-right">
          <el-button @click="dialogDeleteFlagVisible = true">
            <span class="el-icon-delete2"></span>
            Delete
          </el-button>
        </div>
      </div>
      <div class="flag-description">
        {{ flag.description }}
      </div>
      <div class="segments-container">
        <h2>Segments ({{ flag.segments.length }})</h2>
        <ul class="segments-container-inner" v-if="flag.segments.length">
          <li
            v-for="segment in flag.segments"
            class="segment">
            <div
              class="flex-row highlightable"
              @click.prevent="() => expandSegment(segment)">
              <div class="flex-row-left">
                <span
                  v-bind:class="{'el-icon-caret-right': !segment._expanded, 'el-icon-caret-bottom': segment._expanded}">
                </span>
                <el-badge :value="segment.constraints.length" :hidden="!segment.constraints.length">
                  <el-tag>{{ segment.id }}</el-tag>
                </el-badge>
                {{ segment.description }}
              </div>
              <div class="flex-row-right">
                {{ segment.rolloutPercent || 0 }}%
              </div>
            </div>
            <div class="flex-row equal-width align-items-top" v-if="segment._expanded">
              <div class="segment-contraints">
                <h4>Constraints ({{segment.constraints.length}})</h4>
                <div class="constraints">
                  <ol class="constraints-inner" v-if="segment.constraints.length">
                    <li v-for="constraint in segment.constraints">
                      <el-tag type="gray">{{ constraint.property }}</el-tag>
                      <el-tag type="primary">{{ operatorValueToLabelMap[constraint.operator] }}</el-tag>
                      <el-tag type="gray">{{ constraint.value }}</el-tag>
                    </li>
                  </ol>
                  <div class="card--empty" v-else>
                    No constraints for this segment yet
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
                  <li v-for="distribution in segment.distributions">
                    <el-tag type="danger">{{ distribution.variantKey }}</el-tag>
                    <span>{{ distribution.percent }} %</span>
                  </li>
                </ul>
                <div class="card--empty" v-else>
                  No distribution yet
                </div>
                <div>
                  <el-button class="width--full" @click="selectedSegment = segment; dialogEditDistributionOpen = true">
                    <span class="el-icon-edit"></span>
                    Edit distribution
                  </el-button>
                </div>
              </div>
            </div>
          </li>
        </ul>
        <div class="card--empty" v-else>
          No segments created for this feature flag yet
        </div>
        <div class="variants-container">
          <h2>Variants ({{ flag.variants.length }})</h2>
          <div class="variants-container-inner" v-if="flag.variants.length">
            <el-tag type="danger" v-for="variant in flag.variants" :key="variant.id">
              {{ variant.key }}
            </el-tag>
          </div>
          <div class="card--empty" v-else>
            No variants created for this feature flag yet
          </div>
          <div class="variants-input">
            <div class="flex-row equal-width constraints-inputs-container">
              <div>
                <el-input
                  placeholder="Key"
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
        </div>
        <hr/>
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
      </div>
    </div>
    <spinner v-if="!loaded"></spinner>
  </div>
</template>

<script>
import constants from '@/constants'
import fetchHelpers from '@/helpers/fetch'
import Spinner from '@/components/Spinner'
import clone from 'lodash.clone'
import { Button, Dialog, Slider, Checkbox, Tag, Breadcrumb, BreadcrumbItem } from 'element-ui'
import {operators} from '@/../config/operators.json'

const OPERATOR_VALUE_TO_LABEL_MAP = operators.reduce((acc, el) => {
  acc[el.value] = el.label
  return acc
}, {})

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
  segment._expanded = false
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
    'el-breadcrumb-item': BreadcrumbItem
  },
  data () {
    return {
      loaded: false,
      dialogDeleteFlagVisible: false,
      dialogEditDistributionOpen: false,
      flag: null,
      newSegment: clone(DEFAULT_SEGMENT),
      newVariant: clone(DEFAULT_VARIANT),
      selectedSegment: null,
      newDistributions: [],
      distributionPercentages: {},
      operatorOptions: operators,
      operatorValueToLabelMap: OPERATOR_VALUE_TO_LABEL_MAP
    }
  },
  computed: {
    selectedVariants () {
      return this.newDistributions.reduce((acc, el) => {
        acc[el.variantID] = true
        return acc
      }, {})
    }
  },
  methods: {
    expandSegment (segment) {
      segment._expanded = !segment._expanded
    },
    deleteFlag () {
      const flagId = this.$route.params.flagId
      fetch(`${API_URL}/flags/${flagId}`, {method: 'delete'})
        .then(() => {
          this.$router.replace({name: 'home'})
          this.$message(`You deleted flag ${flagId}`)
        })
    },
    selectVariant ($event, variant) {
      const checked = $event.target.checked

      if (checked) {
        const distribution = Object.assign(clone(DEFAULT_DISTRIBUTION), {
          variantKey: variant.key,
          variantID: variant.id
        })

        this.newDistributions.push(distribution)
      } else {
        this.newDistributions = this.newDistributions.filter(distribution => distribution.variantID !== variant.id)
      }
    },
    saveDistribution (segment) {
      const flagId = this.$route.params.flagId

      const distributions = this.newDistributions.map(distribution => {
        const percent = this.distributionPercentages[distribution.variantID]
        if (percent) {
          distribution.percent = percent
        }
        return distribution
      })

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
      const flagId = this.$route.params.flagId
      postJson(`${API_URL}/flags/${flagId}/segments/${segment.id}/constraints`, segment.newConstraint)
        .then(constraint => {
          segment.constraints.push(constraint)
          segment.newConstraint = clone(DEFAULT_CONSTRAINT)
          this.$message('You created a new constraint')
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

<style lang="scss" scoped>
h4 {
  padding: 0;
  margin: 10px 0;
}

.flag-container {
  width: 800px;
}

.flag-description {
  font-size: 1.2em;
  padding: 10px 20px;
  background-color: white;
  border-radius: 3px;
  border: 1px solid #ddd;
}

ul.segments-container-inner {
  li {
    padding: 5px 0;
  }
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

hr {
  border-color: #eee;
  border-width: 1px;
  background-color: #eee;
  margin: 30px 0;
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
  .el-tag {
    margin-right: 5px;
  }
}

.segment-distributions {
  padding-left: 8px;
}

.edit-distribution-choose-variants {
  padding: 10px 0;
  li {

  }
}
</style>
