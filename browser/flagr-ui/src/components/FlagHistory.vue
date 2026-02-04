<template>
  <div>
    <el-card
      v-for="diff in diffs"
      :key="diff.timestamp"
      class="snapshot-container"
    >
      <template #header>
        <div class="el-card-header">
          <el-row>
            <el-col :span="14">
              <div class="diff-snapshot-id-change">
                <el-tag :disable-transitions="true">
                  Snapshot ID: {{ diff.oldId }}
                </el-tag>
                <el-icon><DArrowRight /></el-icon>
                <el-tag :disable-transitions="true">
                  Snapshot ID: {{ diff.newId }}
                </el-tag>
              </div>
            </el-col>
            <el-col
              :span="10"
              style="text-align: right; color: #2e4960"
            >
              <div :class="{ compact: diff.updatedBy }">
                <span size="small">{{ diff.timestamp }}</span>
              </div>
              <div
                v-if="diff.updatedBy"
                class="compact"
              >
                <span size="small">UPDATED BY: {{ diff.updatedBy }}</span>
              </div>
            </el-col>
          </el-row>
        </div>
      </template>
      <pre
        class="diff"
        v-html="diff.flagDiff"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import Axios from "axios";
import { diffJson, convertChangesToXML } from "diff";
import { ElMessage } from "element-plus";

import constants from "@/constants";

const props = defineProps({
  flagId: {
    type: Number,
    required: true,
  },
});

const { API_URL } = constants;

const flagSnapshots = ref([]);

const diffs = computed(() => {
  let ret = [];
  let snapshots = flagSnapshots.value.slice();
  snapshots.push({ flag: {} });
  for (let i = 0; i < snapshots.length - 1; i++) {
    ret.push({
      timestamp: new Date(snapshots[i].updatedAt).toLocaleString(),
      updatedBy: snapshots[i].updatedBy,
      newId: snapshots[i].id,
      oldId: snapshots[i + 1].id || "NULL",
      flagDiff: getDiff(snapshots[i].flag, snapshots[i + 1].flag)
    });
  }
  return ret;
});

function getFlagSnapshots() {
  Axios.get(`${API_URL}/flags/${props.flagId}/snapshots`).then(
    response => {
      flagSnapshots.value = response.data;
    },
    () => {
      ElMessage.error(`failed to get flag snapshots`);
    }
  );
}

function getDiff(newFlag, oldFlag) {
  const o = JSON.parse(JSON.stringify(oldFlag));
  const n = JSON.parse(JSON.stringify(newFlag));
  const d = diffJson(o, n);
  if (d.length === 1) {
    return "No changes";
  }
  return convertChangesToXML(d);
}

onMounted(() => {
  getFlagSnapshots();
});
</script>

<style lang="less">
.snapshot-container {
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
