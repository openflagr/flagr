<template>
  <div>
    <el-row>
      <el-col :span="20" :offset="2">
        <div class="webhooks-container container">
          <el-breadcrumb separator="/" v-if="!loading">
            <el-breadcrumb-item>Home page</el-breadcrumb-item>
            <el-breadcrumb-item>Webhooks</el-breadcrumb-item>
          </el-breadcrumb>

          <div class="header-actions" v-if="!loading">
            <el-button type="primary" @click="showCreateModal = true">
              Add Webhook
            </el-button>
          </div>

          <spinner v-if="loading" />

          <div v-else>
            <el-row>
              <el-input
                placeholder="Search webhooks"
                prefix-icon="el-icon-search"
                v-model="searchTerm"
                v-focus
              ></el-input>
            </el-row>

            <el-table
              :data="filteredWebhooks"
              :stripe="true"
              :highlight-current-row="false"
              :default-sort="{ prop: 'id', order: 'descending' }"
              style="width: 100%"
            >
              <el-table-column prop="id" align="center" label="ID" sortable fixed width="95"></el-table-column>
              <el-table-column prop="description" label="Description" min-width="300"></el-table-column>
              <el-table-column prop="url" label="URL" min-width="300"></el-table-column>
              <el-table-column prop="events" label="Events" min-width="200"></el-table-column>
              <el-table-column
                prop="enabled"
                label="Status"
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
                  >{{ scope.row.enabled ? "Enabled" : "Disabled" }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column
                prop="action"
                label="Action"
                align="center"
                fixed="right"
                width="200"
              >
                <template slot-scope="scope">
                  <el-button
                    @click="editWebhook(scope.row)"
                    type="primary"
                    size="small"
                  >Edit</el-button>
                  <el-button
                    @click="deleteWebhook(scope.row)"
                    type="danger"
                    size="small"
                  >Delete</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- Create/Edit Modal -->
    <div v-if="showCreateModal || showEditModal" class="modal-backdrop">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ showEditModal ? 'Edit Webhook' : 'Create Webhook' }}</h3>
          <button class="close" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <el-form
            ref="webhookForm"
            :model="webhookForm"
            :rules="rules"
            label-position="top"
            @submit.native.prevent="submitWebhook"
          >
            <el-form-item label="Description" prop="description">
              <el-input
                v-model="webhookForm.description"
                placeholder="Optional description"
              ></el-input>
            </el-form-item>

            <el-form-item label="URL" prop="url" required>
              <el-input
                v-model="webhookForm.url"
                placeholder="https://example.com/webhook"
                type="url"
              ></el-input>
            </el-form-item>

            <el-form-item label="Events" prop="events" required>
              <el-checkbox-group v-model="selectedEvents">
                <el-checkbox label="flag.created">Flag Created</el-checkbox>
                <el-checkbox label="flag.updated">Flag Updated</el-checkbox>
                <el-checkbox label="flag.deleted">Flag Deleted</el-checkbox>
                <el-checkbox label="flag.enabled">Flag Enabled</el-checkbox>
                <el-checkbox label="flag.disabled">Flag Disabled</el-checkbox>
                <el-checkbox label="segment.created">Segment Created</el-checkbox>
                <el-checkbox label="segment.updated">Segment Updated</el-checkbox>
                <el-checkbox label="segment.deleted">Segment Deleted</el-checkbox>
                <el-checkbox label="constraint.created">Constraint Created</el-checkbox>
                <el-checkbox label="constraint.updated">Constraint Updated</el-checkbox>
                <el-checkbox label="constraint.deleted">Constraint Deleted</el-checkbox>
                <el-checkbox label="variant.created">Variant Created</el-checkbox>
                <el-checkbox label="variant.updated">Variant Updated</el-checkbox>
                <el-checkbox label="variant.deleted">Variant Deleted</el-checkbox>
                <el-checkbox label="distribution.updated">Distribution Updated</el-checkbox>
              </el-checkbox-group>
              <div class="form-help-text">
                Select the events that should trigger this webhook
              </div>
            </el-form-item>

            <el-form-item label="Secret" prop="secret">
              <el-input
                v-model="webhookForm.secret"
                placeholder="Optional secret for webhook signature"
                show-password
              ></el-input>
            </el-form-item>

            <el-form-item>
              <el-checkbox v-model="webhookForm.enabled">
                Enabled
              </el-checkbox>
            </el-form-item>

            <div class="modal-footer">
              <el-button @click="closeModal">Cancel</el-button>
              <el-button
                type="primary"
                native-type="submit"
                :loading="submitting"
              >
                {{ showEditModal ? 'Save Changes' : 'Create Webhook' }}
              </el-button>
            </div>
          </el-form>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import Spinner from './Spinner.vue'

