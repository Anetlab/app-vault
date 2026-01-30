<template>
  <div class="service-principals-view">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold text-cyan-400">Service Principals</h1>
      <button
        @click="showCreateModal = true"
        class="px-4 py-2 bg-cyan-600 hover:bg-cyan-700 text-white rounded-lg transition-colors flex items-center gap-2"
      >
        <i class="fas fa-plus"></i>
        New Service Principal
      </button>
    </div>

    <div v-if="loading" class="text-center py-8">
      <i class="fas fa-spinner fa-spin text-4xl text-cyan-400"></i>
    </div>

    <div v-else-if="servicePrincipals.length === 0" class="bg-slate-800 rounded-lg p-8 text-center">
      <i class="fas fa-robot text-6xl text-slate-600 mb-4"></i>
      <h2 class="text-xl font-semibold text-slate-400 mb-2">No Service Principals Yet</h2>
      <p class="text-slate-500 mb-4">Create a service principal to enable programmatic access to your vault</p>
      <button
        @click="showCreateModal = true"
        class="px-6 py-3 bg-cyan-600 hover:bg-cyan-700 text-white rounded-lg transition-colors"
      >
        Create Your First Service Principal
      </button>
    </div>

    <div v-else class="bg-slate-800 rounded-lg overflow-hidden">
      <table class="w-full">
        <thead class="bg-slate-900">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
              Name
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
              Description
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
              Status
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
              Created
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-slate-400 uppercase tracking-wider">
              Last Used
            </th>
            <th class="px-6 py-3 text-right text-xs font-medium text-slate-400 uppercase tracking-wider">
              Actions
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-700">
          <tr
            v-for="sp in servicePrincipals"
            :key="sp.id"
            class="hover:bg-slate-700 transition-colors"
          >
            <td class="px-6 py-4">
              <div class="flex items-center gap-3">
                <i class="fas fa-robot text-cyan-400"></i>
                <span class="font-medium text-white">{{ sp.name }}</span>
              </div>
            </td>
            <td class="px-6 py-4 text-slate-300">
              {{ sp.description || '-' }}
            </td>
            <td class="px-6 py-4">
              <span
                :class="[
                  'px-2 py-1 rounded-full text-xs font-medium',
                  sp.is_active ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
                ]"
              >
                {{ sp.is_active ? 'Active' : 'Inactive' }}
              </span>
            </td>
            <td class="px-6 py-4 text-slate-400">
              {{ formatDate(sp.created_at) }}
            </td>
            <td class="px-6 py-4 text-slate-400">
              {{ sp.last_used_at ? formatDate(sp.last_used_at) : 'Never' }}
            </td>
            <td class="px-6 py-4 text-right">
              <div class="flex items-center justify-end gap-2">
                <button
                  @click="viewServicePrincipal(sp)"
                  class="px-3 py-1 bg-slate-700 hover:bg-slate-600 text-cyan-400 rounded transition-colors"
                  title="View Details"
                >
                  <i class="fas fa-eye"></i>
                </button>
                <button
                  @click="confirmRegenerate(sp)"
                  class="px-3 py-1 bg-yellow-600/20 hover:bg-yellow-600/30 text-yellow-400 rounded transition-colors"
                  title="Regenerate Secret"
                >
                  <i class="fas fa-sync-alt"></i>
                </button>
                <button
                  @click="confirmDelete(sp)"
                  class="px-3 py-1 bg-red-600/20 hover:bg-red-600/30 text-red-400 rounded transition-colors"
                  title="Delete"
                >
                  <i class="fas fa-trash"></i>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create Modal -->
    <div
      v-if="showCreateModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      @click.self="showCreateModal = false"
    >
      <div class="bg-slate-800 rounded-lg max-w-md w-full p-6">
        <h2 class="text-2xl font-bold text-cyan-400 mb-4">Create Service Principal</h2>
        <form @submit.prevent="handleCreate">
          <div class="mb-4">
            <label class="block text-slate-300 mb-2">Name *</label>
            <input
              v-model="createForm.name"
              type="text"
              required
              class="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white focus:outline-none focus:border-cyan-500"
              placeholder="my-app-service"
            />
          </div>
          <div class="mb-4">
            <label class="block text-slate-300 mb-2">Description</label>
            <textarea
              v-model="createForm.description"
              rows="3"
              class="w-full px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-white focus:outline-none focus:border-cyan-500"
              placeholder="Service principal for production deployment"
            ></textarea>
          </div>
          <div class="flex justify-end gap-2">
            <button
              type="button"
              @click="showCreateModal = false"
              class="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="createLoading"
              class="px-4 py-2 bg-cyan-600 hover:bg-cyan-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              <i v-if="createLoading" class="fas fa-spinner fa-spin"></i>
              Create
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- View Modal -->
    <div
      v-if="showViewModal && selectedServicePrincipal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      @click.self="showViewModal = false"
    >
      <div class="bg-slate-800 rounded-lg max-w-2xl w-full p-6">
        <div class="flex justify-between items-start mb-6">
          <h2 class="text-2xl font-bold text-cyan-400">Service Principal Details</h2>
          <button
            @click="showViewModal = false"
            class="text-slate-400 hover:text-white transition-colors"
          >
            <i class="fas fa-times text-xl"></i>
          </button>
        </div>

        <div class="space-y-4">
          <div>
            <label class="block text-slate-400 text-sm mb-1">Name</label>
            <div class="text-white font-medium">{{ selectedServicePrincipal.name }}</div>
          </div>

          <div>
            <label class="block text-slate-400 text-sm mb-1">Client ID</label>
            <div class="flex items-center gap-2">
              <code class="flex-1 px-3 py-2 bg-slate-900 rounded text-cyan-400 font-mono text-sm">
                {{ selectedServicePrincipal.client_id }}
              </code>
              <button
                @click="copyToClipboard(selectedServicePrincipal.client_id, 'Client ID')"
                class="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-cyan-400 rounded transition-colors"
              >
                <i class="fas fa-copy"></i>
              </button>
            </div>
          </div>

          <div v-if="selectedServicePrincipal.description">
            <label class="block text-slate-400 text-sm mb-1">Description</label>
            <div class="text-slate-300">{{ selectedServicePrincipal.description }}</div>
          </div>

          <div>
            <label class="block text-slate-400 text-sm mb-1">Status</label>
            <span
              :class="[
                'inline-block px-3 py-1 rounded-full text-sm font-medium',
                selectedServicePrincipal.is_active
                  ? 'bg-green-500/20 text-green-400'
                  : 'bg-red-500/20 text-red-400'
              ]"
            >
              {{ selectedServicePrincipal.is_active ? 'Active' : 'Inactive' }}
            </span>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-slate-400 text-sm mb-1">Created</label>
              <div class="text-slate-300">{{ formatDate(selectedServicePrincipal.created_at) }}</div>
            </div>
            <div>
              <label class="block text-slate-400 text-sm mb-1">Last Used</label>
              <div class="text-slate-300">
                {{ selectedServicePrincipal.last_used_at ? formatDate(selectedServicePrincipal.last_used_at) : 'Never' }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Created Success Modal -->
    <div
      v-if="showCreatedModal && newServicePrincipal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      @click.self="() => {}"
    >
      <div class="bg-slate-800 rounded-lg max-w-2xl w-full p-6">
        <div class="flex items-center gap-3 mb-6">
          <i class="fas fa-check-circle text-green-400 text-3xl"></i>
          <h2 class="text-2xl font-bold text-green-400">Service Principal Created!</h2>
        </div>

        <div class="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-4 mb-6">
          <div class="flex items-start gap-3">
            <i class="fas fa-exclamation-triangle text-yellow-400 mt-1"></i>
            <div class="text-sm text-yellow-200">
              <strong>Important:</strong> Save the client secret below. For security reasons, it won't be shown again.
            </div>
          </div>
        </div>

        <div class="space-y-4">
          <div>
            <label class="block text-slate-400 text-sm mb-1">Client ID</label>
            <div class="flex items-center gap-2">
              <code class="flex-1 px-3 py-2 bg-slate-900 rounded text-cyan-400 font-mono text-sm">
                {{ newServicePrincipal.client_id }}
              </code>
              <button
                @click="copyToClipboard(newServicePrincipal.client_id, 'Client ID')"
                class="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-cyan-400 rounded transition-colors"
              >
                <i class="fas fa-copy"></i>
              </button>
            </div>
          </div>

          <div>
            <label class="block text-slate-400 text-sm mb-1">Client Secret</label>
            <div class="flex items-center gap-2">
              <code class="flex-1 px-3 py-2 bg-slate-900 rounded text-green-400 font-mono text-sm">
                {{ newServicePrincipal.client_secret }}
              </code>
              <button
                @click="copyToClipboard(newServicePrincipal.client_secret, 'Client Secret')"
                class="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-green-400 rounded transition-colors"
              >
                <i class="fas fa-copy"></i>
              </button>
            </div>
          </div>
        </div>

        <div class="mt-6 flex justify-end">
          <button
            @click="closeCreatedModal"
            class="px-6 py-2 bg-cyan-600 hover:bg-cyan-700 text-white rounded-lg transition-colors"
          >
            I've Saved the Secret
          </button>
        </div>
      </div>
    </div>

    <!-- Regenerate Confirmation Modal -->
    <div
      v-if="showRegenerateModal && selectedServicePrincipal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      @click.self="showRegenerateModal = false"
    >
      <div class="bg-slate-800 rounded-lg max-w-md w-full p-6">
        <h2 class="text-2xl font-bold text-yellow-400 mb-4 flex items-center gap-2">
          <i class="fas fa-exclamation-triangle"></i>
          Regenerate Secret?
        </h2>
        <p class="text-slate-300 mb-6">
          This will generate a new client secret for <strong>{{ selectedServicePrincipal.name }}</strong>.
          The old secret will immediately stop working. This action cannot be undone.
        </p>
        <div class="flex justify-end gap-2">
          <button
            @click="showRegenerateModal = false"
            class="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition-colors"
          >
            Cancel
          </button>
          <button
            @click="handleRegenerate"
            :disabled="regenerateLoading"
            class="px-4 py-2 bg-yellow-600 hover:bg-yellow-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            <i v-if="regenerateLoading" class="fas fa-spinner fa-spin"></i>
            Regenerate Secret
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div
      v-if="showDeleteModal && selectedServicePrincipal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      @click.self="showDeleteModal = false"
    >
      <div class="bg-slate-800 rounded-lg max-w-md w-full p-6">
        <h2 class="text-2xl font-bold text-red-400 mb-4 flex items-center gap-2">
          <i class="fas fa-exclamation-triangle"></i>
          Delete Service Principal?
        </h2>
        <p class="text-slate-300 mb-6">
          Are you sure you want to delete <strong>{{ selectedServicePrincipal.name }}</strong>?
          This will immediately revoke all access and cannot be undone.
        </p>
        <div class="flex justify-end gap-2">
          <button
            @click="showDeleteModal = false"
            class="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition-colors"
          >
            Cancel
          </button>
          <button
            @click="handleDelete"
            :disabled="deleteLoading"
            class="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            <i v-if="deleteLoading" class="fas fa-spinner fa-spin"></i>
            Delete
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { servicePrincipalApi } from '../api/servicePrincipals'
import { useToastStore } from '../stores/toast'
import type { ServicePrincipal, CreateServicePrincipalRequest } from '../types'

const toastStore = useToastStore()

const servicePrincipals = ref<ServicePrincipal[]>([])
const loading = ref(false)
const createLoading = ref(false)
const regenerateLoading = ref(false)
const deleteLoading = ref(false)

const showCreateModal = ref(false)
const showViewModal = ref(false)
const showCreatedModal = ref(false)
const showRegenerateModal = ref(false)
const showDeleteModal = ref(false)

const selectedServicePrincipal = ref<ServicePrincipal | null>(null)
const newServicePrincipal = ref<ServicePrincipal & { client_secret: string } | null>(null)

const createForm = ref<CreateServicePrincipalRequest>({
  name: '',
  description: ''
})

const fetchServicePrincipals = async () => {
  loading.value = true
  try {
    servicePrincipals.value = await servicePrincipalApi.list()
  } catch (error: any) {
    toastStore.addToast({
      type: 'error',
      message: error.response?.data?.error || 'Failed to fetch service principals'
    })
  } finally {
    loading.value = false
  }
}

const handleCreate = async () => {
  createLoading.value = true
  try {
    const response = await servicePrincipalApi.create(createForm.value)
    newServicePrincipal.value = response
    showCreateModal.value = false
    showCreatedModal.value = true
    createForm.value = { name: '', description: '' }
    await fetchServicePrincipals()
  } catch (error: any) {
    toastStore.addToast({
      type: 'error',
      message: error.response?.data?.error || 'Failed to create service principal'
    })
  } finally {
    createLoading.value = false
  }
}

const closeCreatedModal = () => {
  showCreatedModal.value = false
  newServicePrincipal.value = null
}

const viewServicePrincipal = async (sp: ServicePrincipal) => {
  try {
    selectedServicePrincipal.value = await servicePrincipalApi.get(sp.id)
    showViewModal.value = true
  } catch (error: any) {
    toastStore.addToast({
      type: 'error',
      message: error.response?.data?.error || 'Failed to fetch service principal details'
    })
  }
}

const confirmRegenerate = (sp: ServicePrincipal) => {
  selectedServicePrincipal.value = sp
  showRegenerateModal.value = true
}

const handleRegenerate = async () => {
  if (!selectedServicePrincipal.value) return

  regenerateLoading.value = true
  try {
    const response = await servicePrincipalApi.regenerate(selectedServicePrincipal.value.id)
    showRegenerateModal.value = false
    newServicePrincipal.value = response
    showCreatedModal.value = true
    await fetchServicePrincipals()
  } catch (error: any) {
    toastStore.addToast({
      type: 'error',
      message: error.response?.data?.error || 'Failed to regenerate client secret'
    })
  } finally {
    regenerateLoading.value = false
  }
}

const confirmDelete = (sp: ServicePrincipal) => {
  selectedServicePrincipal.value = sp
  showDeleteModal.value = true
}

const handleDelete = async () => {
  if (!selectedServicePrincipal.value) return

  deleteLoading.value = true
  try {
    await servicePrincipalApi.delete(selectedServicePrincipal.value.id)
    showDeleteModal.value = false
    toastStore.addToast({
      type: 'success',
      message: 'Service principal deleted successfully'
    })
    await fetchServicePrincipals()
  } catch (error: any) {
    toastStore.addToast({
      type: 'error',
      message: error.response?.data?.error || 'Failed to delete service principal'
    })
  } finally {
    deleteLoading.value = false
  }
}

const copyToClipboard = async (text: string, label: string) => {
  try {
    await navigator.clipboard.writeText(text)
    toastStore.addToast({
      type: 'success',
      message: `${label} copied to clipboard`
    })
  } catch {
    toastStore.addToast({
      type: 'error',
      message: 'Failed to copy to clipboard'
    })
  }
}

const formatDate = (dateString: string): string => {
  return new Date(dateString).toLocaleString()
}

onMounted(() => {
  fetchServicePrincipals()
})
</script>

<style scoped>
.service-principals-view {
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
