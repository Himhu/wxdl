const { storeBindingsBehavior } = require('mobx-miniprogram-bindings')
const userStore = require('../../../store/user')
const stationStore = require('../../../store/station')
const authApi = require('../../../api/auth')

Page({
  behaviors: [storeBindingsBehavior],

  storeBindings: [
    {
      store: userStore,
      fields: ['userInfo', 'isLogin'],
      actions: ['updateUserInfo']
    },
    {
      store: stationStore,
      fields: ['currentSiteInfo']
    }
  ],

  data: {
    isAgent: false,
    stats: {
      totalCards: 0,
      usedCards: 0,
      balance: 0
    }
  },

  onShow() {
    if (this.data.isLogin) {
      this._refreshRole()
    }
  },

  async _refreshRole() {
    try {
      const res = await authApi.refreshUserInfo()
      if (res.userInfo) {
        this.updateUserInfo(res.userInfo)
        this.setData({ isAgent: res.userInfo.role === 'agent' })
        if (res.userInfo.role === 'agent') {
          this.loadStats()
        }
      }
    } catch (err) {
      console.error('刷新角色失败', err)
      const userInfo = this.data.userInfo
      this.setData({ isAgent: !!(userInfo && userInfo.role === 'agent') })
    }
  },

  onLogin() {
    wx.navigateTo({ url: '/pages/common/login/index' })
  },

  async loadStats() {
    // TODO: 调用API获取统计数据
    console.log('加载统计数据')
  },

  onCardManage() {
    if (!this._checkAgent('卡密管理')) return
    wx.switchTab({ url: '/pages/tabbar/cards/index' })
  },

  onRecharge() {
    if (!this._checkAgent('充值积分')) return
    wx.navigateTo({ url: '/pages/finance/recharge/index' })
  },

  onViewLogs() {
    wx.navigateTo({ url: '/pages/logs/index' })
  },

  _checkAgent(featureName) {
    if (!this.data.isLogin) {
      wx.showModal({
        title: '需要登录',
        content: `${featureName}需要先登录`,
        confirmText: '去登录',
        success: (res) => {
          if (res.confirm) this.onLogin()
        }
      })
      return false
    }

    if (!this.data.isAgent) {
      wx.showModal({
        title: '权限不足',
        content: `${featureName}仅代理商可用，请联系管理员开通权限`,
        showCancel: false
      })
      return false
    }

    return true
  }
})
