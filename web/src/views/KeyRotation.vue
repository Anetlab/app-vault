<template>
  <div class="key-rotation-view">
    <h1 class="text-3xl font-bold text-cyan-400 mb-6">Key Rotation Management</h1>

    <div v-if="loading" class="text-center py-8">
      <i class="fas fa-spinner fa-spin text-4xl text-cyan-400"></i>
    </div>

    <div v-else class="space-y-6">
      <!-- Current Key Status -->
      <div class="bg-slate-800 rounded-lg p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-xl font-semibold text-white flex items-center gap-2">
            <i class="fas fa-key text-cyan-400"></i>
            Current Master Key
          </h2>
          <span
            :class="[
              'px-3 py-1 rounded-full text-sm font-medium',
              keyStatus?.is_active ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
            ]"
          >
            {{ keyStatus?.is_active ? 'Active' : 'Inactive' }}
          </span>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div>
            <label class="block text-slate-400 text-sm mb-1">Key Version</label>
            <div class="text-white text-2xl font-bold">{{ keyStatus?.current_version || 'N/A' }}</div>
          </div>
          <div>
            <label class="block text-slate-400 text-sm mb-1">Created</label>
            <div class="text-slate-300">{{ keyStatus?.created_at ? formatDate(keyStatus.created_at) : 'N/A' }}</div>
          </div>
          <div>
            <label class="block text-slate-400 text-sm mb-1">Last Rotated</label>
            <div class="text-slate-300">{{ keyStatus?.rotated_at ? formatDate(keyStatus.rotated_at) : 'Never' }}</div>
          </div>
        </div>

        <div class="mt-6 flex items-center justify-between p-4 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
          <div class="flex items-start gap-3">
            <i class="fas fa-exclamation-triangle text-yellow-400 mt-1"></i>
            <div>
              <p class="text-yellow-200 font-medium">Rotate Master Key Regularly</p>
              <p class="text-sm text-yellow-300/80 mt-1">
                It's recommended to rotate your master encryption key every 90 days for security best practices.
              </p>
            </div>
          </div>
          <button
            @click="showRotateModal = true"
            :disabled="rotateLoading"
            class="px-6 py-3 bg-cyan-600 hover:bg-cyan-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 whitespace-nowrap"
          >
            <i :class="['fas', rotateLoading ? 'fa-spinner fa-spin' : 'fa-sync-alt']"></i>
            Rotate Key
          </button>
        </div>
      </div>

      <!-- Key Statistics -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="bg-slate-800 rounded-lg p-6">
          <div class="flex items-center gap-3 mb-2">
            <i class="fas fa-database text-cyan-400 text-2xl"></i>
            <h3 class="text-slate-400">Secrets Encrypted</h3>
          </div>
          <p class="text-3xl font-bold text-white">{{ keyStatus?.encrypted_secrets || 0 }}</p>
        </div>

        <div class="bg-slate-800 rounded-lg p-6">
          <div class="flex items-center gap-3 mb-2">
            <i class="fas fa-history text-purple-400 text-2xl"></i>
            <h3 class="text-slate-400">Previous Versions</h3>
          </div>
          <p class="text-3xl font-bold text-white">{{ keyStatus?.total_versions ? keyStatus.total_versions - 1 : 0 }}</p>
        </div>

        <div class="bg-slate-800 rounded-lg p-6">
          <div class="flex items-center gap-3 mb-2">
            <i class="fas fa-clock text-yellow-400 text-2xl"></i>
            <h3 class="text-slate-400">Days Since Rotation</h3>
          </div>
          <p class="text-3xl font-bold text-white">{{ daysSinceRotation }}</p>
        </div>
      </div>

      <!-- Rotation History -->
      <div class="bg-slate-800 rounded-lg p-6">
        <h2 class="text-xl font-semibold text-white mb-4 flex items-center gap-2">
          <i class="fas fa-history text-cyan-400"></i>
          Rotation History
        </h2>

        <div v-if="rotationHistory.length === 0" class="text-center py-8 text-slate-400">
          <i class="fas fa-inbox text-4xl mb-3"></i>
          <p>No rotation history available</p>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="entry in rotationHistory"
            :key="entry.version"
            class="flex items-center justify-between p-4 bg-slate-900 rounded-lg"
          >
            <div class="flex items-center gap-4">
              <div class="w-12 h-12 bg-cyan-600/20 rounded-full flex items-center justify-center">
                <span class="text-cyan-400 font-bold">v{{ entry.version }}</span>
              </div>
              <div>
                <p class="text-white font-medium">Master Key Version {{ entry.version }}</p>
                <p class="text-sm text-slate-400">
                  Rotated {{ formatDate(entry.created_at) }}
                  <span v-if="entry.is_active" class="ml-2 px-2 py-0.5 bg-green-500/20 text-green-400 rounded text-xs">
                    Current
                  </span>
                </p>
              </div>
            </div>
            <div class="text-slate-400">
              <i class="fas fa-check-circle"></i>
            </div>
          </div>
        </div>
      </div>

      <!-- Security Best Practices -->
      <div class="bg-slate-800 rounded-lg p-6">
        <h2 class="text-xl font-semibold text-white mb-4 flex items-center gap-2">
          <i class="fas fa-shield-alt text-cyan-400"></i>
          Key Rotation Best Practices
        </h2>
        <ul class="space-y-3 text-slate-300">
          <li class="flex items-start gap-3">
            <i class="fas fa-check-circle text-green-400 mt-1"></i>
            <span>Rotate your master encryption key at least every 90 days</span>
          </li>
          <li class="flex items-start gap-3">
            <i class="fas fa-check-circle text-green-400 mt-1"></i>
            <span>Rotate immediately if you suspect key compromise</span>
          </li>
          <li class="flex items-start gap-3">
            <i class="fas fa-check-circle text-green-400 mt-1"></i>
            <span>Monitor audit logs for suspicious key usage patterns</span>
          </li>
          <li class="flex items-start gap-3">
            <i class="fas fa-check-circle text-green-400 mt-1"></i>
            <span>Keep backup of your secret key in a secure location</span>
          </li>
          <li class="flex items-start gap-3">
            <i class="fas fa-check-circle text-green-400 mt-1"></i>
            <span>Previous key versions are retained for decrypting existing secrets</span>
          </li>
        </ul>
      </div>
    </div>

    <!-- Rotate Confirmation Modal -->
    <div
      v-if="showRotateModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      @click.self="showRotateModal = false"
    >
      <div class="bg-slate-800 rounded-lg max-w-md w-full p-6">
        <h2 class="text-2xl font-bold text-yellow-400 mb-4 flex items-center gap-2">
          <i class="fas fa-exclamation-triangle"></i>
          Rotate Master Key?
        </h2>
        <div class="space-y-3 text-slate-300 mb-6">
          <p>
            This will create a new version of the master encryption key. All new secrets will be encrypted with the new key.
          </p>
          <p class="text-sm">
            <strong>Note:</strong> Existing secrets will remain encrypted with their current key versions and can still be decrypted.
          </p>
        </div>
        <div class="flex justify-end gap-2">
          <button
            @click="showRotateModal = false"
            class="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition-colors"
          >
            Cancel
          </button>
          <button
            @click="handleRotate"
            :disabled="rotateLoading"
            class="px-4 py-2 bg-cyan-600 hover:bg-cyan-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            <i v-if="rotateLoading" class="fas fa-spinner fa-spin"></i>
            Confirm Rotation
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { systemApi } from '../api/system'
import { useToastStore } from '../stores/toast'

