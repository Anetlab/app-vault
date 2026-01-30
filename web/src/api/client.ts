import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig } from 'axios'
import type { ApiError } from '@/types'

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json'
      }
    })

    this.setupInterceptors()
  }

  private setupInterceptors(): void {
    // Request interceptor - add auth token
    this.client.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        const token = localStorage.getItem('authToken')
        const secretKey = sessionStorage.getItem('secretKey')

        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }

        if (secretKey) {
          config.headers['X-Vault-Secret-Key'] = secretKey
        }

        return config
      },
      (error) => {
        return Promise.reject(error)
      }
    )

    // Response interceptor - handle errors
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError<ApiError>) => {
        if (error.response) {
          // Handle 401 Unauthorized
          if (error.response.status === 401) {
            // Don't redirect on login/register endpoints - let the component handle it
            if (error.config?.url && !error.config.url.includes('/auth/login') && !error.config.url.includes('/auth/register')) {
              localStorage.removeItem('authToken')
              sessionStorage.removeItem('secretKey')
              window.location.href = '/login'
            }
          }

          // Handle 403 Forbidden
          if (error.response.status === 403) {
            console.error('Access denied:', error.response.data?.error)
          }
        }

        return Promise.reject(error)
      }
    )
  }

  public getClient(): AxiosInstance {
    return this.client
  }

  // Helper method to set auth token
  public setAuthToken(token: string): void {
    localStorage.setItem('authToken', token)
  }

  // Helper method to set secret key
  public setSecretKey(secretKey: string): void {
    sessionStorage.setItem('secretKey', secretKey)
  }

  // Helper method to clear auth
  public clearAuth(): void {
    localStorage.removeItem('authToken')
    sessionStorage.removeItem('secretKey')
  }
}

export const apiClient = new ApiClient()
export default apiClient.getClient()
