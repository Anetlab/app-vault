<template>
  <div class="p-6">
    <!-- Category Filters -->
    <div class="flex items-center gap-3 mb-6">
      <button
        v-for="category in secretsStore.categories"
        :key="category"
        @click="secretsStore.setCategory(category)"
        :class="[
          'px-6 py-2.5 rounded-xl font-medium transition-all duration-200',
          secretsStore.selectedCategory === category
            ? 'bg-dark-border text-white shadow-lg'
            : 'bg-dark-card text-gray-400 hover:bg-dark-border hover:text-white border border-gray-700'
        ]"
      >
        {{ category }}
      </button>
    </div>

    <!-- Action Buttons -->
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <button
          @click="showAddModal = true"
          class="btn-primary flex items-center gap-2"
        >
          <i class="fas fa-plus"></i>
          Add Secret
        </button>

        <button class="btn-secondary flex items-center gap-2">
          <i class="fas fa-upload"></i>
          Import
        </button>

        <button class="btn-secondary flex items-center gap-2 text-primary">
          <i class="fas fa-download"></i>
          Export
        </button>
      </div>

      <div class="flex items-center gap-3">
        <button class="btn-secondary flex items-center gap-2">
          <i class="fas fa-file-import"></i>
          Bulk Import
        </button>

        <button class="btn-primary flex items-center gap-2">
          <i class="fas fa-file-export"></i>
          Backup
        </button>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-4 gap-4 mb-6">
      <div class="bg-dark-card rounded-xl p-4 border border-gray-800">
        <div class="text-gray-400 text-sm mb-1">Total Secrets</div>
        <div class="text-2xl font-bold text-white">{{ secretsStore.secrets.length }}</div>
      </div>
      <div class="bg-dark-card rounded-xl p-4 border border-gray-800">
        <div class="text-gray-400 text-sm mb-1">Database</div>
        <div class="text-2xl font-bold text-cyan-400">{{ secretsStore.secretsByType['Database'] || 0 }}</div>
      </div>
      <div class="bg-dark-card rounded-xl p-4 border border-gray-800">
        <div class="text-gray-400 text-sm mb-1">API Keys</div>
        <div class="text-2xl font-bold text-purple-400">{{ secretsStore.secretsByType['API Key'] || 0 }}</div>
      </div>
      <div class="bg-dark-card rounded-xl p-4 border border-gray-800">
        <div class="text-gray-400 text-sm mb-1">Other</div>
        <div class="text-2xl font-bold text-green-400">{{ secretsStore.secretsByType['Generic'] || 0 }}</div>
      </div>
    </div>

    <!-- Secrets Table -->
    <div class="bg-dark-card rounded-2xl border border-gray-800 overflow-hidden">
      <!-- Table Header -->
      <div class="grid grid-cols-4 px-6 py-4 border-b border-gray-800 text-sm font-medium text-gray-400">
        <div>Name</div>
        <div class="flex flex-col">
          <span>Type</span>
          <div class="w-6 h-0.5 bg-primary mt-1"></div>
        </div>
        <div>Created Date</div>
        <div class="text-right">Actions</div>
      </div>

      <!-- Table Body -->
      <TransitionGroup name="slide" tag="div" class="divide-y divide-gray-800">
        <div
          v-for="secret in secretsStore.filteredSecrets"
          :key="secret.id"
          @click="secretsStore.selectSecret(secret.id)"
          :class="[
            'grid grid-cols-4 px-6 py-4 items-center cursor-pointer transition-all duration-200',
            secretsStore.selectedSecretId === secret.id
              ? 'bg-primary text-white'
              : 'hover:bg-dark-border text-gray-300'
          ]"
        >
          <div class="flex items-center gap-3">
            <i :class="['fas fa-lock', secretsStore.selectedSecretId === secret.id ? 'text-white' : 'text-gray-400']"></i>
            <span class="font-medium truncate">{{ secret.name }}</span>
          </div>

          <div :class="secretsStore.selectedSecretId === secret.id ? 'text-white/90' : 'text-gray-400'">
            {{ secret.type }}
          </div>

          <div :class="secretsStore.selectedSecretId === secret.id ? 'text-white/90' : 'text-gray-400'">
            {{ formatDate(secret.createdAt) }}
          </div>

          <div class="flex items-center justify-end gap-2">
            <button
              @click.stop="handleCopy(secret)"
              :class="[
                'p-2 rounded-lg transition-all duration-200',
                secretsStore.selectedSecretId === secret.id
                  ? 'hover:bg-white/20 text-white'
                  : 'hover:bg-gray-700 text-gray-400'
              ]"
              title="Copy"
            >
              <i class="fas fa-copy"></i>
            </button>
            <button
              @click.stop="handleView(secret)"
              :class="[
                'p-2 rounded-lg transition-all duration-200',
                secretsStore.selectedSecretId === secret.id
                  ? 'hover:bg-white/20 text-white'
                  : 'hover:bg-gray-700 text-gray-400'
              ]"
              title="View"
            >
              <i class="fas fa-eye"></i>
            </button>
            <button
              @click.stop="handleEdit(secret)"
              :class="[
                'p-2 rounded-lg transition-all duration-200',
                secretsStore.selectedSecretId === secret.id
                  ? 'hover:bg-white/20 text-white'
                  : 'hover:bg-gray-700 text-gray-400'
              ]"
              title="Edit"
            >
              <i class="fas fa-edit"></i>
            </button>
            <button
              @click.stop="handleDelete(secret)"
              :class="[
                'p-2 rounded-lg transition-all duration-200',
                secretsStore.selectedSecretId === secret.id
                  ? 'hover:bg-white/20 text-white'
                  : 'hover:bg-gray-700 text-gray-400 hover:text-red-400'
              ]"
              title="Delete"
            >
              <i class="fas fa-trash"></i>
            </button>
          </div>
        </div>
      </TransitionGroup>

      <!-- Empty State -->
      <div v-if="secretsStore.filteredSecrets.length === 0" class="text-center py-12">
        <div class="w-16 h-16 bg-gray-800 rounded-full flex items-center justify-center mx-auto mb-4">
          <i class="fas fa-search text-gray-500 text-2xl"></i>
        </div>
        <p class="text-gray-400">No secrets found</p>
      </div>

      <!-- Footer -->
      <div class="px-6 py-4 border-t border-gray-800 text-sm text-gray-500 flex items-center justify-between">
        <span>Showing {{ secretsStore.filteredSecrets.length }} of {{ secretsStore.secrets.length }} secrets</span>
        <span>Last updated: {{ new Date().toLocaleTimeString() }}</span>
      </div>
    </div>

    <!-- Add/Edit Secret Modal -->
    <AddSecretModal
      v-model:show="showAddModal"
      :secret="editingSecret"
      @saved="handleSecretSaved"
    />

    <!-- View Secret Modal -->
    <ViewSecretModal
      v-model:show="showViewModal"
      :secret="viewingSecret"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSecretsStore } from '@/stores/secrets'