const toastStore = useToastStore()

interface KeyStatus {
  current_version: number
  is_active: boolean
  created_at: string
  rotated_at?: string
  encrypted_secrets: number
  total_versions: number
}

interface RotationHistoryEntry {
  version: number
  created_at: string
  is_active: boolean
}

const keyStatus = ref<KeyStatus | null>(null)
const rotationHistory = ref<RotationHistoryEntry[]>([])
const loading = ref(false)
const rotateLoading = ref(false)
const showRotateModal = ref(false)

const daysSinceRotation = computed(() => {
  if (!keyStatus.value?.rotated_at && !keyStatus.value?.created_at) return 0
  const date = keyStatus.value.rotated_at || keyStatus.value.created_at
  const days = Math.floor((Date.now() - new Date(date).getTime()) / (1000 * 60 * 60 * 24))
  return days
})

const fetchKeyStatus = async () => {
  loading.value = true
  try {
    const response = await systemApi.getKeyStatus()
    keyStatus.value = response.status
    rotationHistory.value = response.history || []
  } catch (error: any) {
    toastStore.error(error.response?.data?.error || 'Failed to fetch key status')
  } finally {
    loading.value = false
  }
}

const handleRotate = async () => {
  rotateLoading.value = true
  try {
    await systemApi.rotateKey()
    showRotateModal.value = false
    toastStore.success('Master key rotated successfully')
    await fetchKeyStatus()
  } catch (error: any) {
    toastStore.error(error.response?.data?.error || 'Failed to rotate key')
  } finally {
    rotateLoading.value = false
  }
}

const formatDate = (dateString: string): string => {
  return new Date(dateString).toLocaleString()
}

onMounted(() => {
  fetchKeyStatus()
})
</script>

<style scoped>
.key-rotation-view {
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
