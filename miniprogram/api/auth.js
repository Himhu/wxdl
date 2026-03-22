const config = require('../config/index')

const TOKEN_KEY = 'token'
const USER_INFO_KEY = 'userInfo'

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

module.exports = {
  async wechatLogin({ code, userInfo, inviteCode } = {}) {
    return request('/api/v1/user/auth/wechat/login', 'POST', {
      code,
      nickname: userInfo ? userInfo.nickName : '',
      avatar: userInfo ? userInfo.avatarUrl : '',
      inviteCode: inviteCode || ''
    })
  },

  async refreshUserInfo() {
    return request('/api/v1/user/auth/profile', 'GET')
  },

  async getCurrentInvite() {
    return request('/api/v1/user/auth/invite', 'GET')
  },

  getMiniProgramCodeUrl() {
    const token = wx.getStorageSync(TOKEN_KEY)
    const authQuery = token ? `?token=${encodeURIComponent(token)}` : ''
    return `${config.API_BASE_URL}/api/v1/user/auth/invite/mini-code${authQuery}`
  },

  async applyAgent(data = {}) {
    return request('/api/v1/user/auth/apply-agent', 'POST', data)
  },

  async updateAvatar(avatarUrl) {
    return request('/api/v1/user/auth/avatar', 'PUT', { avatar: avatarUrl })
  }
}

