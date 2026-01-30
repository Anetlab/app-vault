<template>
  <header class="flex items-center justify-between px-6 py-4 border-b border-gray-800 bg-dark-bg">
    <!-- Search -->
    <div class="relative w-96">
      <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
        <i class="fas fa-search text-gray-500"></i>
      </div>
      <input
        v-model="searchQuery"
        @input="handleSearchInput"
        type="text"
        placeholder="Search secrets..."
        class="w-full pl-10 pr-4 py-2.5 bg-dark-card border border-gray-700 rounded-xl text-gray-300 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
      />
    </div>

    <!-- Right Section -->
    <div class="flex items-center gap-4">
      <!-- Notifications -->
      <button class="relative p-2 text-gray-400 hover:text-white transition-colors">
        <i class="fas fa-bell text-lg"></i>
        <span v-if="notificationCount > 0" class="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full animate-pulse"></span>
      </button>

      <!-- Messages -->
      <button class="relative p-2 text-gray-400 hover:text-white transition-colors">
        <i class="fas fa-envelope text-lg"></i>
      </button>

      <!-- System Status -->
      <div
        :class="[
          'flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium',
          systemStatus === 'healthy' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
        ]"
      >
        <span class="w-2 h-2 rounded-full bg-current animate-pulse"></span>
        {{ systemStatus === 'healthy' ? 'All Systems Operational' : 'System Issues' }}
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { systemApi } from '@/api/system'
import { useSecretsStore } from '@/stores/secrets'

const secretsStore = useSecretsStore()

const searchQuery = ref('')
const notificationCount = ref(0)
const systemStatus = ref<'healthy' | 'unhealthy'>('healthy')

let searchTimeout: ReturnType<typeof setTimeout>

function handleSearchInput() {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    secretsStore.setSearchQuery(searchQuery.value)
  }, 300)
}

async function checkSystemHealth() {
  try {
    const health = await systemApi.health()
    systemStatus.value = health.status === 'healthy' ? 'healthy' : 'unhealthy'
  } catch (error) {
    systemStatus.value = 'unhealthy'
  }
}

onMounted(() => {
  checkSystemHealth()
  // Check health every 30 seconds
  setInterval(checkSystemHealth, 30000)
})
</script>
