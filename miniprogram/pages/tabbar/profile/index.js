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
    menuList: [
      { title: '下级代理', url: '/pages/proxy/list/index', icon: '◆', agentOnly: true },
      { title: '充值记录', url: '/pages/finance/records/index', icon: '◇', agentOnly: true },
      { title: '操作日志', url: '/pages/logs/index', icon: '☰', agentOnly: false }
    ]
  },

  onShow() {
    this._refreshRole()
  },

  async _refreshRole() {
    try {
      const res = await authApi.refreshUserInfo()
      if (res.userInfo) {
        this.updateUserInfo(res.userInfo)
        this.setData({ isAgent: res.userInfo.role === 'agent' })
      }
    } catch (err) {
      console.error('刷新角色失败', err)
      // 降级使用本地缓存
      const userInfo = this.data.userInfo
      this.setData({ isAgent: !!(userInfo && userInfo.role === 'agent') })
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

  // Mock: 测试用 — 模拟管理员提升为代理
  async onMockPromote() {
    try {
      const res = await authApi.mockPromoteToAgent(this.data.userInfo.id, 1)
      this.updateUserInfo(res.userInfo)
      this.setData({ isAgent: true })
      wx.showToast({ title: '已提升为代理', icon: 'success' })
    } catch (err) {
      wx.showToast({ title: err.message, icon: 'none' })
    }
  },

  // Mock: 测试用 — 模拟降级为普通用户
  async onMockDemote() {
    try {
      const res = await authApi.mockDemoteToUser(this.data.userInfo.id)
      this.updateUserInfo(res.userInfo)
      this.setData({ isAgent: false })
      wx.showToast({ title: '已降为普通用户', icon: 'success' })
    } catch (err) {
      wx.showToast({ title: err.message, icon: 'none' })
    }
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
