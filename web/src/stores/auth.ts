import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import { apiClient } from '@/api/client'
import type { User, LoginRequest, RegisterRequest } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  // State
  const token = ref<string | null>(localStorage.getItem('authToken'))
  const user = ref<User | null>(null)
  const secretKey = ref<string | null>(sessionStorage.getItem('secretKey'))
  const password = ref<string | null>(sessionStorage.getItem('vaultPassword'))
  const isAuthenticated = computed(() => !!token.value && !!secretKey.value)

  // Actions
  async function register(email: string, password: string) {
    try {
      const response = await authApi.register({ email, password })
      
      // Store token and user
      token.value = response.token
      user.value = response.user
      apiClient.setAuthToken(response.token)
      
      // Return secret key (only shown once)
      return {
        success: true,
        secretKey: response.secretKey,
        user: response.user
      }
    } catch (error: any) {
      console.error('Registration error:', error)
      return {
        success: false,
        error: error.response?.data?.error || 'Registration failed'
      }
    }
  }

  async function login(email: string, userPassword: string, userSecretKey: string) {
    try {
      const response = await authApi.login({ email, password: userPassword, secretKey: userSecretKey })
      
      // Store auth data
      token.value = response.token
      user.value = response.user
      secretKey.value = userSecretKey
      password.value = userPassword
      
      apiClient.setAuthToken(response.token)
      apiClient.setSecretKey(userSecretKey)
      apiClient.setPassword(userPassword)
      
      return {
        success: true,
        user: response.user
      }
    } catch (error: any) {
      console.error('Login error:', error)
      return {
        success: false,
        error: error.response?.data?.error || 'Login failed'
      }
    }
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      // Clear all auth data
      token.value = null
      user.value = null
      secretKey.value = null
      password.value = null
      apiClient.clearAuth()
    }
  }

  async function changePassword(oldPassword: string, newPassword: string) {
    if (!secretKey.value) {
      throw new Error('Secret key not available')
    }

    try {
      await authApi.changePassword(oldPassword, newPassword, secretKey.value)
      return { success: true }
    } catch (error: any) {
      console.error('Change password error:', error)
      return {
        success: false,
        error: error.response?.data?.error || 'Failed to change password'
      }
    }
  }

  // Initialize auth state on store creation
  function initialize() {
    const storedToken = localStorage.getItem('authToken')
    const storedSecretKey = sessionStorage.getItem('secretKey')
    
    if (storedToken && storedSecretKey) {
      token.value = storedToken
      secretKey.value = storedSecretKey
      apiClient.setAuthToken(storedToken)
      apiClient.setSecretKey(storedSecretKey)
    }
  }

  return {
    // State
    token,
    user,
    secretKey,
    isAuthenticated,
    
    // Actions
    register,
    login,
    logout,
    changePassword,
    initialize
  }
})
