import http from './http'

export interface AgentInfo {
  id: number
  username: string
  realName: string
  phone: string
  balance: string
  status: number
  stationId: number
  wechatBound: boolean
  createdAt: string
  updatedAt: string
}

export interface AgentListParams {
  page?: number
  pageSize?: number
  status?: number | ''
  keyword?: string
}

export interface AgentListResponse {
  list: AgentInfo[]
  total: number
  page: number
  pageSize: number
}

const agentApi = {
  list(params?: AgentListParams) {
    return http.get<any, AgentListResponse>('/api/v1/admin/agents', { params })
  },

  updateStatus(id: number, status: number) {
    return http.put<any, { agent: AgentInfo }>(`/api/v1/admin/agents/${id}/status`, { status })
  },
}

export default agentApi
