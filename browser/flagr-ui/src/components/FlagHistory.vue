<template>
  <div>
    <el-card
      v-for="diff in diffs"
      :key="diff.timestamp"
      class="snapshot-container"
    >
      <template #header>
        <div class="el-card-header">
          <div class="snapshot-header">
            <div class="snapshot-header-left">
              <div class="diff-snapshot-id-change">
                <el-tag>Snapshot ID: {{ diff.oldId }}</el-tag>
                <el-icon><DArrowRight /></el-icon>
                <el-tag>Snapshot ID: {{ diff.newId }}</el-tag>
              </div>
            </div>
            <div class="snapshot-header-right">
              <div :class="{ compact: diff.updatedBy }">
                <span>{{ diff.timestamp }}</span>
              </div>
              <div
                v-if="diff.updatedBy"
                class="compact"
              >
                <span>UPDATED BY: {{ diff.updatedBy }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>
      <pre
        class="diff"
        v-html="diff.flagDiff"
      />
    </el-card>
  </div>
</template>

<script lang="ts">
import { diffJson, convertChangesToXML } from 'diff'
import { DArrowRight } from '@element-plus/icons-vue'
import type { Flag, FlagSnapshot } from '@/api/types'

export default {
  name: 'FlagHistory',
  components: { DArrowRight },
  props: {
    snapshots: {
      type: Array as () => FlagSnapshot[],
      required: true,
    },
  },
  computed: {
    diffs() {
      const ret: Array<Record<string, unknown>> = []
      const snapshots = this.snapshots.slice()
      snapshots.push({ flag: {} as Flag, id: 0 })
      for (let i = 0; i < snapshots.length - 1; i++) {
        ret.push({
          timestamp: new Date(snapshots[i].updatedAt ?? '').toLocaleString(),
          updatedBy: snapshots[i].updatedBy,
          newId: snapshots[i].id,
          oldId: snapshots[i + 1].id || 'NULL',
          flagDiff: this.getDiff(snapshots[i].flag, snapshots[i + 1].flag),
        })
      }
      return ret
    },
  },
  methods: {
    getDiff(newFlag: Flag, oldFlag: Flag) {
      const o = JSON.parse(JSON.stringify(oldFlag))
      const n = JSON.parse(JSON.stringify(newFlag))
      const d = diffJson(o, n)
      if (d.length === 1) {
        return 'No changes'
      }
      return convertChangesToXML(d)
    },
  },
}
</script>

<style lang="scss" scoped>
.snapshot-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2xs);
  flex-wrap: wrap;
}
.snapshot-header-right {
  text-align: right;
  color: var(--el-text-color-secondary);
}
@media (max-width: 640px) {
  .snapshot-header {
    flex-direction: column;
    align-items: flex-start;
  }
  .snapshot-header-right {
    text-align: left;
  }
}
</style>
