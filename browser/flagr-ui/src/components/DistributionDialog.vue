<template>
  <el-dialog title="Edit distribution" :model-value="visible" @update:model-value="$emit('update:visible', $event)">
    <div v-if="flag">
      <div v-for="variant in flag.variants" :key="'distribution-variant-' + variant.id">
        <div>
          <el-checkbox
            @change="(e) => selectVariant(e, variant)"
            :checked="!!draft[variant.id]"
          ></el-checkbox>
          <el-tag type="danger">{{ variant.key }}</el-tag>
        </div>
        <el-slider
          v-if="!draft[variant.id]"
          :value="0"
          :disabled="true"
          show-input
        ></el-slider>
        <div v-if="!!draft[variant.id]">
          <el-slider
            v-model="draft[variant.id].percent"
            :disabled="false"
            show-input
          ></el-slider>
        </div>
      </div>
    </div>
    <el-button
      class="width--full"
      :disabled="!isValid"
      @click.prevent="$emit('save', draft)"
    >Save</el-button>

    <el-alert
      class="edit-distribution-alert"
      v-if="!isValid"
      :title="'Percentages must add up to 100% (currently at ' + percentageSum + '%)'"
      type="error"
      show-icon
    ></el-alert>
  </el-dialog>
</template>

<script>
import helpers from "@/helpers/helpers"

export default {
  name: "distribution-dialog",
  props: {
    visible: Boolean,
    flag: Object,
    initialDistributions: {
      type: Object,
      default: () => ({})
    }
  },
  emits: ["update:visible", "save"],
  data() {
    return {
      draft: {}
    }
  },
  computed: {
    percentageSum() {
      return helpers.sum(helpers.pluck(Object.values(this.draft), "percent"))
    },
    isValid() {
      return this.percentageSum === 100
    }
  },
  methods: {
    selectVariant($event, variant) {
      if ($event) {
        this.draft[variant.id] = { variantKey: variant.key, variantID: variant.id, percent: 0, bitmap: "" }
      } else {
        delete this.draft[variant.id]
      }
    }
  },
  watch: {
    visible: {
      immediate: true,
      handler(open) {
        if (open) {
          // Clone initial distributions into draft
          this.draft = {}
          for (const [id, dist] of Object.entries(this.initialDistributions)) {
            this.draft[id] = { ...dist }
          }
        }
      }
    }
  }
}
</script>

<style lang="less" scoped>
.edit-distribution-alert {
  margin-top: 10px;
}
</style>
