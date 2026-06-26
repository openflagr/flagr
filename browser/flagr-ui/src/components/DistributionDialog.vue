<template>
  <el-dialog title="Edit Distribution" :model-value="visible" @update:model-value="$emit('update:visible', $event)">
    <div v-if="flag" class="dist-dialog-body">
      <div v-for="variant in flag.variants" :key="'distribution-variant-' + variant.id" class="dist-variant-row">
        <div class="dist-variant-header">
          <el-checkbox
            @change="(e: boolean) => selectVariant(e, variant)"
            :checked="!!draft[String(variant.id)]"
          />
          <span class="dist-variant-key">{{ variant.key }}</span>
        </div>
        <div v-if="!!draft[String(variant.id)]">
          <div class="dist-slider-row">
            <el-slider v-model="draft[String(variant.id)].percent" show-input :max="100" :step="1" />
          </div>
        </div>
        <div v-else class="dist-disabled-hint">
          <el-progress :percentage="0" :stroke-width="4" :show-text="false" color="var(--el-border-color-light)" />
        </div>
      </div>

      <el-alert
        class="edit-distribution-alert"
        v-if="!isValid"
        :title="'Percentages must add up to 100% (currently ' + percentageSum + '%)'"
        type="error"
        show-icon
      />
    </div>

    <template #footer>
      <el-button @click="$emit('update:visible', false)">Cancel</el-button>
      <el-button type="primary" :disabled="!isValid" @click.prevent="$emit('save', draft)">Save</el-button>
    </template>
  </el-dialog>
</template>

<script lang="ts">
import helpers from '@/helpers/helpers'
import type { PropType } from 'vue'
import type { DistributionDraft, FlagView, Variant } from '@/api/types'

export default {
  name: 'distribution-dialog',
  props: {
    visible: Boolean,
    flag: { type: Object as PropType<FlagView | null>, default: null },
    initialDistributions: {
      type: Object as PropType<Record<string, DistributionDraft>>,
      default: () => ({}),
    },
  },
  emits: ['update:visible', 'save'],
  data() {
    return {
      draft: {} as Record<string, DistributionDraft>,
    }
  },
  computed: {
    percentageSum(): number {
      return helpers.sum(helpers.pluck(Object.values(this.draft), 'percent'))
    },
    isValid(): boolean {
      return this.percentageSum === 100
    },
  },
  methods: {
    selectVariant($event: boolean, variant: Variant) {
      const key = String(variant.id)
      if ($event) {
        this.draft[key] = {
          variantKey: variant.key,
          variantID: variant.id!,
          percent: 0,
          bitmap: '',
        }
      } else {
        delete this.draft[key]
      }
    },
  },
  watch: {
    visible: {
      immediate: true,
      handler(open: boolean) {
        if (open) {
          this.draft = {}
          for (const [id, dist] of Object.entries(this.initialDistributions)) {
            this.draft[id] = { ...dist }
          }
        }
      },
    },
  },
}
</script>

<style lang="scss" scoped>
.dist-dialog-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.dist-variant-row {
  background: var(--el-fill-color-light);
  border-radius: 8px;
  padding: 12px;
}
.dist-variant-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}
.dist-variant-key {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}
.dist-slider-row {
  padding-left: 28px;
}
.dist-disabled-hint {
  padding-left: 28px;
}
.edit-distribution-alert {
  margin-top: 0;
}
</style>