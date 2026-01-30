import axios from '@/api/client'
import type {
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  User
} from '@/types'

export const authApi = {
  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await axios.post<AuthResponse>('/auth/register', data)
    return response.data
  },

  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await axios.post<AuthResponse>('/auth/login', {
      email: data.email,
      password: data.password,
      secret_key: data.secretKey
    })
    return response.data
  },

  async logout(): Promise<void> {
    await axios.post('/auth/logout')
  },

  async getCurrentUser(): Promise<User> {
    const response = await axios.get<User>('/auth/me')
    return response.data
  },

  async changePassword(oldPassword: string, newPassword: string, secretKey: string): Promise<void> {
    await axios.post('/account/change-password', {
      oldPassword,
      newPassword,
      secretKey
    })
  }
}
