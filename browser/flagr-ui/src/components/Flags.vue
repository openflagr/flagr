<template>
  <el-row>
    <el-col
      :span="20"
      :offset="2"
    >
      <div class="flags-container container">
        <el-breadcrumb
          v-if="loaded"
          separator="/"
        >
          <el-breadcrumb-item>Home page</el-breadcrumb-item>
        </el-breadcrumb>

        <spinner v-if="!loaded" />

        <div v-if="loaded">
          <el-row>
            <el-col>
              <el-input
                v-model="newFlag.description"
                placeholder="Specific new flag description"
              >
                <template #prepend>
                  <el-icon><Plus /></el-icon>
                </template>
                <template #append>
                  <el-dropdown
                    split-button
                    type="primary"
                    :disabled="!newFlag.description"
                    @command="onCommandDropdown"
                    @click="createFlag"
                  >
                    Create New Flag
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item
                          command="simple_boolean_flag"
                          :disabled="!newFlag.description"
                        >
                          Create Simple Boolean Flag
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </template>
              </el-input>
            </el-col>
          </el-row>

          <el-row>
            <el-input
              v-model="searchTerm"
              v-focus
              placeholder="Search a flag"
              :prefix-icon="Search"
            />
          </el-row>

          <el-table
            :data="filteredFlags"
            :stripe="true"
            :highlight-current-row="false"
            :default-sort="{ prop: 'id', order: 'descending' }"
            style="width: 100%"
            @row-click="goToFlag"
          >
            <el-table-column
              prop="id"
              align="center"
              label="Flag ID"
              sortable
              fixed
              width="95"
            />
            <el-table-column
              prop="description"
              label="Description"
              min-width="300"
            />
            <el-table-column
              prop="tags"
              label="Tags"
              min-width="200"
            >
              <template #default="scope">
                <el-tag
                  v-for="tag in scope.row.tags"
                  :key="tag.id"
                  disable-transitions
                >
                  {{ tag.value }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column
              prop="updatedBy"
              label="Last Updated By"
              sortable
              width="200"
            />
            <el-table-column
              prop="updatedAt"
              label="Updated At (UTC)"
              :formatter="datetimeFormatter"
              sortable
              width="180"
            />
            <el-table-column
              prop="enabled"
              label="Enabled"
              sortable
              align="center"
              fixed="right"
              width="140"
              :filters="[{ text: 'Enabled', value: true }, { text: 'Disabled', value: false }]"
              :filter-method="filterStatus"
            >
              <template #default="scope">
                <el-tag
                  :type="scope.row.enabled ? 'primary' : 'danger'"
                  disable-transitions
                >
                  {{ scope.row.enabled ? "on" : "off" }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>

          <el-collapse
            class="deleted-flags-table"
            @change="fetchDeletedFlags"
          >
            <el-collapse-item title="Deleted Flags">
              <el-table
                :data="deletedFlags"
                :stripe="true"
                :highlight-current-row="false"
                :default-sort="{ prop: 'id', order: 'descending' }"
                style="width: 100%"
              >
                <el-table-column
                  prop="id"
                  align="center"
                  label="Flag ID"
                  sortable
                  fixed
                  width="95"
                />
                <el-table-column
                  prop="description"
                  label="Description"
                  min-width="300"
                />
                <el-table-column
                  prop="tags"
                  label="Tags"
                  min-width="200"
                >
                  <template #default="scope">
                    <el-tag
                      v-for="tag in scope.row.tags"
                      :key="tag.id"
                      disable-transitions
                    >
                      {{ tag.value }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="updatedBy"
                  label="Last Updated By"
                  sortable
                  width="200"
                />
                <el-table-column
                  prop="updatedAt"
                  label="Updated At (UTC)"
                  :formatter="datetimeFormatter"
                  sortable
                  width="180"
                />
                <el-table-column
                  prop="action"
                  label="Action"
                  align="center"
                  fixed="right"
                  width="100"
                >
                  <template #default="scope">
                    <el-button
                      type="warning"
                      size="small"
                      @click="restoreFlag(scope.row)"
                    >
                      Restore
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-collapse-item>
          </el-collapse>
        </div>
      </div>
    </el-col>
  </el-row>
</template>

<script setup>
import { ref, computed } from "vue";
import { useRouter } from "vue-router";
import Axios from "axios";
import { Search } from "@element-plus/icons-vue";
import { ElMessage, ElMessageBox } from "element-plus";

import constants from "@/constants";
import Spinner from "@/components/Spinner";
import helpers from "@/helpers/helpers";

const { handleErr } = helpers;
const { API_URL } = constants;

const router = useRouter();

const loaded = ref(false);
const deletedFlagsLoaded = ref(false);
const flags = ref([]);
const deletedFlags = ref([]);
const searchTerm = ref("");
const newFlag = ref({ description: "" });

// created() equivalent — runs at setup time
Axios.get(`${API_URL}/flags`).then(response => {
  let data = response.data;
  loaded.value = true;
  data.reverse();
  flags.value = data;
}, handleErr);

const filteredFlags = computed(() => {
  if (searchTerm.value) {
    return flags.value.filter(({ id, key, description, tags }) =>
      searchTerm.value
        .split(",")
        .map(term => {
          const termLowerCase = term.toLowerCase();
          return (
            id.toString().includes(term) ||
            key.includes(term) ||
            description.toLowerCase().includes(termLowerCase) ||
            tags
              .map(tag =>
                tag.value.toLowerCase().includes(termLowerCase)
              )
              .includes(true)
          );
        })
        .every(e => e)
    );
  }
  return flags.value;
});

function datetimeFormatter(row, col, val) {
  return val ? val.split(".")[0] : "";
}

function goToFlag(row) {
  router.push({ name: "flag", params: { flagId: row.id } });
}

function onCommandDropdown(command) {
  if (command === "simple_boolean_flag") {
    createFlag({ template: command });
  }
}

function createFlag(params) {
  if (!newFlag.value.description) {
    return;
  }
  Axios.post(`${API_URL}/flags`, {
    ...newFlag.value,
    ...(params || {})
  }).then(response => {
    let flag = response.data;
    newFlag.value.description = "";
    ElMessage.success("flag created");

    flag._new = true;
    flags.value.unshift(flag);
  }, handleErr);
}

function restoreFlag(row) {
  ElMessageBox.confirm('This will recover the deleted flag. Continue?', 'Warning', {
    confirmButtonText: 'OK',
    cancelButtonText: 'Cancel',
    type: 'warning'
  }).then(() => {
    Axios.put(`${API_URL}/flags/${row.id}/restore`).then(response => {
      let flag = response.data;
      ElMessage.success(`Flag updated`);
      flags.value.push(flag);
      deletedFlags.value = deletedFlags.value.filter(function(el) {
        return el.id != flag.id;
      });
    }, handleErr);
  });
}

function fetchDeletedFlags() {
  if (!deletedFlagsLoaded.value) {
    Axios.get(`${API_URL}/flags?deleted=true`).then(response => {
      let data = response.data;
      data.reverse();
      deletedFlags.value = data;
      deletedFlagsLoaded.value = true;
    }, handleErr);
  }
}

function filterStatus(value, row) {
  return row.enabled === value;
}
</script>

<style lang="less">
.flags-container {
  .el-table {
    margin-top: 2em;
  }
  .el-table__row {
    cursor: pointer;
  }
  .el-button-group .el-button--primary:first-child {
    border-right-color: #dcdfe6;
  }
  .deleted-flags-table {
    margin-top: 2rem;
  }
}
</style>
