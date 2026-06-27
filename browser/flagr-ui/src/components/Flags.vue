<template>
  <div class="container">
    <spinner v-if="!loaded" />

    <div v-if="loaded">
      <!-- Toolbar: Search + Create Flag button -->
      <div class="flags-toolbar">
        <div class="flags-search-wrap">
          <el-icon class="flags-search-icon">
            <Search />
          </el-icon>
          <el-input
            v-model="searchTerm"
            v-focus
            placeholder="Search flags by ID, key, description, or tags..."
            size="large"
            data-testid="search-input"
            class="flags-search"
            clearable
          />
        </div>
        <el-button
          type="primary"
          size="large"
          data-testid="create-flag-btn"
          @click="showCreateModal = true"
        >
          <el-icon><Plus /></el-icon>
          Create Flag
        </el-button>
      </div>

      <!-- Flags Table -->
      <el-card
        shadow="never"
        class="flags-table-card"
      >
        <el-table
          :data="filteredFlags"
          :highlight-current-row="false"
          :default-sort="{ prop: 'id', order: 'descending' }"
          virtual-scroll
          size="small"
          max-height="calc(100vh - 240px)"
          style="width: 100%"
          data-testid="flags-table"
          @row-click="goToRow"
        >
          <el-table-column
            prop="id"
            align="center"
            label="Flag ID"
            sortable
            width="100"
          />
          <el-table-column
            prop="description"
            label="Description"
            min-width="250"
          >
            <template #default="scope">
              <div class="flag-desc-cell">
                <span class="flag-desc-text">{{ scope.row.description }}</span>
                <span
                  v-if="scope.row.key"
                  class="flag-key-tag"
                >key: {{ scope.row.key }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column
            prop="tags"
            label="Tags"
            min-width="160"
          >
            <template #default="scope">
              <div class="flag-tags-cell">
                <el-tag
                  v-for="tag in scope.row.tags"
                  :key="tag.id"
                  type="success"
                  size="small"
                  effect="light"
                >
                  {{ tag.value }}
                </el-tag>
                <span
                  v-if="!scope.row.tags.length"
                  class="flags-empty-tag"
                >—</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column
            prop="updatedBy"
            label="Updated By"
            sortable
            width="140"
          />
          <el-table-column
            prop="updatedAt"
            label="Updated At"
            :formatter="datetimeFormatter"
            sortable
            width="130"
          />
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
              >
                {{ scope.row.enabled ? "Enabled" : "Disabled" }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- Deleted Flags -->
      <el-card
        shadow="never"
        class="flags-table-card"
      >
        <template #header>
          <div class="el-card-header">
            <div class="flex-row">
              <div class="flex-row-left">
                <h2>Deleted Flags</h2>
              </div>
              <div class="flex-row-right" />
            </div>
          </div>
        </template>
        <el-collapse
          class="deleted-flags-inner"
          data-testid="deleted-flags-section"
          @change="flagsListPage.fetchDeletedFlags(page)"
        >
          <el-collapse-item title="View deleted flags">
            <div
              v-if="!deletedFlagsLoaded"
              class="card--empty"
            >
              Loading...
            </div>
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
                <el-table-column
                  prop="id"
                  align="center"
                  label="Flag ID"
                  sortable
                  width="140"
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
                      type="success"
                      size="small"
                      effect="light"
                    >
                      {{ tag.value }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="updatedBy"
                  label="Updated By"
                  sortable
                  width="200"
                />
                <el-table-column
                  prop="updatedAt"
                  label="Updated At"
                  :formatter="datetimeFormatter"
                  sortable
                  width="180"
                />
                <el-table-column
                  prop="action"
                  label="Action"
                  align="center"
                  width="100"
                >
                  <template #default="scope">
                    <el-button
                      type="warning"
                      size="small"
                      plain
                      :data-testid="'restore-flag-' + scope.row.id"
                      @click="flagsListPage.restoreFlag(page, scope.row)"
                    >
                      Restore
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
            <div
              v-else
              class="card--empty"
            >
              No deleted flags
            </div>
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
        @keyup.enter="flagsListPage.createFlag(page)"
      />
      <p class="create-flag-hint">
        A short description of what this feature flag controls.
      </p>

      <div class="create-flag-options">
        <div class="create-flag-option">
          <el-button
            type="primary"
            :disabled="!newFlag.description"
            data-testid="create-flag-submit-btn"
            @click="flagsListPage.createFlag(page)"
          >
            Create Flag
          </el-button>
          <span class="create-flag-option-sub">Blank flag — configure variants and segments yourself</span>
        </div>
        <div class="create-flag-option">
          <el-button
            type="primary"
            plain
            :disabled="!newFlag.description"
            data-testid="create-boolean-flag-btn"
            @click="flagsListPage.createBooleanFlag(page)"
          >
            Create Boolean Flag
          </el-button>
          <span class="create-flag-option-sub">Ready-to-use flag with on/off variants and a 100% rollout</span>
        </div>
      </div>
    </div>
    <template #footer>
      <el-button @click="showCreateModal = false">
        Cancel
      </el-button>
    </template>
  </el-dialog>
</template>


<script lang="ts">
import { Plus, Search } from '@element-plus/icons-vue'
import Spinner from '@/components/Spinner.vue'
import { getFlagsCache } from '@/pages/flagsList'
import helpers from '@/helpers/helpers'
import { castFlagsList } from '@/helpers/vuePageCast'
import * as flagsListPage from '@/pages/flagsListPage'
import {
  datetimeFormatter,
  filterStatus,
  mountFlagsList,
} from '@/pages/flagsListPage'
import type { Flag } from '@/api/types'

const { debounce } = helpers

export default {
  name: 'Flags',
  components: {
    spinner: Spinner,
    Plus,
    Search,
  },
  data() {
    const cached = getFlagsCache()
    return {
      flagsListPage,
      loaded: !!cached,
      flags: cached ? cached.flags : [] as Flag[],
      deletedFlagsLoaded: false,
      deletedFlags: [] as Flag[],
      searchTerm: '',
      debouncedSearchTerm: '',
      showCreateModal: false,
      newFlag: {
        description: '',
      },
      visHandler: null as (() => void) | null,
      debouncedUpdate: null as (() => void) | null,
    }
  },
  computed: {
    page() {
      return castFlagsList(this)
    },
    filteredFlags(): Flag[] {
      if (this.debouncedSearchTerm) {
        const terms = this.debouncedSearchTerm.split(',')
        return this.flags.filter(({ id, key, description, tags }) =>
          terms.every((term) => {
            const t = term.toLowerCase()
            return (
              id?.toString().includes(t) ||
              (key && key.toLowerCase().includes(t)) ||
              (description && description.toLowerCase().includes(t)) ||
              (tags && tags.some((tag) => tag.value && tag.value.toLowerCase().includes(t)))
            )
          }),
        )
      }
      return this.flags
    },
  },
  watch: {
    searchTerm() {
      this.debouncedUpdate?.()
    },
  },

  beforeUnmount() {
    if (this.visHandler) document.removeEventListener('visibilitychange', this.visHandler)
  },

  mounted() {
    this.visHandler = () => {
      if (!document.hidden) flagsListPage.refreshFlags(this.page)
    }
    document.addEventListener('visibilitychange', this.visHandler)
  },

  created() {
    this.debouncedUpdate = debounce(() => {
      this.debouncedSearchTerm = this.searchTerm
    }, 150)
    mountFlagsList(this.page)
  },
  methods: {
    goToRow(row: Flag) {
      flagsListPage.goToFlag(this.page, row)
    },
    datetimeFormatter,
    filterStatus,
  },
}
</script>

<style lang="scss" scoped>
.flags-toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  margin-bottom: var(--space-sm);
}
.flags-search-wrap {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
}
.flags-search-icon {
  position: absolute;
  left: var(--space-xs);
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
  padding: var(--space-2xs) 0;
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
  gap: var(--space-3xs);
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
    padding: var(--space-3xs) 0;
  }
  :deep(.el-collapse-item__content) {
    padding-bottom: var(--space-3xs);
  }
}
.create-flag-body {
  padding: var(--space-2xs) 0;
}
.create-flag-label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
  margin-bottom: var(--space-2xs);
}
.create-flag-hint {
  margin: var(--space-2xs) 0 0;
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}
.create-flag-options {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  margin-top: var(--space-md);
  padding-top: var(--space-md);
  border-top: 1px solid var(--el-border-color-light);
}
.create-flag-option {
  display: flex;
  flex-direction: column;
  gap: var(--space-3xs);
  align-items: flex-start;
}
.create-flag-option-sub {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  line-height: 1.4;
}
@media (max-width: 640px) {
  .flags-toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: var(--space-2xs);
  }
  .flags-search-wrap {
    width: 100%;
  }
}
</style>
