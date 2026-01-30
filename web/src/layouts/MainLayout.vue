<template>
  <div class="flex h-screen overflow-hidden">
    <!-- Sidebar -->
    <Sidebar :activeMenu="activeMenu" @menu-change="handleMenuChange" />

    <!-- Main Content -->
    <main class="flex-1 flex flex-col min-w-0 bg-dark-bg">
      <!-- Header -->
      <Header @search="handleSearch" />

      <!-- Content -->
      <div class="flex-1 overflow-auto">
        <RouterView v-slot="{ Component }">
          <Transition name="fade" mode="out-in">
            <component :is="Component" :key="$route.path" />
          </Transition>
        </RouterView>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute, RouterView } from 'vue-router'
import Sidebar from '@/components/Sidebar.vue'
import Header from '@/components/Header.vue'

const route = useRoute()
const activeMenu = ref('Dashboard')

watch(() => route.name, (newRoute) => {
  if (newRoute) {
    activeMenu.value = String(newRoute).split('-').map(word => 
      word.charAt(0).toUpperCase() + word.slice(1)
    ).join(' ')
  }
})

function handleMenuChange(menu: string) {
  activeMenu.value = menu
}

function handleSearch(query: string) {
  // Search logic handled by individual views
  console.log('Search query:', query)
}
</script>
