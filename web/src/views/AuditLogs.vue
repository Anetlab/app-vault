<template>
  <div class="audit-logs-view">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold text-cyan-400">Audit Logs</h1>
      <button
        @click="fetchAuditLogs"
        :disabled="loading"
        class="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition-colors disabled:opacity-50 flex items-center gap-2"
      >
        <i :class="['fas', loading ? 'fa-spinner fa-spin' : 'fa-sync-alt']"></i>
        Refresh
      </button>
    </div>

    <!-- Filters -->
    <div class="bg-slate-800 rounded-lg p-4 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div>
          <label class="block text-slate-300 text-sm mb-2">Action Type</label>
          <select
            v-model="filters.action"
            @change="fetchAuditLogs"
            class="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white focus:outline-none focus:border-cyan-500"
          >
            <option value="">All Actions</option>
            <option value="login">Login</option>
            <option value="logout">Logout</option>
            <option value="create_secret">Create Secret</option>
            <option value="read_secret">Read Secret</option>
            <option value="update_secret">Update Secret</option>
            <option value="delete_secret">Delete Secret</option>
            <option value="rotate_key">Rotate Key</option>
            <option value="create_service_principal">Create Service Principal</option>
            <option value="delete_service_principal">Delete Service Principal</option>
          </select>
        </div>

        <div>
          <label class="block text-slate-300 text-sm mb-2">Status</label>
          <select
            v-model="filters.status"
            @change="fetchAuditLogs"
            class="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white focus:outline-none focus:border-cyan-500"
          >
            <option value="">All Statuses</option>
            <option value="success">Success</option>
            <option value="failure">Failure</option>
          </select>
        </div>

        <div>
          <label class="block text-slate-300 text-sm mb-2">From Date</label>
          <input
            v-model="filters.from_date"
            type="datetime-local"
            @change="fetchAuditLogs"
            class="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white focus:outline-none focus:border-cyan-500"
          />
        </div>

        <div>
          <label class="block text-slate-300 text-sm mb-2">To Date</label>
          <input
            v-model="filters.to_date"
            type="datetime-local"
            @change="fetchAuditLogs"
            class="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white focus:outline-none focus:border-cyan-500"
          />
        </div>
      </div>

      <div class="mt-4 flex justify-between items-center">
        <div class="text-sm text-slate-400">
          Showing {{ auditLogs.length }} {{ auditLogs.length === 1 ? 'entry' : 'entries' }}
        </div>
        <button
          @click="resetFilters"
          class="text-sm text-cyan-400 hover:text-cyan-300 transition-colors"
        >
          Reset Filters
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-center py-8">
      <i class="fas fa-spinner fa-spin text-4xl text-cyan-400"></i>
    </div>

    <div v-else-if="auditLogs.length === 0" class="bg-slate-800 rounded-lg p-8 text-center">
      <i class="fas fa-inbox text-6xl text-slate-600 mb-4"></i>
      <h2 class="text-xl font-semibold text-slate-400 mb-2">No Audit Logs Found</h2>
      <p class="text-slate-500">Try adjusting your filters or check back later</p>
    </div>

    <div v-else class="bg-slate-800 rounded-lg overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-slate-900">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
                Timestamp
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
                Action
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
                Resource
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
                Status
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
                IP Address
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
                Details
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-700">
            <tr
              v-for="log in auditLogs"
              :key="log.id"
              class="hover:bg-slate-700 transition-colors"
            >
              <td class="px-6 py-4 whitespace-nowrap text-slate-300">
                {{ formatTimestamp(log.createdAt) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center gap-2">
                  <i :class="['fas', getActionIcon(log.action), getActionColor(log.action)]"></i>
                  <span class="text-white font-medium">{{ formatAction(log.action) }}</span>
                </div>
              </td>
              <td class="px-6 py-4 text-slate-300">
                {{ log.resourceType || '-' }}
                <span v-if="log.resourceId" class="text-slate-500 text-sm">
                  ({{ log.resourceId.substring(0, 8) }}...)
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  :class="[
                    'px-2 py-1 rounded-full text-xs font-medium',
                    log.status === 'success'
                      ? 'bg-green-500/20 text-green-400'
                      : 'bg-red-500/20 text-red-400'
                  ]"
                >
                  {{ log.status }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-slate-400 font-mono text-sm">
                {{ log.ipAddress || '-' }}
              </td>
              <td class="px-6 py-4">
                <button
                  v-if="log.details"
                  @click="viewDetails(log)"
                  class="text-cyan-400 hover:text-cyan-300 transition-colors"
                >
                  <i class="fas fa-eye"></i>
                </button>
                <span v-else class="text-slate-600">-</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Details Modal -->
    <div
      v-if="showDetailsModal && selectedLog"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      @click.self="showDetailsModal = false"
    >
      <div class="bg-slate-800 rounded-lg max-w-2xl w-full p-6">
        <div class="flex justify-between items-start mb-6">
          <h2 class="text-2xl font-bold text-cyan-400">Audit Log Details</h2>
          <button
            @click="showDetailsModal = false"
            class="text-slate-400 hover:text-white transition-colors"
          >
            <i class="fas fa-times text-xl"></i>
          </button>
        </div>

        <div class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-slate-400 text-sm mb-1">Timestamp</label>
              <div class="text-white">{{ formatTimestamp(selectedLog.createdAt) }}</div>
            </div>
            <div>
              <label class="block text-slate-400 text-sm mb-1">Status</label>
              <span
                :class="[
                  'inline-block px-3 py-1 rounded-full text-sm font-medium',
                  selectedLog.status === 'success'
                    ? 'bg-green-500/20 text-green-400'
                    : 'bg-red-500/20 text-red-400'
                ]"
              >
                {{ selectedLog.status }}
              </span>
            </div>
          </div>

          <div>
            <label class="block text-slate-400 text-sm mb-1">Action</label>
            <div class="text-white font-medium">{{ formatAction(selectedLog.action) }}</div>
          </div>

          <div v-if="selectedLog.resourceType" class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-slate-400 text-sm mb-1">Resource Type</label>
              <div class="text-white">{{ selectedLog.resourceType }}</div>
            </div>
            <div v-if="selectedLog.resourceId">
              <label class="block text-slate-400 text-sm mb-1">Resource ID</label>
              <code class="text-cyan-400 font-mono text-sm">{{ selectedLog.resourceId }}</code>
            </div>
          </div>

          <div v-if="selectedLog.ipAddress">
            <label class="block text-slate-400 text-sm mb-1">IP Address</label>
            <code class="text-white font-mono">{{ selectedLog.ipAddress }}</code>
          </div>

          <div v-if="selectedLog.userAgent">
            <label class="block text-slate-400 text-sm mb-1">User Agent</label>
            <div class="text-slate-300 text-sm">{{ selectedLog.userAgent }}</div>
          </div>

          <div v-if="selectedLog.details">
            <label class="block text-slate-400 text-sm mb-1">Additional Details</label>
            <pre class="bg-slate-900 rounded p-3 text-slate-300 text-sm overflow-x-auto">{{ JSON.stringify(selectedLog.details, null, 2) }}</pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { systemApi } from '../api/system'
import { useToastStore } from '../stores/toast'
import type { AuditLog } from '../types'

const toastStore = useToastStore()

const auditLogs = ref<AuditLog[]>([])
const loading = ref(false)
const showDetailsModal = ref(false)
const selectedLog = ref<AuditLog | null>(null)

const filters = ref({
  action: '',
  status: '',
  from_date: '',
  to_date: ''
})

const fetchAuditLogs = async () => {
  loading.value = true
  try {
    const params: any = {}
    if (filters.value.action) params.action = filters.value.action
    if (filters.value.status) params.status = filters.value.status
    if (filters.value.from_date) params.from_date = filters.value.from_date
    if (filters.value.to_date) params.to_date = filters.value.to_date

    auditLogs.value = await systemApi.getAuditLogs(params)
  } catch (error: any) {
    toastStore.error(error.response?.data?.error || 'Failed to fetch audit logs')
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.value = {
    action: '',
    status: '',
    from_date: '',
    to_date: ''
  }
  fetchAuditLogs()
}

const viewDetails = (log: AuditLog) => {
  selectedLog.value = log
  showDetailsModal.value = true
}

const formatTimestamp = (timestamp: string): string => {
  return new Date(timestamp).toLocaleString()
}

const formatAction = (action: string): string => {
  return action
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

const getActionIcon = (action: string): string => {
  const iconMap: Record<string, string> = {
    login: 'fa-sign-in-alt',
    logout: 'fa-sign-out-alt',
    create_secret: 'fa-plus',
    read_secret: 'fa-eye',
    update_secret: 'fa-edit',
    delete_secret: 'fa-trash',
    rotate_key: 'fa-sync-alt',
    create_service_principal: 'fa-user-plus',
    delete_service_principal: 'fa-user-times'
  }
  return iconMap[action] || 'fa-circle'
}

const getActionColor = (action: string): string => {
  if (action.includes('delete')) return 'text-red-400'
  if (action.includes('create') || action.includes('login')) return 'text-green-400'
  if (action.includes('update') || action.includes('rotate')) return 'text-yellow-400'
  if (action.includes('read')) return 'text-cyan-400'
  return 'text-slate-400'
}

onMounted(() => {
  fetchAuditLogs()
})
</script>

<style scoped>
.audit-logs-view {
  animation: fadeIn 0.3s ease-in;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
