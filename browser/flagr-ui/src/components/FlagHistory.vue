<template>
  <div>
    <el-card
      v-for="diff in diffs"
      :id="snapshotElementId(diff.newId)"
      :key="diff.newId"
      class="snapshot-container"
      :data-testid="`snapshot-${diff.newId}`"
    >
      <template #header>
        <div class="el-card-header">
          <div class="snapshot-header">
            <div class="snapshot-header-left">
              <div class="diff-snapshot-id-change">
                <el-tag disable-transitions>
                  Snapshot ID: {{ diff.oldId }}
                </el-tag>
                <el-icon><DArrowRight /></el-icon>
                <el-tag disable-transitions>
                  Snapshot ID: {{ diff.newId }}
                </el-tag>
                <copy-link-button
                  :url="snapshotShareUrl(diff.newId)"
                  aria-label="Copy link to this change"
                  tooltip="Copy link to this change"
                  :test-id="`copy-snapshot-url-btn-${diff.newId}`"
                />
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
      <!-- eslint-disable vue/no-v-html -- sanitized via xss() in getDiff -->
      <pre
        class="diff"
        v-html="diff.flagDiff"
      />
      <!-- eslint-enable vue/no-v-html -->
    </el-card>
    <div
      v-if="hasMore"
      class="history-load-older"
    >
      <el-button
        class="load-older-btn"
        :loading="loadingOlder"
        data-testid="load-older-snapshots-btn"
        @click="$emit('load-older')"
      >
        Load older changes
        <el-icon class="el-icon--right">
          <ArrowDown />
        </el-icon>
      </el-button>
    </div>
  </div>
</template>

<script lang="ts">
import xss from 'xss'
import { diffJson, convertChangesToXML } from 'diff'
import { ArrowDown, DArrowRight } from '@element-plus/icons-vue'
import CopyLinkButton from '@/components/CopyLinkButton.vue'
import { flagSnapshotUrl, snapshotElementId } from '@/helpers/shareLinks'
import type { Flag, FlagHistoryDiffRow, FlagSnapshot } from '@/api/types'

export default {
  name: 'FlagHistory',
  components: { ArrowDown, DArrowRight, CopyLinkButton },
  props: {
    snapshots: {
      type: Array as () => FlagSnapshot[],
      required: true,
    },
    flagId: {
      type: [String, Number],
      required: true,
    },
    /**
     * True when snapshots hold only the newest pages of a longer history.
     * Skips the empty-flag sentinel (the oldest loaded snapshot is withheld
     * as the diff base instead of being rendered as a bogus creation diff)
     * and shows the "load older" control.
     */
    hasMore: {
      type: Boolean,
      default: false,
    },
    loadingOlder: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['load-older'],
  computed: {
    diffs() {
      const ret: FlagHistoryDiffRow[] = []
      const snapshots = this.snapshots.slice()
      if (!this.hasMore) {
        snapshots.push({ flag: {} as Flag, id: 0 })
      }
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
    snapshotElementId,
    snapshotShareUrl(snapshotId: number) {
      return flagSnapshotUrl(this.flagId, snapshotId, window.location)
    },
    getDiff(newFlag: Flag, oldFlag: Flag) {
      const o = JSON.parse(JSON.stringify(oldFlag))
      const n = JSON.parse(JSON.stringify(newFlag))
      const d = diffJson(o, n)
      if (d.length === 1) {
        return 'No changes'
      }
      return xss(convertChangesToXML(d))
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
.history-load-older {
  text-align: center;
  // The card above already carries a bottom margin; pull the button up so the
  // gap doesn't read as a dead band, and give it room below.
  margin-top: -10px;
  padding-bottom: var(--space-sm);
}
.load-older-btn {
  min-width: 240px;
}
.diff-snapshot-id-change {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-3xs);
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
