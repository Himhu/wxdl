const { storeBindingsBehavior } = require('mobx-miniprogram-bindings')
const userStore = require('../../../store/user')
const authApi = require('../../../api/auth')

Page({
  behaviors: [storeBindingsBehavior],

  storeBindings: {
    store: userStore,
    fields: ['userInfo'],
    actions: ['logout', 'updateUserInfo']
  },

  data: {
    isAgent: false,
    agentLevel: 0,
    hasPendingApplication: false,
    hasShownApprovedToast: false,
    inviteCode: '',
    displayNickname: '微信用户',
    displayAvatarText: '?',
    roleTagClass: 'role-user',
    roleText: '普通用户',
    menuList: []
  },

  onShow() {
    this._refreshRole()
  },

  _buildMenuList(isAgent) {
    const app = getApp()
    const cfg = app.globalData.bootstrapConfig
    const rechargeEnabled = !(cfg && cfg.feature && cfg.feature.recharge_enabled === false)

    const list = [
      {
        title: '我的下级',
        url: '/pages/proxy/list/index',
        iconText: '◆',
        agentOnly: true,
        showLock: !isAgent,
        itemClass: !isAgent ? 'menu-item-disabled' : ''
      },
      {
        title: '数据转移',
        url: '/pages/transfer/index',
        iconText: '⇄',
        agentOnly: false,
        showLock: false,
        itemClass: ''
      }
    ]

    if (rechargeEnabled) {
      list.push({
        title: '充值记录',
        url: '/pages/finance/records/index',
        iconText: '◇',
        agentOnly: true,
        showLock: !isAgent,
        itemClass: !isAgent ? 'menu-item-disabled' : ''
      })
    }

    return list
  },

  async _refreshRole() {
    try {
      const prevIsAgent = this.data.isAgent
      const res = await authApi.refreshUserInfo()
      if (res.userInfo) {
        const nickname = res.userInfo.nickname || '微信用户'
        this.updateUserInfo(res.userInfo)
        const isAgent = res.userInfo.role === 'agent'
        const hasPendingApplication = !!res.pendingApplication
        this.setData({
          isAgent,
          agentLevel: isAgent ? (res.userInfo.agentLevel || 1) : 0,
          hasPendingApplication,
          inviteCode: res.inviteCode || '',
          displayNickname: nickname,
          displayAvatarText: nickname.slice(0, 1) || '?',
          roleTagClass: isAgent ? 'role-agent' : 'role-user',
          roleText: isAgent ? '代理商' : (hasPendingApplication ? '待审核' : '普通用户'),
          menuList: this._buildMenuList(isAgent)
        })

        if (!prevIsAgent && isAgent && !wx.getStorageSync('approvedToastShown')) {
          wx.setStorageSync('approvedToastShown', true)
          wx.showToast({ title: '恭喜，代理申请已通过', icon: 'success' })
        }
      }
    } catch (err) {
      console.error('刷新角色失败', err)
      const userInfo = this.data.userInfo
      const isAgent = !!(userInfo && userInfo.role === 'agent')
      const nickname = (userInfo && userInfo.nickname) || '微信用户'
      this.setData({
        isAgent,
        displayNickname: nickname,
        displayAvatarText: nickname.slice(0, 1) || '?',
        roleTagClass: isAgent ? 'role-agent' : 'role-user',
        roleText: isAgent ? '代理商' : '普通用户',
        menuList: this._buildMenuList(isAgent)
      })
    }
  },

  onMenuTap(e) {
    const { url } = e.currentTarget.dataset
    const agentOnly = e.currentTarget.dataset.agentOnly === 'true' || e.currentTarget.dataset.agentOnly === true

    if (agentOnly && !this.data.isAgent) {
      wx.showModal({
        title: '权限不足',
        content: '该功能仅代理商可用，请联系管理员开通权限',
        showCancel: false
      })
      return
    }

    wx.navigateTo({ url })
  },

  onApplyAgent() {
    if (this.data.hasPendingApplication) {
      wx.showToast({ title: '已提交申请，待审核', icon: 'none' })
      return
    }

    wx.showModal({
      title: '申请成为代理',
      content: '将提交代理申请，请等待管理员审核。',
      success: async (res) => {
        if (!res.confirm) return
        try {
          const pendingInviteCode = wx.getStorageSync('pendingInviteCode') || ''
          await authApi.applyAgent({ inviteCode: pendingInviteCode })
          if (pendingInviteCode) {
            wx.removeStorageSync('pendingInviteCode')
          }
          this.setData({
            hasPendingApplication: true,
            roleText: '待审核'
          })
          wx.showToast({ title: '提交成功，待审核', icon: 'success' })
        } catch (err) {
          wx.showToast({ title: err.message || '提交失败', icon: 'none' })
        }
      }
    })
  },

  onInviteAgent() {
    wx.navigateTo({ url: '/pages/invite/index' })
  },

  onLogout() {
    wx.showModal({
      title: '提示',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          this.logout()
        }
      }
    })
  }
})
