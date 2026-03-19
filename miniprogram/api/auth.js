const config = require('../config/index')

// 开发模式开关：true 使用本地 mock，false 对接真实后端
const USE_MOCK = false

const TOKEN_KEY = 'token'
const USER_INFO_KEY = 'userInfo'

// ==================== HTTP 请求封装 ====================

const request = (url, method, data) => {
  return new Promise((resolve, reject) => {
    const token = wx.getStorageSync(TOKEN_KEY)
    wx.request({
      url: `${config.API_BASE_URL}${url}`,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {})
      },
      success: (res) => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data.data || res.data)
        } else if (res.statusCode === 401) {
          wx.removeStorageSync(TOKEN_KEY)
          wx.removeStorageSync(USER_INFO_KEY)
          wx.reLaunch({ url: '/pages/common/login/index' })
          reject(new Error('登录已过期'))
        } else {
          reject(new Error(res.data.message || '请求失败'))
        }
      },
      fail: (err) => reject(new Error(err.errMsg || '网络错误'))
    })
  })
}

// ==================== 真实 API ====================

const realApi = {
  async wechatLogin({ code, userInfo } = {}) {
    return request('/api/v1/user/auth/wechat/login', 'POST', {
      code,
      nickname: userInfo ? userInfo.nickName : '',
      avatar: userInfo ? userInfo.avatarUrl : ''
    })
  },

  async refreshUserInfo() {
    return request('/api/v1/user/auth/profile', 'GET')
  }
}

// ==================== Mock API ====================

const delay = (ms = 500) => new Promise((resolve) => setTimeout(resolve, ms))

const MOCK_USERS_KEY = 'mockUsers'

const getDefaultMockUsers = () => [
  {
    id: 'user_001',
    openId: 'mock_openid_agent1',
    nickname: '张三',
    avatar: '',
    mobile: '13800000001',
    role: 'agent',
    agentLevel: 1,
    agentLevelName: '一级代理',
    parentUserId: null,
    lastLoginAt: ''
  },
  {
    id: 'user_002',
    openId: 'mock_openid_agent2',
    nickname: '李四',
    avatar: '',
    mobile: '13800000002',
    role: 'agent',
    agentLevel: 2,
    agentLevelName: '二级代理',
    parentUserId: 'user_001',
    lastLoginAt: ''
  }
]

const getMockUsers = () => {
  let users = wx.getStorageSync(MOCK_USERS_KEY)
  if (!users || !Array.isArray(users) || users.length === 0) {
    users = getDefaultMockUsers()
    wx.setStorageSync(MOCK_USERS_KEY, users)
  }
  return users
}

const saveMockUsers = (users) => {
  wx.setStorageSync(MOCK_USERS_KEY, users)
}

const getMockOpenId = () => {
  let openId = wx.getStorageSync('mockOpenId')
  if (!openId) {
    openId = `mock_openid_${Date.now().toString(36)}`
    wx.setStorageSync('mockOpenId', openId)
  }
  return openId
}

const buildUserResponse = (user) => ({
  id: user.id,
  openId: user.openId,
  nickname: user.nickname,
  avatar: user.avatar,
  mobile: user.mobile,
  role: user.role,
  agentLevel: user.agentLevel,
  agentLevelName: user.agentLevelName,
  parentUserId: user.parentUserId,
  lastLoginAt: user.lastLoginAt
})

const mockApi = {
  async wechatLogin({ code, userInfo } = {}) {
    await delay()
    if (!code) throw new Error('获取微信登录凭证失败')

    const openId = getMockOpenId()
    const users = getMockUsers()
    let user = users.find((u) => u.openId === openId)

    const now = new Date().toISOString()

    if (user) {
      if (userInfo) {
        user.nickname = userInfo.nickName || user.nickname
        user.avatar = userInfo.avatarUrl || user.avatar
      }
      user.lastLoginAt = now
      saveMockUsers(users)
    } else {
      user = {
        id: `user_${Date.now().toString(36)}`,
        openId,
        nickname: userInfo ? userInfo.nickName : '微信用户',
        avatar: userInfo ? userInfo.avatarUrl : '',
        mobile: '',
        role: 'user',
        agentLevel: null,
        agentLevelName: '',
        parentUserId: null,
        lastLoginAt: now
      }
      users.push(user)
      saveMockUsers(users)
    }

    return {
      token: `mock_token_${user.id}_${Date.now()}`,
      userInfo: buildUserResponse(user)
    }
  },

  async refreshUserInfo() {
    await delay(200)
    const openId = getMockOpenId()
    const users = getMockUsers()
    const user = users.find((u) => u.openId === openId)
    if (!user) throw new Error('用户不存在')
    return { userInfo: buildUserResponse(user) }
  },

  async mockPromoteToAgent(userId, agentLevel = 1) {
    await delay(300)
    const users = getMockUsers()
    const user = users.find((u) => u.id === userId)
    if (!user) throw new Error('用户不存在')

    const levelInfo = Object.values(config.AGENT_LEVEL).find((l) => l.value === agentLevel)
    user.role = 'agent'
    user.agentLevel = agentLevel
    user.agentLevelName = levelInfo ? levelInfo.label : '代理商'
    saveMockUsers(users)
    return { userInfo: buildUserResponse(user) }
  },

  async mockDemoteToUser(userId) {
    await delay(300)
    const users = getMockUsers()
    const user = users.find((u) => u.id === userId)
    if (!user) throw new Error('用户不存在')

    user.role = 'user'
    user.agentLevel = null
    user.agentLevelName = ''
    user.parentUserId = null
    saveMockUsers(users)
    return { userInfo: buildUserResponse(user) }
  }
}

// ==================== 导出（根据 USE_MOCK 切换） ====================

const api = USE_MOCK ? mockApi : realApi

module.exports = {
  wechatLogin: api.wechatLogin,
  refreshUserInfo: api.refreshUserInfo,
  // Mock 专用（仅开发阶段）
  mockPromoteToAgent: mockApi.mockPromoteToAgent,
  mockDemoteToUser: mockApi.mockDemoteToUser,
  // 常量
  USE_MOCK
}
