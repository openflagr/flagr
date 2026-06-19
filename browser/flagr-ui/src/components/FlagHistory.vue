<template>
  <div>
    <el-card v-for="diff in diffs" :key="diff.timestamp" class="snapshot-container">
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
            <div class="compact" v-if="diff.updatedBy">
              <span>UPDATED BY: {{ diff.updatedBy }}</span>
            </div>
          </div>
        </div>
      </div>
      </template>
      <pre class="diff" v-html="diff.flagDiff"></pre>
    </el-card>
  </div>
</template>

<script>
import Axios from "axios";
import { diffJson, convertChangesToXML } from "diff";
import { DArrowRight } from "@element-plus/icons-vue";

import constants from "@/constants";

const { API_URL } = constants;

export default {
  name: "flag-history",
  components: { DArrowRight },
  props: ["flagId"],
  data() {
    return {
      flagSnapshots: []
    };
  },
  computed: {
    diffs() {
      let ret = [];
      let snapshots = this.flagSnapshots.slice();
      snapshots.push({ flag: {} });
      for (let i = 0; i < snapshots.length - 1; i++) {
        ret.push({
          timestamp: new Date(snapshots[i].updatedAt).toLocaleString(),
          updatedBy: snapshots[i].updatedBy,
          newId: snapshots[i].id,
          oldId: snapshots[i + 1].id || "NULL",
          flagDiff: this.getDiff(snapshots[i].flag, snapshots[i + 1].flag)
        });
      }
      return ret;
    }
  },
  methods: {
    getFlagSnapshots() {
      Axios.get(`${API_URL}/flags/${this.$props.flagId}/snapshots`).then(
        response => {
          this.flagSnapshots = response.data;
        },
        () => {
          this.$message.error(`failed to get flag snapshots`);
        }
      );
    },
    getDiff(newFlag, oldFlag) {
      const o = JSON.parse(JSON.stringify(oldFlag));
      const n = JSON.parse(JSON.stringify(newFlag));
      const d = diffJson(o, n);
      if (d.length === 1) {
        return "No changes";
      }
      return convertChangesToXML(d);
    }
  },
  mounted() {
    this.getFlagSnapshots();
  }
};
</script>

<style lang="less" scoped>
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
