<template>
  <div class="fixed top-4 right-4 z-50 space-y-2 w-96">
    <TransitionGroup name="slide">
      <div
        v-for="toast in toastStore.toasts"
        :key="toast.id"
        :class="[
          'px-6 py-3 rounded-xl shadow-lg flex items-center gap-3 cursor-pointer',
          toastClasses[toast.type]
        ]"
        @click="toastStore.remove(toast.id)"
      >
        <i :class="toastIcons[toast.type]"></i>
        <span class="flex-1">{{ toast.message }}</span>
        <button
          @click.stop="toastStore.remove(toast.id)"
          class="hover:opacity-70 transition-opacity"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { useToastStore } from '@/stores/toast'

const toastStore = useToastStore()

const toastClasses = {
  success: 'bg-green-500 text-white',
  error: 'bg-red-500 text-white',
  warning: 'bg-yellow-500 text-white',
  info: 'bg-blue-500 text-white'
}

const toastIcons = {
  success: 'fas fa-check-circle',
  error: 'fas fa-exclamation-circle',
  warning: 'fas fa-exclamation-triangle',
  info: 'fas fa-info-circle'
}
</script>
