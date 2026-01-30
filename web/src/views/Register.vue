<template>
  <div class="min-h-screen flex items-center justify-center p-8 bg-dark-bg">
    <div class="w-full max-w-2xl">
      <!-- Logo -->
      <div class="flex items-center gap-3 mb-8 justify-center">
        <div class="w-12 h-12 bg-primary rounded-xl flex items-center justify-center">
          <i class="fas fa-lock text-white text-xl"></i>
        </div>
        <div>
          <h1 class="text-2xl font-bold text-white">SecureVault</h1>
          <p class="text-sm text-gray-400">Enterprise Secret Management</p>
        </div>
      </div>

      <!-- Registration Complete -->
      <div v-if="registrationComplete" class="card">
        <div class="text-center mb-6">
          <div class="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center mx-auto mb-4">
            <i class="fas fa-check text-green-500 text-2xl"></i>
          </div>
          <h2 class="text-2xl font-bold mb-2">Registration Successful!</h2>
          <p class="text-gray-400">Your account has been created</p>
        </div>

        <!-- Secret Key Display -->
        <div class="p-6 bg-yellow-500/10 border border-yellow-500/30 rounded-xl mb-6">
          <div class="flex items-start gap-3 mb-4">
            <i class="fas fa-exclamation-triangle text-yellow-500 text-xl mt-1"></i>
            <div>
              <h3 class="font-semibold text-yellow-500 mb-2">IMPORTANT: Save Your Secret Key</h3>
              <p class="text-sm text-gray-300 mb-4">
                This is your unique secret key. <strong>Write it down and store it securely.</strong>
                You will need it every time you log in. We cannot recover it if you lose it.
              </p>
            </div>
          </div>

          <div class="bg-dark-bg p-4 rounded-lg">
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm text-gray-400">Your Secret Key:</span>
              <button
                @click="copySecretKey"
                class="text-primary hover:text-primary-hover text-sm flex items-center gap-2"
              >
                <i :class="copied ? 'fas fa-check' : 'fas fa-copy'"></i>
                {{ copied ? 'Copied!' : 'Copy' }}
              </button>
            </div>
            <div class="font-mono text-sm text-white bg-dark-card p-3 rounded border border-gray-700 break-all">
              {{ generatedSecretKey }}
            </div>
            <p class="text-xs text-gray-500 mt-2">
              <i class="fas fa-info-circle mr-1"></i>
              Click "Copy" button above to safely copy your secret key
            </p>
          </div>
        </div>

        <div class="space-y-3">
          <label class="flex items-center gap-3 cursor-pointer p-3 bg-dark-card rounded-lg border border-gray-700">
            <input
              v-model="confirmedSaved"
              type="checkbox"
              class="w-5 h-5 rounded border-gray-700 bg-dark-bg text-primary focus:ring-primary"
            />
            <span class="text-sm text-gray-300">
              I have saved my secret key in a secure location
            </span>
          </label>

          <button
            @click="goToLogin"
            :disabled="!confirmedSaved"
            class="btn-primary w-full"
            :class="{ 'opacity-50 cursor-not-allowed': !confirmedSaved }"
          >
            Continue to Login
          </button>
        </div>
      </div>

      <!-- Registration Form -->
      <div v-else class="card">
        <h2 class="text-2xl font-bold mb-2">Create Account</h2>
        <p class="text-gray-400 mb-6">Start managing your secrets securely</p>

        <form @submit.prevent="handleRegister" class="space-y-4">
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
                minlength="8"
                class="input-field pr-10"
                placeholder="Min. 8 characters"
                autocomplete="new-password"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-300"
              >
                <i :class="showPassword ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
              </button>
            </div>
            <div class="mt-2 text-xs text-gray-500">
              Password must be at least 8 characters long
            </div>
          </div>

          <!-- Confirm Password -->
          <div>
            <label class="block text-sm text-gray-400 mb-2">Confirm Password</label>
            <div class="relative">
              <input
                v-model="confirmPassword"
                :type="showConfirmPassword ? 'text' : 'password'"
                required
                class="input-field pr-10"
                placeholder="Re-enter your password"
                autocomplete="new-password"
              />
              <button
                type="button"
                @click="showConfirmPassword = !showConfirmPassword"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-300"
              >
                <i :class="showConfirmPassword ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
              </button>
            </div>
          </div>

          <!-- Terms -->
          <label class="flex items-start gap-3 cursor-pointer">
            <input
              v-model="acceptedTerms"
              type="checkbox"
              required
              class="w-4 h-4 mt-1 rounded border-gray-700 bg-dark-bg text-primary focus:ring-primary"
            />
            <span class="text-sm text-gray-400">
              I agree to the <a href="#" class="text-primary hover:text-primary-hover">Terms of Service</a>
              and <a href="#" class="text-primary hover:text-primary-hover">Privacy Policy</a>
            </span>
          </label>

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
            {{ loading ? 'Creating Account...' : 'Create Account' }}
          </button>
        </form>

        <!-- Login Link -->
        <div class="mt-6 text-center">
          <span class="text-gray-400">Already have an account?</span>
          <RouterLink to="/login" class="text-primary hover:text-primary-hover ml-2">
            Sign In
          </RouterLink>
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

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const acceptedTerms = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const loading = ref(false)
const errorMessage = ref('')

const registrationComplete = ref(false)
const generatedSecretKey = ref('')
const confirmedSaved = ref(false)
const copied = ref(false)

async function handleRegister() {
  errorMessage.value = ''

  // Validation
  if (password.value !== confirmPassword.value) {
    errorMessage.value = 'Passwords do not match'
    return
  }

  if (password.value.length < 8) {
    errorMessage.value = 'Password must be at least 8 characters long'
    return
  }

  if (!acceptedTerms.value) {
    errorMessage.value = 'You must accept the terms and conditions'
    return
  }

  loading.value = true

  try {
    const result = await authStore.register(email.value, password.value)
    
    if (result.success && result.secretKey) {
      generatedSecretKey.value = result.secretKey.trim()
      registrationComplete.value = true
      toastStore.success('Account created successfully!')
    } else {
      errorMessage.value = result.error || 'Registration failed'
    }
  } catch (error: any) {
    errorMessage.value = error.message || 'An unexpected error occurred'
  } finally {
    loading.value = false
  }
}

async function copySecretKey() {
  try {
    await navigator.clipboard.writeText(generatedSecretKey.value)
    copied.value = true
    toastStore.success('Secret key copied to clipboard')
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (error) {
    toastStore.error('Failed to copy to clipboard')
  }
}

function goToLogin() {
  router.push('/login')
}
</script>
