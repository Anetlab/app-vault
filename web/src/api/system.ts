import axios from '@/api/client'
import type {
  KeyVersion,
  KeyRotationRequest,
  KeyRotationResponse,
  KeyStatus,
  AuditLog,
  HealthCheck
} from '@/types'

export const keyRotationApi = {
  async rotate(data: KeyRotationRequest): Promise<KeyRotationResponse> {
    const response = await axios.post<KeyRotationResponse>('/keys/rotate', data)
    return response.data
  },

  async getStatus(): Promise<KeyStatus> {
    const response = await axios.get<KeyStatus>('/keys/status')
    return response.data
  },

  async getHistory(): Promise<KeyVersion[]> {
    const response = await axios.get<KeyVersion[]>('/keys/history')
    return response.data
  }
}

export const auditLogsApi = {
  async list(limit = 100, offset = 0): Promise<AuditLog[]> {
    const response = await axios.get<AuditLog[]>(`/audit-logs?limit=${limit}&offset=${offset}`)
    return response.data
  },

  async getByResourceId(resourceId: string): Promise<AuditLog[]> {
    const response = await axios.get<AuditLog[]>(`/audit-logs/resource/${resourceId}`)
    return response.data
  }
}

export const systemApi = {
  async health(): Promise<HealthCheck> {
    const response = await axios.get<HealthCheck>('/health')
    return response.data
  }
}
