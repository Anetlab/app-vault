import axios from '@/api/client'
import type {
  Secret,
  SecretWithValue,
  CreateSecretRequest,
  UpdateSecretRequest
} from '@/types'

// Transform API response to match frontend types
function transformSecret(data: any): Secret {
  return {
    id: data.id,
    userId: data.user_id || data.userId,
    name: data.name,
    type: data.secret_type || data.type,
    tags: data.tags || [],
    version: data.version,
    createdAt: data.created_at || data.createdAt,
    updatedAt: data.updated_at || data.updatedAt,
    expiresAt: data.expires_at || data.expiresAt,
    accessCount: data.access_count || data.accessCount,
    lastAccessed: data.last_accessed_at || data.lastAccessed
  }
}

function transformSecretWithValue(data: any): SecretWithValue {
  return {
    ...transformSecret(data),
    value: data.value,
    publicKey: data.public_key || data.publicKey
  }
}

export const secretsApi = {
  async create(data: CreateSecretRequest): Promise<Secret> {
    const response = await axios.post<any>('/secrets', data)
    return transformSecret(response.data)
  },

  async list(name?: string, tags?: string[]): Promise<Secret[]> {
    const params = new URLSearchParams()
    if (name) params.append('name', name)
    if (tags && tags.length > 0) {
      tags.forEach(tag => params.append('tags', tag))
    }
    
    const response = await axios.get<any[]>(`/secrets?${params.toString()}`)
    return response.data.map(transformSecret)
  },

  async getById(id: string): Promise<SecretWithValue> {
    const response = await axios.get<any>(`/secrets/${id}`)
    return transformSecretWithValue(response.data)
  },

  async getByName(name: string): Promise<SecretWithValue> {
    const response = await axios.get<any>(`/secrets/by-name?name=${encodeURIComponent(name)}`)
    return transformSecretWithValue(response.data)
  },

  async update(id: string, data: UpdateSecretRequest): Promise<Secret> {
    const response = await axios.put<any>(`/secrets/${id}`, data)
    return transformSecret(response.data)
  },

  async delete(id: string): Promise<void> {
    await axios.delete(`/secrets/${id}`)
  }
}
