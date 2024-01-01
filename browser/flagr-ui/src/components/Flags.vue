<template>
  <el-row>
    <el-col :span="20" :offset="2">
      <div class="flags-container container">
        <el-breadcrumb separator="/" v-if="loaded">
          <el-breadcrumb-item>Home page</el-breadcrumb-item>
        </el-breadcrumb>

        <spinner v-if="!loaded" />

        <div v-if="loaded">
          <el-row>
            <el-col>
              <el-input placeholder="Specific new flag description" v-model="newFlag.description">
                <template slot="prepend">
                  <span class="el-icon-plus" />
                </template>
                <template slot="append">
                  <el-dropdown
                    split-button
                    type="primary"
                    :disabled="!newFlag.description"
                    @command="onCommandDropdown"
                    @click.prevent="createFlag"
                  >
                    Create New Flag
                    <el-dropdown-menu slot="dropdown">
                      <el-dropdown-item
                        command="simple_boolean_flag"
                        :disabled="!newFlag.description"
                      >Create Simple Boolean Flag</el-dropdown-item>
                    </el-dropdown-menu>
                  </el-dropdown>
                </template>
              </el-input>
            </el-col>
          </el-row>

          <el-row>
            <el-input
              placeholder="Search a flag"
              prefix-icon="el-icon-search"
              v-model="searchTerm"
              v-focus
            ></el-input>
          </el-row>

          <el-table
            :data="filteredFlags"
            :stripe="true"
            :highlight-current-row="false"
            :default-sort="{ prop: 'id', order: 'descending' }"
            v-on:row-click="goToFlag"
            style="width: 100%"
          >
            <el-table-column prop="id" align="center" label="Flag ID" sortable fixed width="95"></el-table-column>
            <el-table-column prop="description" label="Description" min-width="300"></el-table-column>
            <el-table-column prop="tags" label="Tags" min-width="200">
              <template slot-scope="scope">
                <el-tag
                  v-for="tag in scope.row.tags"
                  :key="tag.id"
                  :type="warning"
                  disable-transitions
                >{{ tag.value }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="updatedBy" label="Last Updated By" sortable width="200"></el-table-column>
            <el-table-column
              prop="updatedAt"
              label="Updated At (UTC)"
              :formatter="datetimeFormatter"
              sortable
              width="180"
            ></el-table-column>
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
              <template slot-scope="scope">
                <el-tag
                  :type="scope.row.enabled ? 'primary' : 'danger'"
                  disable-transitions
                >{{ scope.row.enabled ? "on" : "off" }}</el-tag>
              </template>
            </el-table-column>
          </el-table>

          <el-collapse class="deleted-flags-table" @change="fetchDeletedFlags">
            <el-collapse-item title="Deleted Flags">
              <el-table
                :data="getDeletedFlags"
                :stripe="true"
                :highlight-current-row="false"
                :default-sort="{ prop: 'id', order: 'descending' }"
                style="width: 100%"
              >
                <el-table-column prop="id" align="center" label="Flag ID" sortable fixed width="95"></el-table-column>
                <el-table-column prop="description" label="Description" min-width="300"></el-table-column>
                <el-table-column prop="tags" label="Tags" min-width="200">
                  <template slot-scope="scope">
                    <el-tag
                      v-for="tag in scope.row.tags"
                      :key="tag.id"
                      :type="warning"
                      disable-transitions
                    >{{ tag.value }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="updatedBy" label="Last Updated By" sortable width="200"></el-table-column>
                <el-table-column
                  prop="updatedAt"
                  label="Updated At (UTC)"
                  :formatter="datetimeFormatter"
                  sortable
                  width="180"
                ></el-table-column>
                <el-table-column
                  prop="action"
                  label="Action"
                  align="center"
                  fixed="right"
                  width="100"
                >
                  <template slot-scope="scope">
                    <el-button
                      @click="restoreFlag(scope.row)"
                      type="warning"
                      size="small"
                    >Restore</el-button>
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

<script>
import Axios from "axios";

import constants from "@/constants";
import Spinner from "@/components/Spinner";
import helpers from "@/helpers/helpers";

const { handleErr } = helpers;

const { API_URL } = constants;

export default {
  name: "flags",
  components: {
    spinner: Spinner
  },
  data() {
    return {
      loaded: false,
      deletedFlagsLoaded: false,
      flags: [],
      deletedFlags: [],
      searchTerm: "",
      newFlag: {
        description: ""
      }
    };
  },
  created() {
    Axios.get(`${API_URL}/flags`).then(response => {
      let flags = response.data;
      this.loaded = true;
      flags.reverse();
      this.flags = flags;
    }, handleErr.bind(this));
  },
  computed: {
    filteredFlags: function() {
      if (this.searchTerm) {
        return this.flags.filter(({ id, key, description, tags }) =>
          this.searchTerm
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
      return this.flags;
    },
    getDeletedFlags: function() {
      return this.deletedFlags;
    }
  },
  methods: {
    flagEnabledFormatter(row, col, val) {
      return val ? "on" : "off";
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
      if (!this.newFlag.description) {
        return;
      }
      Axios.post(`${API_URL}/flags`, {
        ...this.newFlag,
        ...(params || {})
      }).then(response => {
        let flag = response.data;
        this.newFlag.description = "";
        this.$message.success("flag created");

        flag._new = true;
        this.flags.unshift(flag);
      }, handleErr.bind(this));
    },
    restoreFlag(row) {
      this.$confirm('This will recover the deleted flag. Continue?', 'Warning', {
        confirmButtonText: 'OK',
        cancelButtonText: 'Cancel',
        type: 'warning'
      }).then(() => {
        Axios.put(`${API_URL}/flags/${row.id}/restore`).then(response => {
          let flag = response.data;
          this.$message.success(`Flag updated`);
          this.flags.push(flag);
          this.deletedFlags = this.deletedFlags.filter(function(el) {
            return el.id != flag.id;
          });
        }, handleErr.bind(this));
      });

    },
    fetchDeletedFlags() {
      if (!this.deletedFlagsLoaded) {
        var self = this;
        Axios.get(`${API_URL}/flags?deleted=true`).then(response => {
          let flags = response.data;
          flags.reverse();
          self.deletedFlags = flags;
          self.deletedFlagsLoaded = true;
        }, handleErr.bind(this));
      }
    },
    filterStatus(value, row) {
      return row.enabled === value;
    }
  }
};
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
