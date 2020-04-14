<template>
  <div>
    <el-card
      v-for="diff in diffs"
      :key="diff.timestamp"
      class="snapshot-container"
    >
      <div slot="header" class="el-card-header">
        <el-row>
          <el-col :span="14">
            <div class="diff-snapshot-id-change">
              <el-tag :disable-transitions="true"
                >Snapshot ID: {{ diff.oldId }}</el-tag
              >
              <span class="el-icon-d-arrow-right" />
              <el-tag :disable-transitions="true"
                >Snapshot ID: {{ diff.newId }}</el-tag
              >
            </div>
          </el-col>
          <el-col :span="10" style="text-align: right; color: #2e4960">
            <div v-bind:class="{ compact: diff.updatedBy }">
              <span size="small">{{ diff.timestamp }}</span>
            </div>
            <div class="compact" v-if="diff.updatedBy">
              <span size="small">UPDATED BY: {{ diff.updatedBy }}</span>
            </div>
          </el-col>
        </el-row>
      </div>
      <pre class="diff" v-html="diff.flagDiff"></pre>
    </el-card>
  </div>
</template>

<script>
import Axios from "axios";
import { diffJson, convertChangesToXML } from "diff";

import constants from "@/constants";

const { API_URL } = constants;

export default {
  name: "flag-history",
  props: ["flagId"],
  data() {
    return {
      flagSnapshots: [],
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
          flagDiff: this.getDiff(snapshots[i].flag, snapshots[i + 1].flag),
        });
      }
      return ret;
    },
  },
  methods: {
    getFlagSnapshots() {
      Axios.get(`${API_URL}/flags/${this.$props.flagId}/snapshots`).then(
        (response) => {
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
    },
  },
  mounted() {
    this.getFlagSnapshots();
  },
};
</script>

<style lang="less">
.snapshot-container {
  .compact {
    margin: -8px 0px;
  }
  .diff-snapshot-id-change {
    color: white;
    .el-tag {
      color: #2e4960;
      background-color: white;
    }
  }
  .diff {
    margin: 0;
    del {
      background-color: #f7b3b3;
      text-decoration: none;
    }
    ins {
      background-color: #b6ddc6;
      text-decoration: none;
    }
    overflow-x: auto;
  }
}
</style>
