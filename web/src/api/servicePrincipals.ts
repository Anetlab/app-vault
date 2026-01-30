import axios from '@/api/client'
import type {
  ServicePrincipal,
  CreateServicePrincipalRequest,
  ServicePrincipalResponse
} from '@/types'

export const servicePrincipalsApi = {
  async create(data: CreateServicePrincipalRequest): Promise<ServicePrincipalResponse> {
    const response = await axios.post<ServicePrincipalResponse>('/service-principals', data)
    return response.data
  },

  async list(): Promise<ServicePrincipal[]> {
    const response = await axios.get<ServicePrincipal[]>('/service-principals')
    return response.data
  },

  async getById(id: string): Promise<ServicePrincipal> {
    const response = await axios.get<ServicePrincipal>(`/service-principals/${id}`)
    return response.data
  },

  async update(id: string, data: Partial<CreateServicePrincipalRequest>): Promise<ServicePrincipal> {
    const response = await axios.put<ServicePrincipal>(`/service-principals/${id}`, data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await axios.delete(`/service-principals/${id}`)
  },

  async regenerateSecret(id: string): Promise<{ clientSecret: string }> {
    const response = await axios.post<{ clientSecret: string }>(`/service-principals/${id}/regenerate`)
    return response.data
  }
}
