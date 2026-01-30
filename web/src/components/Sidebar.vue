<template>
  <aside class="w-64 bg-dark-card border-r border-gray-800 flex flex-col">
    <!-- Logo -->
    <div class="flex items-center gap-3 px-6 py-4 border-b border-gray-800">
      <div class="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
        <i class="fas fa-lock text-white text-sm"></i>
      </div>
      <span class="text-lg font-semibold text-gray-300">SecureVault</span>
    </div>

    <!-- Navigation -->
    <nav class="px-4 py-4 space-y-1 flex-1 overflow-y-auto">
      <RouterLink
        v-for="item in menuItems"
        :key="item.name"
        :to="item.route"
        v-slot="{ isActive }"
        custom
      >
        <button
          @click="$router.push(item.route)"
          :class="[
            'flex items-center gap-3 w-full px-4 py-3 text-left rounded-lg transition-all duration-200',
            isActive
              ? 'bg-primary/20 text-primary border-l-2 border-primary'
              : 'text-gray-400 hover:bg-gray-800 hover:text-gray-300'
          ]"
        >
          <i :class="[item.icon, 'w-5']"></i>
          <span class="font-medium">{{ item.name }}</span>
        </button>
      </RouterLink>
    </nav>

    <!-- Footer -->
    <div class="px-6 py-4 border-t border-gray-800 space-y-3">
      <!-- User Info -->
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-blue-600 flex items-center justify-center">
          <i class="fas fa-user text-white"></i>
        </div>
        <div class="flex-1 min-w-0">
          <div class="text-sm font-medium text-gray-300 truncate">
            {{ authStore.user?.email || 'User' }}
          </div>
          <div class="text-xs text-gray-500">Administrator</div>
        </div>
        <button
          @click="handleLogout"
          class="text-gray-400 hover:text-red-400 transition-colors"
          title="Logout"
        >
          <i class="fas fa-sign-out-alt"></i>
        </button>
      </div>

      <!-- Version -->
      <div class="text-xs text-gray-500 text-center">
        v{{ version }} - Enterprise
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'

const router = useRouter()
const authStore = useAuthStore()
const toastStore = useToastStore()

const version = import.meta.env.VITE_VERSION || '2.1.0'

interface MenuItem {
  name: string
  route: string
  icon: string
}

const menuItems: MenuItem[] = [
  { name: 'Dashboard', route: '/', icon: 'fas fa-chart-line' },
  { name: 'Secrets', route: '/secrets', icon: 'fas fa-key' },
  { name: 'Service Principals', route: '/service-principals', icon: 'fas fa-users-cog' },
  { name: 'Key Rotation', route: '/key-rotation', icon: 'fas fa-sync-alt' },
  { name: 'Audit Logs', route: '/audit-logs', icon: 'fas fa-history' },
  { name: 'Settings', route: '/settings', icon: 'fas fa-cog' }
]

async function handleLogout() {
  await authStore.logout()
  toastStore.info('Logged out successfully')
  router.push('/login')
}
</script>
