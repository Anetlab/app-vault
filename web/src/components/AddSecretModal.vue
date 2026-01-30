<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="show"
        class="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4"
        @click.self="close"
      >
        <div class="bg-dark-card rounded-2xl p-6 w-full max-w-lg border border-gray-700 shadow-2xl">
          <h2 class="text-xl font-bold mb-4 text-white">
            {{ secret ? 'Edit Secret' : 'Add New Secret' }}
          </h2>
          
          <form @submit.prevent="handleSubmit" class="space-y-4">
            <!-- Name -->
            <div>
              <label class="block text-sm text-gray-400 mb-2">Secret Name</label>
              <input
                v-model="form.name"
                type="text"
                required
                class="input-field"
                placeholder="e.g., database-password"
                :disabled="!!secret"
              />
            </div>

            <!-- Type -->
            <div>
              <label class="block text-sm text-gray-400 mb-2">Type</label>
              <select
                v-model="form.type"
                required
                class="input-field"
              >
                <option value="Generic">Generic</option>
                <option value="Password">Password</option>
                <option value="API Key">API Key</option>
                <option value="Certificate">Certificate</option>
                <option value="Connection String">Connection String</option>
                <option value="Database">Database</option>
              </select>
            </div>

            <!-- Value -->
            <div>
              <label class="block text-sm text-gray-400 mb-2">Secret Value</label>
              <div class="relative">
                <textarea
                  v-model="form.value"
                  required
                  rows="4"
                  class="input-field pr-10 font-mono text-sm"
                  :type="showValue ? 'text' : 'password'"
                  placeholder="Enter secret value..."
                ></textarea>
                <button
                  type="button"
                  @click="showValue = !showValue"
                  class="absolute right-3 top-3 text-gray-400 hover:text-gray-300"
                >
                  <i :class="showValue ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
                </button>
              </div>
            </div>

            <!-- Tags -->
            <div>
              <label class="block text-sm text-gray-400 mb-2">Tags (comma separated)</label>
              <input
                v-model="form.tagsString"
                type="text"
                class="input-field"
                placeholder="production, database, postgres"
              />
            </div>

            <!-- Expiration (optional) -->
            <div>
              <label class="block text-sm text-gray-400 mb-2">
                Expiration Date (Optional)
              </label>
              <input
                v-model="form.expiresAt"
                type="datetime-local"
                class="input-field"
              />
            </div>

            <!-- Error Message -->
            <div v-if="errorMessage" class="p-3 bg-red-500/20 border border-red-500/50 rounded-lg text-red-400 text-sm">
              <i class="fas fa-exclamation-triangle mr-2"></i>
              {{ errorMessage }}
            </div>

            <!-- Buttons -->
            <div class="flex gap-3 mt-6">
              <button
                type="button"
                @click="close"
                class="flex-1 btn-secondary"
              >
                Cancel
              </button>
              <button
                type="submit"
                :disabled="loading"
                class="flex-1 btn-primary"
              >
                <i v-if="loading" class="fas fa-spinner fa-spin mr-2"></i>
                {{ secret ? 'Update' : 'Add Secret' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useSecretsStore } from '@/stores/secrets'
import { useToastStore } from '@/stores/toast'
import type { Secret } from '@/types'

interface Props {
  show: boolean
  secret?: Secret | null
}

const props = defineProps<Props>()
const emit = defineEmits(['update:show', 'saved'])

const secretsStore = useSecretsStore()
const toastStore = useToastStore()

const form = ref({
  name: '',
  type: 'Generic',
  value: '',
  tagsString: '',
  expiresAt: ''
})

const showValue = ref(false)
const loading = ref(false)
const errorMessage = ref('')

watch(() => props.show, (newVal) => {
  if (newVal) {
    if (props.secret) {
      // Editing existing secret - load current value
      loadSecretForEdit()
    } else {
      // Reset form for new secret
      form.value = {
        name: '',
        type: 'Generic',
        value: '',
        tagsString: '',
        expiresAt: ''
      }
    }
    errorMessage.value = ''
    showValue.value = false
  }
})

async function loadSecretForEdit() {
  if (!props.secret) return
  
  loading.value = true
  try {
    const secretWithValue = await secretsStore.getSecretValue(props.secret.id)
    if (secretWithValue) {
      form.value = {
        name: secretWithValue.name,
        type: secretWithValue.type,
        value: secretWithValue.value,
        tagsString: secretWithValue.tags.join(', '),
        expiresAt: secretWithValue.expiresAt ? new Date(secretWithValue.expiresAt).toISOString().slice(0, 16) : ''
      }
    }
  } catch (error) {
    toastStore.error('Failed to load secret value')
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  loading.value = true
  errorMessage.value = ''

  try {
    const tags = form.value.tagsString
      .split(',')
      .map(t => t.trim())
      .filter(t => t.length > 0)

    if (props.secret) {
      // Update existing secret
      const result = await secretsStore.updateSecret(
        props.secret.id,
        form.value.value,
        props.secret.version
      )
      
      if (result.success) {
        toastStore.success('Secret updated successfully')
        emit('saved')
        close()
      } else {
        errorMessage.value = result.error || 'Failed to update secret'
      }
    } else {
      // Create new secret
      const result = await secretsStore.createSecret({
        name: form.value.name,
        value: form.value.value,
        type: form.value.type,
        tags: tags,
        expiresAt: form.value.expiresAt || undefined
      })

      if (result.success) {
        toastStore.success('Secret created successfully')
        emit('saved')
        close()
      } else {
        errorMessage.value = result.error || 'Failed to create secret'
      }
    }
  } catch (error: any) {
    errorMessage.value = error.message || 'An unexpected error occurred'
  } finally {
    loading.value = false
  }
}

function close() {
  emit('update:show', false)
}
</script>
