import http from './http'

export interface UserInfo {
  id: number
  openId: string
  nickname: string
  avatar: string
  mobile: string
  role: string
  agentLevel: number | null
  agentLevelName: string
  parentUserId: number | null
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
  agentLevel?: number | null
  parentUserId?: number | null
  remark?: string
}

const userApi = {
  list(params?: UserListParams) {
    return http.get<any, { code: number; message: string; data: UserListResponse }>(
      '/api/v1/admin/users',
      { params }
    )
  },

  getById(id: number) {
    return http.get<any, { code: number; message: string; data: { userInfo: UserInfo } }>(
      `/api/v1/admin/users/${id}`
    )
  },

  updateRole(id: number, data: UpdateRoleParams) {
    return http.put<any, { code: number; message: string; data: { userInfo: UserInfo } }>(
      `/api/v1/admin/users/${id}/role`,
      data
    )
  },
}

export default userApi
