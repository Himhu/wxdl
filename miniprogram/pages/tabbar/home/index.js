const { storeBindingsBehavior } = require('mobx-miniprogram-bindings')
const userStore = require('../../../store/user')
const authApi = require('../../../api/auth')
const http = require('../../../api/request')

Page({
  behaviors: [storeBindingsBehavior],

  storeBindings: [
    {
      store: userStore,
      fields: ['userInfo', 'isLogin'],
      actions: ['updateUserInfo']
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
        const nextUserInfo = {
          ...this.data.userInfo,
          ...res.userInfo,
          agentBalance: res.userInfo.agentBalance || (this.data.userInfo && this.data.userInfo.agentBalance) || '0.00'
        }
        const isAgent = nextUserInfo.role === 'agent'
        this.updateUserInfo(nextUserInfo)
        this.setData({
          isAgent,
          stats: {
            ...this.data.stats,
            balance: Number(nextUserInfo.agentBalance || 0)
          }
        })

        if (isAgent) {
          this._loadCardStats()
        } else {
          this.setData({ stats: { totalCards: 0, usedCards: 0, balance: 0 } })
        }
      }
    } catch (err) {
      console.error('刷新角色失败', err)
      const userInfo = this.data.userInfo
      this.setData({
        isAgent: !!(userInfo && userInfo.role === 'agent'),
        stats: {
          ...this.data.stats,
          balance: Number((userInfo && userInfo.agentBalance) || 0)
        }
      })
    }
  },

  onLogin() {
    wx.navigateTo({ url: '/pages/common/login/index' })
  },

  async _loadCardStats() {
    try {
      const res = await http.get('/api/v1/cards/stats')
      this.setData({
        stats: {
          ...this.data.stats,
          totalCards: Number(res.total || 0),
          usedCards: Number(res.used || 0)
        }
      })
    } catch (err) {
      console.error('加载卡密统计失败', err)
    }
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
    if (!this._checkAgent('操作日志')) return
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