import { useToastStore } from '@/stores/toast'
import type { Secret } from '@/types'
import AddSecretModal from '@/components/AddSecretModal.vue'
import ViewSecretModal from '@/components/ViewSecretModal.vue'

const secretsStore = useSecretsStore()
const toastStore = useToastStore()

const showAddModal = ref(false)
const showViewModal = ref(false)
const editingSecret = ref<Secret | null>(null)
const viewingSecret = ref<Secret | null>(null)

onMounted(async () => {
  await secretsStore.fetchSecrets()
})

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString()
}

async function handleCopy(secret: Secret) {
  try {
    const secretWithValue = await secretsStore.getSecretValue(secret.id)
    if (secretWithValue) {
      await navigator.clipboard.writeText(secretWithValue.value)
      toastStore.success('Secret copied to clipboard')
    }
  } catch (error) {
    toastStore.error('Failed to copy secret')
  }
}

function handleView(secret: Secret) {
  viewingSecret.value = secret
  showViewModal.value = true
}

function handleEdit(secret: Secret) {
  editingSecret.value = secret
  showAddModal.value = true
}

async function handleDelete(secret: Secret) {
  if (confirm(`Are you sure you want to delete "${secret.name}"?`)) {
    const result = await secretsStore.deleteSecret(secret.id)
    if (result.success) {
      toastStore.success('Secret deleted successfully')
    } else {
      toastStore.error(result.error || 'Failed to delete secret')
    }
  }
}

function handleSecretSaved() {
  editingSecret.value = null
  secretsStore.fetchSecrets()
}
</script>