export default {
  name: 'Webhooks',
  components: {
    Spinner
  },
  props: {
    flagId: {
      type: [String, Number],
      required: false
    }
  },
  data() {
    return {
      loading: true,
      submitting: false,
      webhooks: [],
      showCreateModal: false,
      showEditModal: false,
      webhookForm: {
        description: '',
        url: '',
        events: '',
        secret: '',
        enabled: true
      },
      selectedEvents: [],
      rules: {
        url: [
          { required: true, message: 'Please enter webhook URL', trigger: 'blur' },
          { type: 'url', message: 'Please enter a valid URL', trigger: 'blur' }
        ],
        events: [
          { 
            required: true, 
            validator: (rule, value, callback) => {
              if (this.selectedEvents.length === 0) {
                callback(new Error('Please select at least one event'))
              } else {
                callback()
              }
            },
            trigger: 'change'
          }
        ]
      },
      editingWebhookId: null,
      searchTerm: ''
    }
  },
  watch: {
    selectedEvents: {
      handler(newVal) {
        this.webhookForm.events = newVal.join(',')
      },
      immediate: true
    }
  },
  created() {
    this.fetchWebhooks()
  },
  computed: {
    filteredWebhooks() {
      if (this.searchTerm) {
        return this.webhooks.filter(webhook =>
          webhook.description.toLowerCase().includes(this.searchTerm.toLowerCase()) ||
          webhook.url.toLowerCase().includes(this.searchTerm.toLowerCase()) ||
          webhook.events.toLowerCase().includes(this.searchTerm.toLowerCase())
        )
      }
      return this.webhooks
    }
  },
  methods: {
    async fetchWebhooks() {
      try {
        const response = await axios.get('/api/v1/webhooks')
        this.webhooks = response.data
      } catch (error) {
        console.error('Failed to fetch webhooks:', error)
      } finally {
        this.loading = false
      }
    },
    editWebhook(webhook) {
      this.editingWebhookId = webhook.id
      this.webhookForm = {
        description: webhook.description || '',
        url: webhook.url,
        events: webhook.events,
        secret: webhook.secret || '',
        enabled: webhook.enabled
      }
      this.selectedEvents = webhook.events.split(',').filter(Boolean)
      this.showEditModal = true
    },
    async deleteWebhook(webhook) {
      if (!confirm('Are you sure you want to delete this webhook?')) {
        return
      }
      try {
        await axios.delete(`/api/v1/webhooks/${webhook.id}`)
        this.webhooks = this.webhooks.filter(w => w.id !== webhook.id)
      } catch (error) {
        console.error('Failed to delete webhook:', error)
        alert('Failed to delete webhook. Please try again.')
      }
    },
    async submitWebhook() {
      try {
        this.submitting = true
        const baseUrl = '/api/v1/webhooks'
        
        if (this.showEditModal) {
          const response = await axios.put(
            `${baseUrl}/${this.editingWebhookId}`,
            this.webhookForm
          )
          const updatedWebhooks = [...this.webhooks]
          const index = updatedWebhooks.findIndex(w => w.id === this.editingWebhookId)
          if (index !== -1) {
            updatedWebhooks[index] = response.data
            this.webhooks = updatedWebhooks
          }
        } else {
          const response = await axios.post(
            baseUrl,
            this.webhookForm
          )
          this.webhooks = [...this.webhooks, response.data]
        }
        this.closeModal()
      } catch (error) {
        console.error('Failed to save webhook:', error)
        this.$message.error('Failed to save webhook. Please try again.')
      } finally {
        this.submitting = false
      }
    },
    closeModal() {
      this.showCreateModal = false
      this.showEditModal = false
      this.editingWebhookId = null
      this.webhookForm = {
        description: '',
        url: '',
        events: '',
        secret: '',
        enabled: true
      }
      this.selectedEvents = []
    },
    filterStatus(value, row) {
      return row.enabled === value
    }
  }
}
</script>

<style scoped>
.webhooks-container {
  margin-top: 20px;
}

.header-actions {
  margin: 20px 0;
  display: flex;
  justify-content: flex-end;
}

.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal {
  background: #fff;
  border-radius: 4px;
  width: 600px;
  max-width: 90%;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.modal-header {
  padding: 20px;
  border-bottom: 1px solid #ebeef5;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: #303133;
}

.modal-body {
  padding: 20px;
}

.modal-footer {
  padding: 20px;
  border-top: 1px solid #ebeef5;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.form-help-text {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.close {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  padding: 0;
  color: #909399;
}

.close:hover {
  color: #303133;
}

.loading {
  display: flex;
  justify-content: center;
  padding: 40px;
}

.el-checkbox-group {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 10px;
  margin-bottom: 10px;
}

.el-checkbox {
  margin-right: 0;
  margin-bottom: 0;
}
</style> 