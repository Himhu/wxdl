import http from './http'

export interface UserInfo {
  id: number
  openId: string
  nickname: string
  avatar: string
  mobile: string
  role: string
  inviterUserId: number | null
  lastLoginAt: string
}

export interface UserListParams {
  page?: number
  pageSize?: number
  role?: string
  keyword?: string
}

export interface UserListResponse {
  list: UserInfo[]
  total: number
  page: number
  pageSize: number
}

export interface UpdateRoleParams {
  role: string
  remark?: string
}

export interface AgentApplication {
  id: number
  inviteCode: string
  status: string
  rejectReason: string
  reviewedByAdminId: number | null
  reviewedAt: string | null
  createdAt: string
  applicant: UserInfo
  inviter?: UserInfo | null
}

const userApi = {
  list(params?: UserListParams) {
    return http.get<any, UserListResponse>(
      '/api/v1/admin/users',
      { params }
    )
  },

  getById(id: number) {
    return http.get<any, { userInfo: UserInfo }>(
      `/api/v1/admin/users/${id}`
    )
  },

  updateRole(id: number, data: UpdateRoleParams) {
    return http.put<any, { userInfo: UserInfo }>(
      `/api/v1/admin/users/${id}/role`,
      data
    )
  },

  listApplications(status?: string) {
    return http.get<any, { list: AgentApplication[] }>(
      '/api/v1/admin/users/applications',
      { params: { status } }
    )
  },

  reviewApplication(id: number, data: { approved: boolean; rejectReason?: string }) {
    return http.post<any, { application: AgentApplication }>(
      `/api/v1/admin/users/applications/${id}/review`,
      data
    )
  },
}

export default userApi
