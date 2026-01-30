import axios from '@/api/client'
import type {
  Secret,
  SecretWithValue,
  CreateSecretRequest,
  UpdateSecretRequest
} from '@/types'

export const secretsApi = {
  async create(data: CreateSecretRequest): Promise<Secret> {
    const response = await axios.post<Secret>('/secrets', data)
    return response.data
  },

  async list(name?: string, tags?: string[]): Promise<Secret[]> {
    const params = new URLSearchParams()
    if (name) params.append('name', name)
    if (tags && tags.length > 0) {
      tags.forEach(tag => params.append('tags', tag))
    }
    
    const response = await axios.get<Secret[]>(`/secrets?${params.toString()}`)
    return response.data
  },

  async getById(id: string): Promise<SecretWithValue> {
    const response = await axios.get<SecretWithValue>(`/secrets/${id}`)
    return response.data
  },

  async getByName(name: string): Promise<SecretWithValue> {
    const response = await axios.get<SecretWithValue>(`/secrets/by-name?name=${encodeURIComponent(name)}`)
    return response.data
  },

  async update(id: string, data: UpdateSecretRequest): Promise<Secret> {
    const response = await axios.put<Secret>(`/secrets/${id}`, data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await axios.delete(`/secrets/${id}`)
  }
}
