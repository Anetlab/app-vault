<template>
  <div class="p-6 space-y-6">
    <!-- Page Header -->
    <div>
      <h1 class="text-3xl font-bold text-white mb-2">Dashboard</h1>
      <p class="text-gray-400">Overview of your secret management system</p>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <div class="card">
        <div class="flex items-center justify-between">
          <div>
            <div class="text-sm text-gray-400 mb-1">Total Secrets</div>
            <div class="text-3xl font-bold text-white">{{ stats.totalSecrets }}</div>
          </div>
          <div class="w-12 h-12 bg-primary/20 rounded-xl flex items-center justify-center">
            <i class="fas fa-key text-primary text-xl"></i>
          </div>
        </div>
        <div class="mt-4 text-xs text-gray-500">
          <i class="fas fa-arrow-up text-green-500 mr-1"></i>
          12% from last month
        </div>
      </div>

      <div class="card">
        <div class="flex items-center justify-between">
          <div>
            <div class="text-sm text-gray-400 mb-1">Service Principals</div>
            <div class="text-3xl font-bold text-cyan-400">{{ stats.servicePrincipals }}</div>
          </div>
          <div class="w-12 h-12 bg-cyan-500/20 rounded-xl flex items-center justify-center">
            <i class="fas fa-users-cog text-cyan-400 text-xl"></i>
          </div>
        </div>
        <div class="mt-4 text-xs text-gray-500">
          <i class="fas fa-check text-green-500 mr-1"></i>
          All active
        </div>
      </div>

      <div class="card">
        <div class="flex items-center justify-between">
          <div>
            <div class="text-sm text-gray-400 mb-1">Key Rotation Age</div>
            <div class="text-3xl font-bold text-purple-400">{{ stats.keyAge }} days</div>
          </div>
          <div class="w-12 h-12 bg-purple-500/20 rounded-xl flex items-center justify-center">
            <i class="fas fa-sync-alt text-purple-400 text-xl"></i>
          </div>
        </div>
        <div class="mt-4 text-xs text-gray-500">
          <i class="fas fa-info-circle text-blue-500 mr-1"></i>
          Next rotation in {{ 90 - stats.keyAge }} days
        </div>
      </div>

      <div class="card">
        <div class="flex items-center justify-between">
          <div>
            <div class="text-sm text-gray-400 mb-1">Audit Events</div>
            <div class="text-3xl font-bold text-green-400">{{ stats.auditEvents }}</div>
          </div>
          <div class="w-12 h-12 bg-green-500/20 rounded-xl flex items-center justify-center">
            <i class="fas fa-history text-green-400 text-xl"></i>
          </div>
        </div>
        <div class="mt-4 text-xs text-gray-500">
          <i class="fas fa-clock text-gray-500 mr-1"></i>
          Last 30 days
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="card">
      <h2 class="text-xl font-bold text-white mb-4">Quick Actions</h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <RouterLink
          to="/secrets"
          class="flex items-center gap-4 p-4 bg-dark-bg rounded-xl border border-gray-700 hover:border-primary transition-all group"
        >
          <div class="w-12 h-12 bg-primary/20 rounded-lg flex items-center justify-center group-hover:bg-primary/30 transition-colors">
            <i class="fas fa-plus text-primary text-xl"></i>
          </div>
          <div>
            <div class="font-semibold text-white mb-1">Add Secret</div>
            <div class="text-sm text-gray-400">Create a new encrypted secret</div>
          </div>
        </RouterLink>

        <RouterLink
          to="/service-principals"
          class="flex items-center gap-4 p-4 bg-dark-bg rounded-xl border border-gray-700 hover:border-cyan-500 transition-all group"
        >
          <div class="w-12 h-12 bg-cyan-500/20 rounded-lg flex items-center justify-center group-hover:bg-cyan-500/30 transition-colors">
            <i class="fas fa-user-plus text-cyan-400 text-xl"></i>
          </div>
          <div>
            <div class="font-semibold text-white mb-1">Create Service Principal</div>
            <div class="text-sm text-gray-400">Setup app authentication</div>
          </div>
        </RouterLink>

        <RouterLink
          to="/key-rotation"
          class="flex items-center gap-4 p-4 bg-dark-bg rounded-xl border border-gray-700 hover:border-purple-500 transition-all group"
        >
          <div class="w-12 h-12 bg-purple-500/20 rounded-lg flex items-center justify-center group-hover:bg-purple-500/30 transition-colors">
            <i class="fas fa-sync-alt text-purple-400 text-xl"></i>
          </div>
          <div>
            <div class="font-semibold text-white mb-1">Rotate Keys</div>
            <div class="text-sm text-gray-400">Trigger key rotation</div>
          </div>
        </RouterLink>
      </div>
    </div>

    <!-- Recent Activity -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Recent Secrets -->
      <div class="card">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-xl font-bold text-white">Recent Secrets</h2>
          <RouterLink to="/secrets" class="text-sm text-primary hover:text-primary-hover">
            View All <i class="fas fa-arrow-right ml-1"></i>
          </RouterLink>
        </div>
        <div class="space-y-3">
          <div
            v-for="secret in recentSecrets"
            :key="secret.id"
            class="flex items-center gap-3 p-3 bg-dark-bg rounded-lg"
          >
            <div class="w-10 h-10 bg-primary/20 rounded-lg flex items-center justify-center flex-shrink-0">
              <i class="fas fa-lock text-primary"></i>
            </div>
            <div class="flex-1 min-w-0">
              <div class="font-medium text-white truncate">{{ secret.name }}</div>
              <div class="text-sm text-gray-400">{{ secret.type }}</div>
            </div>
            <div class="text-xs text-gray-500">
              {{ formatDate(secret.createdAt) }}
            </div>
          </div>
        </div>
      </div>

      <!-- System Health -->
      <div class="card">
        <h2 class="text-xl font-bold text-white mb-4">System Health</h2>
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-green-500/20 rounded-lg flex items-center justify-center">
                <i class="fas fa-database text-green-400"></i>
              </div>
              <div>
                <div class="font-medium text-white">Database</div>
                <div class="text-sm text-gray-400">PostgreSQL</div>
              </div>
            </div>
            <span class="px-3 py-1 bg-green-500/20 text-green-400 rounded-full text-xs font-medium">
              Healthy
            </span>
          </div>

          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-green-500/20 rounded-lg flex items-center justify-center">
                <i class="fas fa-server text-green-400"></i>
              </div>
              <div>
                <div class="font-medium text-white">API Server</div>
                <div class="text-sm text-gray-400">Port 8888</div>
              </div>
            </div>
            <span class="px-3 py-1 bg-green-500/20 text-green-400 rounded-full text-xs font-medium">
              Online
            </span>
          </div>

          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-green-500/20 rounded-lg flex items-center justify-center">
                <i class="fas fa-shield-alt text-green-400"></i>
              </div>
              <div>
                <div class="font-medium text-white">Encryption</div>
                <div class="text-sm text-gray-400">XChaCha20-Poly1305</div>
              </div>
            </div>
            <span class="px-3 py-1 bg-green-500/20 text-green-400 rounded-full text-xs font-medium">
              Active
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useSecretsStore } from '@/stores/secrets'

const secretsStore = useSecretsStore()

const stats = ref({
  totalSecrets: 0,
  servicePrincipals: 0,
  keyAge: 45,
  auditEvents: 1247
})

const recentSecrets = ref<any[]>([])

onMounted(async () => {
  await secretsStore.fetchSecrets()
  stats.value.totalSecrets = secretsStore.secrets.length
  recentSecrets.value = secretsStore.secrets.slice(0, 5)
})

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  
  if (days === 0) return 'Today'
  if (days === 1) return 'Yesterday'
  if (days < 7) return `${days} days ago`
  return date.toLocaleDateString()
}
</script>
