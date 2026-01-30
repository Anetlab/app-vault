<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="show"
        class="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4"
        @click.self="close"
      >
        <div class="bg-dark-card rounded-2xl p-6 w-full max-w-2xl border border-gray-700 shadow-2xl">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-bold text-white">Secret Details</h2>
            <button
              @click="close"
              class="text-gray-400 hover:text-white transition-colors"
            >
              <i class="fas fa-times text-xl"></i>
            </button>
          </div>

          <div v-if="loading" class="flex justify-center py-12">
            <i class="fas fa-spinner fa-spin text-3xl text-primary"></i>
          </div>

          <div v-else-if="secretValue" class="space-y-6">
            <!-- Secret Name -->
            <div>
              <label class="block text-sm text-gray-400 mb-2">Name</label>
              <div class="flex items-center gap-3 p-3 bg-dark-bg rounded-lg">
                <i class="fas fa-lock text-primary"></i>
                <span class="font-medium text-white">{{ secretValue.name }}</span>
              </div>
            </div>

            <!-- Type and Tags -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm text-gray-400 mb-2">Type</label>
                <div class="p-3 bg-dark-bg rounded-lg text-white">
                  {{ secretValue.type }}
                </div>
              </div>
              <div>
                <label class="block text-sm text-gray-400 mb-2">Version</label>
                <div class="p-3 bg-dark-bg rounded-lg text-white">
                  v{{ secretValue.version }}
                </div>
              </div>
            </div>

            <!-- Tags -->
            <div v-if="secretValue.tags.length > 0">
              <label class="block text-sm text-gray-400 mb-2">Tags</label>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="tag in secretValue.tags"
                  :key="tag"
                  class="px-3 py-1 bg-primary/20 text-primary rounded-full text-sm"
                >
                  {{ tag }}
                </span>
              </div>
            </div>

            <!-- Public Key (if available) -->
            <div v-if="secretValue.publicKey">
              <label class="block text-sm text-gray-400 mb-2">Public Key (for other projects)</label>
              <div class="relative p-4 bg-dark-bg rounded-lg border border-gray-700">
                <div class="flex items-center justify-between gap-2 mb-2">
                  <span class="text-xs text-gray-500">Base64 encoded</span>
                  <button
                    @click="copyPublicKey"
                    class="text-sm text-primary hover:text-primary-hover flex items-center gap-2"
                  >
                    <i class="fas fa-copy"></i>
                    {{ publicKeyCopied ? 'Copied!' : 'Copy' }}
                  </button>
                </div>
                <pre
                  class="font-mono text-sm text-white whitespace-pre-wrap break-all max-h-32 overflow-y-auto"
                >{{ secretValue.publicKey }}</pre>
              </div>
            </div>

            <!-- Secret Value -->
            <div>
              <div class="flex items-center justify-between mb-2">
                <label class="block text-sm text-gray-400">Secret Value</label>
                <div class="flex items-center gap-2">
                  <button
                    @click="toggleVisibility"
                    class="text-sm text-primary hover:text-primary-hover flex items-center gap-2"
                  >
                    <i :class="showValue ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
                    {{ showValue ? 'Hide' : 'Show' }}
                  </button>
                  <button
                    @click="copyValue"
                    class="text-sm text-primary hover:text-primary-hover flex items-center gap-2"
                  >
                    <i class="fas fa-copy"></i>
                    {{ copied ? 'Copied!' : 'Copy' }}
                  </button>
                </div>
              </div>
              <div class="relative p-4 bg-dark-bg rounded-lg border border-gray-700">
                <pre
                  v-if="showValue"
                  class="font-mono text-sm text-white whitespace-pre-wrap break-all"
                >{{ secretValue.value }}</pre>
                <div v-else class="flex items-center gap-2 text-gray-500">
                  <i class="fas fa-eye-slash"></i>
                  <span>Click "Show" to reveal the secret value</span>
                </div>
              </div>
            </div>

            <!-- Metadata -->
            <div class="grid grid-cols-2 gap-4 pt-4 border-t border-gray-800">
              <div>
                <label class="block text-sm text-gray-400 mb-1">Created</label>
                <div class="text-sm text-white">{{ formatDateTime(secretValue.createdAt) }}</div>
              </div>
              <div>
                <label class="block text-sm text-gray-400 mb-1">Last Updated</label>
                <div class="text-sm text-white">{{ formatDateTime(secretValue.updatedAt) }}</div>
              </div>
              <div>
                <label class="block text-sm text-gray-400 mb-1">Access Count</label>
                <div class="text-sm text-white">{{ secretValue.accessCount }} times</div>
              </div>
              <div v-if="secretValue.expiresAt">
                <label class="block text-sm text-gray-400 mb-1">Expires</label>
                <div class="text-sm text-yellow-400">{{ formatDateTime(secretValue.expiresAt) }}</div>
              </div>
            </div>
          </div>

          <div v-else class="text-center py-12 text-gray-400">
            Failed to load secret
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useSecretsStore } from '@/stores/secrets'
import { useToastStore } from '@/stores/toast'
import type { Secret, SecretWithValue } from '@/types'

interface Props {
  show: boolean
  secret: Secret | null
}

const props = defineProps<Props>()
const emit = defineEmits(['update:show'])

const secretsStore = useSecretsStore()
const toastStore = useToastStore()

const secretValue = ref<SecretWithValue | null>(null)
const loading = ref(false)
const showValue = ref(false)
const copied = ref(false)
const publicKeyCopied = ref(false)

watch(() => props.show, async (newVal) => {
  if (newVal && props.secret) {
    await loadSecret()
  } else {
    showValue.value = false
    copied.value = false
    publicKeyCopied.value = false
  }
})

async function loadSecret() {
  if (!props.secret) return
  
  loading.value = true
  try {
    secretValue.value = await secretsStore.getSecretValue(props.secret.id)
  } catch (error) {
    toastStore.error('Failed to load secret')
  } finally {
    loading.value = false
  }
}

function toggleVisibility() {
  showValue.value = !showValue.value
}

async function copyValue() {
  if (!secretValue.value) return
  
  try {
    await navigator.clipboard.writeText(secretValue.value.value)
    copied.value = true
    toastStore.success('Secret copied to clipboard')
    
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (error) {
    toastStore.error('Failed to copy to clipboard')
  }
}

async function copyPublicKey() {
  if (!secretValue.value?.publicKey) return
  
  try {
    await navigator.clipboard.writeText(secretValue.value.publicKey)
    publicKeyCopied.value = true
    toastStore.success('Public key copied to clipboard')
    
    setTimeout(() => {
      publicKeyCopied.value = false
    }, 2000)
  } catch (error) {
    toastStore.error('Failed to copy public key')
  }
}

function formatDateTime(dateString: string): string {
  return new Date(dateString).toLocaleString()
}

function close() {
  emit('update:show', false)
}
</script>
