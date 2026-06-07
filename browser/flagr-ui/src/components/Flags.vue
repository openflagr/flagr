<template>
  <div class="container">

        <spinner v-if="!loaded" />

        <div v-if="loaded">
          <!-- Toolbar: Search + Create Flag button -->
          <div class="flags-toolbar">
            <div class="flags-search-wrap">
              <el-icon class="flags-search-icon"><Search /></el-icon>
              <el-input
                placeholder="Search flags by ID, key, description, or tags..."
                v-model="searchTerm"
                v-focus
                size="large"
                data-testid="search-input"
                class="flags-search"
                clearable
              />
            </div>
            <el-button type="primary" size="large" @click="showCreateModal = true" data-testid="create-flag-btn">
              <el-icon><Plus /></el-icon>
              Create Flag
            </el-button>
          </div>

          <!-- Flags Table -->
          <el-card shadow="never" class="flags-table-card">
            <el-table
              :data="filteredFlags"
              :highlight-current-row="false"
              :default-sort="{ prop: 'id', order: 'descending' }"
              @row-click="goToFlag"
              virtual-scroll
              size="small"
              max-height="calc(100vh - 240px)"
              style="width: 100%"
              data-testid="flags-table"
            >
              <el-table-column prop="id" align="center" label="Flag ID" sortable width="100"></el-table-column>
              <el-table-column prop="description" label="Description" min-width="250">
                <template #default="scope">
                  <div class="flag-desc-cell">
                    <span class="flag-desc-text">{{ scope.row.description }}</span>
                    <span class="flag-key-tag" v-if="scope.row.key">key: {{ scope.row.key }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="tags" label="Tags" min-width="160">
                <template #default="scope">
                  <div class="flag-tags-cell">
                    <el-tag
                      v-for="tag in scope.row.tags"
                      :key="tag.id"
                      type="success"
                      size="small"
                      effect="light"
                    >{{ tag.value }}</el-tag>
                    <span v-if="!scope.row.tags.length" class="flags-empty-tag">—</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="updatedBy" label="Updated By" sortable width="140"></el-table-column>
              <el-table-column
                prop="updatedAt"
                label="Updated At"
                :formatter="datetimeFormatter"
                sortable
                width="130"
              ></el-table-column>
              <el-table-column
                prop="enabled"
                label="Status"
                sortable
                align="center"
                width="120"
                :filters="[{ text: 'Enabled', value: true }, { text: 'Disabled', value: false }]"
                :filter-method="filterStatus"
              >
                <template #default="scope">
                  <el-tag
                    :type="scope.row.enabled ? 'primary' : 'info'"
                    effect="light"
                    size="small"
                    round
                  >{{ scope.row.enabled ? "Enabled" : "Disabled" }}</el-tag>
                </template>
              </el-table-column>
            </el-table>
          </el-card>

          <!-- Deleted Flags -->
          <el-card shadow="never" class="flags-table-card">
            <template #header>
              <div class="el-card-header">
                <div class="flex-row">
                  <div class="flex-row-left"><h2>Deleted Flags</h2></div>
                  <div class="flex-row-right">
                  </div>
                </div>
              </div>
            </template>
            <el-collapse class="deleted-flags-inner" @change="fetchDeletedFlags" data-testid="deleted-flags-section">
              <el-collapse-item title="View deleted flags">
                <div v-if="!deletedFlagsLoaded" class="card--empty">Loading...</div>
                <div v-else-if="deletedFlags.length">
                  <el-table
                    :data="deletedFlags"
                    :highlight-current-row="false"
                    :default-sort="{ prop: 'id', order: 'descending' }"
                    virtual-scroll
                    size="small"
                    max-height="calc(100vh - 240px)"
                    style="width: 100%"
                  >
                    <el-table-column prop="id" align="center" label="Flag ID" sortable width="140"></el-table-column>
                    <el-table-column prop="description" label="Description" min-width="300"></el-table-column>
                    <el-table-column prop="tags" label="Tags" min-width="200">
                      <template #default="scope">
                        <el-tag
                          v-for="tag in scope.row.tags"
                          :key="tag.id"
                          type="success"
                          size="small"
                          effect="light"
                        >{{ tag.value }}</el-tag>
                      </template>
                    </el-table-column>
                    <el-table-column prop="updatedBy" label="Updated By" sortable width="200"></el-table-column>
                    <el-table-column
                      prop="updatedAt"
                      label="Updated At"
                      :formatter="datetimeFormatter"
                      sortable
                      width="180"
                    ></el-table-column>
              <el-table-column
                prop="action"
                label="Action"
                align="center"
                width="100"
              >
                      <template #default="scope">
                        <el-button
                          @click="restoreFlag(scope.row)"
                          type="warning"
                          size="small"
                          plain
                          :data-testid="'restore-flag-' + scope.row.id"
                        >Restore</el-button>
                      </template>
                    </el-table-column>
                  </el-table>
                </div>
                <div v-else class="card--empty">No deleted flags</div>
              </el-collapse-item>
            </el-collapse>
          </el-card>
        </div>
      </div>

  <!-- Create Flag Modal -->
  <el-dialog
    v-model="showCreateModal"
    title="Create Flag"
    width="500px"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <div class="create-flag-body">
      <label class="create-flag-label">Description</label>
      <el-input
        v-model="newFlag.description"
        placeholder="e.g. Enable new onboarding flow"
        data-testid="new-flag-input"
        size="large"
        @keyup.enter="createFlag"
      />
      <p class="create-flag-hint">A short description of what this feature flag controls.</p>
    </div>
    <template #footer>
      <el-button @click="showCreateModal = false">Cancel</el-button>
      <el-button
        type="primary"
        :disabled="!newFlag.description"
        @click="createFlag"
      >Create Flag</el-button>
      <el-dropdown
        trigger="click"
        @command="onCommandDropdown"
        v-if="newFlag.description"
      >
        <el-button type="primary" :disabled="!newFlag.description">
          <el-icon><ArrowDown /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item
              command="simple_boolean_flag"
              data-testid="create-boolean-flag-btn"
            >Create Simple Boolean Flag</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </template>
  </el-dialog>
</template>

<script>
import Axios from "axios";
import { Plus, Search, ArrowDown } from "@element-plus/icons-vue";

import constants from "@/constants";
import Spinner from "@/components/Spinner";
import helpers from "@/helpers/helpers";

const { handleErr, debounce } = helpers;
const { API_URL } = constants;


let flagsCache = null; // { flags: [], maxSnapshotID: number } | null

export default {
  name: "flags",
  components: {
    spinner: Spinner,
    Plus,
    Search,
    ArrowDown,
  },
  data() {
    const cached = flagsCache;
    return {
      loaded: !!cached,
      flags: cached ? cached.flags : [],
      deletedFlagsLoaded: false,
      deletedFlags: [],
      searchTerm: "",
      debouncedSearchTerm: "",
      showCreateModal: false,
      newFlag: {
        description: ""
      }
    };
  },

  beforeUnmount() {
    if (this._visHandler) document.removeEventListener('visibilitychange', this._visHandler);
  },

  mounted() {
    this._visHandler = () => { if (!document.hidden) this.refreshFlags(); };
    document.addEventListener('visibilitychange', this._visHandler);
  },

  created() {
    this._debouncedUpdate = debounce(() => {
      this.debouncedSearchTerm = this.searchTerm;
    }, 150);
    this.refreshFlags();
  },
  computed: {
    filteredFlags() {
      if (this.debouncedSearchTerm) {
        const terms = this.debouncedSearchTerm.split(",");
        return this.flags.filter(({ id, key, description, tags }) =>
          terms.every(term => {
            const t = term.toLowerCase();
            return (
              id.toString().includes(t) ||
              (key && key.toLowerCase().includes(t)) ||
              (description && description.toLowerCase().includes(t)) ||
              (tags && tags.some(tag => tag.value && tag.value.toLowerCase().includes(t)))
            );
          })
        );
      }
      return this.flags;
    },
  },
  watch: {
    searchTerm() {
      this._debouncedUpdate();
    },
  },
  methods: {
    async refreshFlags() {
      try {
        const { data: { maxID } } = await Axios.get(`${API_URL}/flags/snapshots/max_id`);
        if (flagsCache && maxID === flagsCache.maxSnapshotID) return;
        const { data } = await Axios.get(`${API_URL}/flags`);
        const flags = [...data].reverse();
        flagsCache = { flags, maxSnapshotID: maxID };
        this.flags = flags;
        this.loaded = true;
      } catch (err) {
        handleErr.call(this, err);
      }
    },
    datetimeFormatter(row, col, val) {
      return val ? val.split(".")[0] : "";
    },
    goToFlag(row) {
      this.$router.push({ name: "flag", params: { flagId: row.id } });
    },
    onCommandDropdown(command) {
      if (command === "simple_boolean_flag") {
        this.createFlag({ template: command });
      }
    },
    createFlag(params) {
      if (!this.newFlag.description) return;
      const payload = params ? { ...this.newFlag, ...params } : { ...this.newFlag };
      Axios.post(`${API_URL}/flags`, payload).then(response => {
        const flag = response.data;
        this.newFlag.description = "";
        this.showCreateModal = false;
        this.$message.success("Flag created");
        this.flags.unshift(flag);
        if (flagsCache) flagsCache.flags.unshift(flag);
      }, handleErr.bind(this));
    },
    restoreFlag(row) {
      this.$confirm("This will recover the deleted flag. Continue?", "Warning", {
        confirmButtonText: "OK",
        cancelButtonText: "Cancel",
        type: "warning"
      }).then(() => {
        Axios.put(`${API_URL}/flags/${row.id}/restore`).then(response => {
          const flag = response.data;
          this.$message.success(`Flag restored`);
          this.flags.push(flag);
          if (flagsCache) flagsCache.flags.push(flag);
          this.deletedFlags = this.deletedFlags.filter(el => el.id !== flag.id);
        }, handleErr.bind(this));
      });
    },
    fetchDeletedFlags() {
      if (!this.deletedFlagsLoaded) {
        Axios.get(`${API_URL}/flags?deleted=true`).then(response => {
          this.deletedFlags = [...response.data].reverse();
          this.deletedFlagsLoaded = true;
        }, handleErr.bind(this));
      }
    },
    filterStatus(value, row) {
      return row.enabled === value;
    }
  }
};
</script>

<style lang="less" scoped>
.flags-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.flags-search-wrap {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
}
.flags-search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  z-index: 1;
  font-size: 18px;
  color: var(--el-text-color-placeholder);
  pointer-events: none;
}
.flags-search :deep(.el-input__wrapper) {
  padding-left: 40px;
  border-radius: 24px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
  background-color: var(--el-bg-color);
  transition: box-shadow 0.2s;
}
.flags-search :deep(.el-input__wrapper):hover,
.flags-search :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}
.flags-table-card {
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  border-radius: 12px;
}
.flags-table-card :deep(.el-table) {
  border: none;
}
.flags-table-card :deep(.el-table) th.el-table__cell {
  background-color: var(--el-fill-color-light);
  color: var(--el-text-color-secondary);
  font-weight: 600;
  font-size: 12px;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}
