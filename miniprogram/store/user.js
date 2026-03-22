const { observable, action } = require('mobx-miniprogram')

const userStore = observable({
  // 状态
  token: wx.getStorageSync('token') || '',
  userInfo: wx.getStorageSync('userInfo') || null,
  isLogin: !!wx.getStorageSync('token'),

  // 计算属性用方法代替（mobx-miniprogram 不支持 getter）
  getIsAgent() {
    return !!(this.userInfo && this.userInfo.role === 'agent')
  },

  getRole() {
    return this.userInfo ? this.userInfo.role : ''
  },

  setLoginState: action(function({ token, userInfo }) {
    this.token = token || ''
    this.userInfo = userInfo || null
    this.isLogin = !!token

    if (token) {
      wx.setStorageSync('token', token)
    } else {
      wx.removeStorageSync('token')
    }

    if (userInfo) {
      wx.setStorageSync('userInfo', userInfo)
    } else {
      wx.removeStorageSync('userInfo')
    }
  }),

  // 仅更新用户信息（角色刷新时用）
  updateUserInfo: action(function(userInfo) {
    this.userInfo = userInfo || null
    if (userInfo) {
      wx.setStorageSync('userInfo', userInfo)
    } else {
      wx.removeStorageSync('userInfo')
    }
  }),

  restoreLoginState: action(function() {
    const token = wx.getStorageSync('token') || ''
    const userInfo = wx.getStorageSync('userInfo') || null

    this.token = token
    this.userInfo = userInfo
    this.isLogin = !!token
  }),

  clearLoginState: action(function() {
    this.token = ''
    this.userInfo = null
    this.isLogin = false
    wx.removeStorageSync('token')
    wx.removeStorageSync('userInfo')
    wx.removeStorageSync('approvedToastShown')
  }),

  logout: action(function() {
    this.clearLoginState()
    wx.reLaunch({
      url: '/pages/common/login/index'
    })
  })
})

module.exports = userStore
