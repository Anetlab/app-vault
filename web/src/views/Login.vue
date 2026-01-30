<template>
  <div class="min-h-screen flex">
    <!-- Left Side - Form -->
    <div class="flex-1 flex items-center justify-center p-8 bg-dark-bg">
      <div class="w-full max-w-md">
        <!-- Logo -->
        <div class="flex items-center gap-3 mb-8">
          <div class="w-12 h-12 bg-primary rounded-xl flex items-center justify-center">
            <i class="fas fa-lock text-white text-xl"></i>
          </div>
          <div>
            <h1 class="text-2xl font-bold text-white">SecureVault</h1>
            <p class="text-sm text-gray-400">Enterprise Secret Management</p>
          </div>
        </div>

        <!-- Form -->
        <div class="card">
          <h2 class="text-2xl font-bold mb-2">Welcome Back</h2>
          <p class="text-gray-400 mb-6">Sign in to your account</p>

          <form @submit.prevent="handleLogin" class="space-y-4">
            <!-- Email -->
            <div>
              <label class="block text-sm text-gray-400 mb-2">Email Address</label>
              <input
                v-model="email"
                type="email"
                required
                class="input-field"
                placeholder="admin@example.com"
                autocomplete="email"
              />
            </div>

            <!-- Password -->
            <div>
              <label class="block text-sm text-gray-400 mb-2">Password</label>
              <div class="relative">
                <input
                  v-model="password"
                  :type="showPassword ? 'text' : 'password'"
                  required
                  class="input-field pr-10"
                  placeholder="Enter your password"
                  autocomplete="current-password"
                />
                <button
                  type="button"
                  @click="showPassword = !showPassword"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-300"
                >
                  <i :class="showPassword ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
                </button>
              </div>
            </div>

            <!-- Secret Key -->
            <div>
              <label class="block text-sm text-gray-400 mb-2">
                Secret Key
                <i
                  class="fas fa-question-circle ml-1 cursor-help"
                  title="Your unique secret key provided during registration"
                ></i>
              </label>
              <div class="relative">
                <input
                  v-model="secretKey"
                  :type="showSecretKey ? 'text' : 'password'"
                  required
                  class="input-field pr-10 font-mono text-sm"
                  placeholder="A3-XXXXXX-XXXXXX-XXXXX"
                  autocomplete="off"
                />
                <button
                  type="button"
                  @click="showSecretKey = !showSecretKey"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-300"
                >
                  <i :class="showSecretKey ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
                </button>
              </div>
            </div>

            <!-- Remember Me -->
            <div class="flex items-center justify-between">
              <label class="flex items-center gap-2 cursor-pointer">
                <input
                  v-model="rememberMe"
                  type="checkbox"
                  class="w-4 h-4 rounded border-gray-700 bg-dark-bg text-primary focus:ring-primary"
                />
                <span class="text-sm text-gray-400">Remember me</span>
              </label>
              <a href="#" class="text-sm text-primary hover:text-primary-hover">Forgot password?</a>
            </div>

            <!-- Error Message -->
            <div v-if="errorMessage" class="p-3 bg-red-500/20 border border-red-500/50 rounded-lg text-red-400 text-sm">
              <i class="fas fa-exclamation-triangle mr-2"></i>
              {{ errorMessage }}
            </div>

            <!-- Submit Button -->
            <button
              type="submit"
              :disabled="loading"
              class="btn-primary w-full"
            >
              <i v-if="loading" class="fas fa-spinner fa-spin mr-2"></i>
              {{ loading ? 'Signing In...' : 'Sign In' }}
            </button>
          </form>

          <!-- Register Link -->
          <div class="mt-6 text-center">
            <span class="text-gray-400">Don't have an account?</span>
            <RouterLink to="/register" class="text-primary hover:text-primary-hover ml-2">
              Create Account
            </RouterLink>
          </div>
        </div>

        <!-- Version -->
        <div class="mt-6 text-center text-sm text-gray-500">
          v{{ version }} - Enterprise Edition
        </div>
      </div>
    </div>

    <!-- Right Side - Info -->
    <div class="hidden lg:flex flex-1 items-center justify-center p-12 bg-gradient-to-br from-dark-card to-dark-bg">
      <div class="max-w-md space-y-8">
        <div class="space-y-4">
          <h2 class="text-4xl font-bold text-white">Military-Grade Security</h2>
          <p class="text-gray-400 text-lg">
            Double-layer encryption with XChaCha20-Poly1305 and Argon2id key derivation
          </p>
        </div>

        <div class="space-y-4">
          <div class="flex items-start gap-3">
            <div class="w-10 h-10 bg-primary/20 rounded-lg flex items-center justify-center flex-shrink-0">
              <i class="fas fa-shield-alt text-primary"></i>
            </div>
            <div>
              <h3 class="font-semibold text-white mb-1">Zero-Knowledge Architecture</h3>
              <p class="text-sm text-gray-400">Your secrets are encrypted on the client side. We never see your data.</p>
            </div>
          </div>

          <div class="flex items-start gap-3">
            <div class="w-10 h-10 bg-primary/20 rounded-lg flex items-center justify-center flex-shrink-0">
              <i class="fas fa-sync-alt text-primary"></i>
            </div>
            <div>
              <h3 class="font-semibold text-white mb-1">Automatic Key Rotation</h3>
              <p class="text-sm text-gray-400">Built-in key rotation with configurable grace periods.</p>
            </div>
          </div>

          <div class="flex items-start gap-3">
            <div class="w-10 h-10 bg-primary/20 rounded-lg flex items-center justify-center flex-shrink-0">
              <i class="fas fa-history text-primary"></i>
            </div>
            <div>
              <h3 class="font-semibold text-white mb-1">Complete Audit Trail</h3>
              <p class="text-sm text-gray-400">Every action is logged with full context for compliance.</p>
            </div>
          </div>
        </div>

        <div class="pt-6 border-t border-gray-700">
          <p class="text-xs text-gray-500">
            <i class="fas fa-lock mr-2"></i>
            Protected by TLS 1.3 with perfect forward secrecy
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'

const router = useRouter()
const authStore = useAuthStore()
const toastStore = useToastStore()

const version = import.meta.env.VITE_VERSION || '2.1.0'

const email = ref('')
const password = ref('')
const secretKey = ref('')
const rememberMe = ref(false)
const showPassword = ref(false)
const showSecretKey = ref(false)
const loading = ref(false)
const errorMessage = ref('')

async function handleLogin() {
  loading.value = true
  errorMessage.value = ''

  try {
    const result = await authStore.login(email.value, password.value, secretKey.value)
    
    if (result.success) {
      toastStore.success('Welcome back!')
      router.push('/')
    } else {
      errorMessage.value = result.error || 'Login failed'
    }
  } catch (error: any) {
    errorMessage.value = error.message || 'An unexpected error occurred'
  } finally {
    loading.value = false
  }
}
</script>