.flags-table-card :deep(.el-table) .el-table__row {
  cursor: pointer;
}
.flags-table-card :deep(.el-table) .el-table__row:hover {
  background-color: var(--el-color-primary-light-9);
}
.flags-table-card :deep(.el-table) td.el-table__cell {
  border-bottom: 1px solid var(--el-border-color-lighter);
  white-space: nowrap;
  padding: 6px 0;
}
.flags-table-card :deep(.el-table) .el-table__header-wrapper tr:first-child th:first-child {
  border-top-left-radius: 8px;
}
.flags-table-card :deep(.el-table) .el-table__header-wrapper tr:first-child th:last-child {
  border-top-right-radius: 8px;
}
.flags-table-card :deep(.el-table) td.el-table__cell:first-child {
  white-space: nowrap;
}
.flag-desc-cell {
  display: flex;
  flex-direction: column;
}
.flag-desc-text {
  font-weight: 500;
  font-size: 13px;
  color: var(--el-text-color-primary);
}
.flag-key-tag {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
  font-family: "SF Mono", "Menlo", "Consolas", monospace;
  margin-top: 1px;
}
.flag-tags-cell {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
}
.flags-empty-tag {
  color: var(--el-text-color-placeholder);
  font-size: 13px;
}
.deleted-flags-inner {
  :deep(.el-collapse-item__header) {
    font-size: 12px;
    font-weight: 500;
    color: var(--el-text-color-secondary);
    padding: 4px 0;
  }
  :deep(.el-collapse-item__content) {
    padding-bottom: 4px;
  }
}
.create-flag-body {
  padding: 8px 0;
}
.create-flag-label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
  margin-bottom: 6px;
}
.create-flag-hint {
  margin: 6px 0 0;
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}
</style>
