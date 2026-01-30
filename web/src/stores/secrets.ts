import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { secretsApi } from '@/api/secrets'
import type { Secret, SecretWithValue, CreateSecretRequest } from '@/types'

export const useSecretsStore = defineStore('secrets', () => {
  // State
  const secrets = ref<Secret[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const searchQuery = ref('')
  const selectedCategory = ref<string>('All')
  const selectedSecretId = ref<string | null>(null)

  // Computed
  const filteredSecrets = computed(() => {
    let filtered = secrets.value

    // Filter by category
    if (selectedCategory.value !== 'All') {
      filtered = filtered.filter(s => s.type === selectedCategory.value)
    }

    // Filter by search query
    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase()
      filtered = filtered.filter(s =>
        s.name.toLowerCase().includes(query) ||
        s.tags.some(tag => tag.toLowerCase().includes(query))
      )
    }

    return filtered
  })

  const categories = computed(() => {
    const types = new Set(secrets.value.map(s => s.type))
    return ['All', ...Array.from(types)]
  })

  const secretsByType = computed(() => {
    const byType: Record<string, number> = {}
    secrets.value.forEach(secret => {
      byType[secret.type] = (byType[secret.type] || 0) + 1
    })
    return byType
  })

  // Actions
  async function fetchSecrets() {
    loading.value = true
    error.value = null
    
    try {
      secrets.value = await secretsApi.list()
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to fetch secrets'
      console.error('Fetch secrets error:', err)
    } finally {
      loading.value = false
    }
  }

  async function createSecret(data: CreateSecretRequest) {
    loading.value = true
    error.value = null
    
    try {
      const newSecret = await secretsApi.create(data)
      secrets.value.unshift(newSecret)
      return { success: true, secret: newSecret }
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to create secret'
      console.error('Create secret error:', err)
      return { success: false, error: error.value }
    } finally {
      loading.value = false
    }
  }

  async function getSecretValue(id: string): Promise<SecretWithValue | null> {
    try {
      return await secretsApi.getById(id)
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to get secret value'
      console.error('Get secret value error:', err)
      return null
    }
  }

  async function updateSecret(id: string, value: string, version: number) {
    loading.value = true
    error.value = null
    
    try {
      const updated = await secretsApi.update(id, { value, version })
      const index = secrets.value.findIndex(s => s.id === id)
      if (index !== -1) {
        secrets.value[index] = updated
      }
      return { success: true, secret: updated }
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to update secret'
      console.error('Update secret error:', err)
      return { success: false, error: error.value }
    } finally {
      loading.value = false
    }
  }

  async function deleteSecret(id: string) {
    loading.value = true
    error.value = null
    
    try {
      await secretsApi.delete(id)
      secrets.value = secrets.value.filter(s => s.id !== id)
      if (selectedSecretId.value === id) {
        selectedSecretId.value = null
      }
      return { success: true }
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Failed to delete secret'
      console.error('Delete secret error:', err)
      return { success: false, error: error.value }
    } finally {
      loading.value = false
    }
  }

  function setSearchQuery(query: string) {
    searchQuery.value = query
  }

  function setCategory(category: string) {
    selectedCategory.value = category
  }

  function selectSecret(id: string | null) {
    selectedSecretId.value = id
  }

  return {
    // State
    secrets,
    loading,
    error,
    searchQuery,
    selectedCategory,
    selectedSecretId,
    
    // Computed
    filteredSecrets,
    categories,
    secretsByType,
    
    // Actions
    fetchSecrets,
    createSecret,
    getSecretValue,
    updateSecret,
    deleteSecret,
    setSearchQuery,
    setCategory,
    selectSecret
  }
})
