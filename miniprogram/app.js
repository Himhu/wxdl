const userStore = require('./store/user')
const http = require('./api/request')

function cacheInviteCode(options) {
  const query = options && options.query ? options.query : {}
  if (query.inviteCode) {
    wx.setStorageSync('pendingInviteCode', query.inviteCode)
  }
}

App({
  onLaunch(options) {
    userStore.restoreLoginState()
    cacheInviteCode(options)
    this._loadBootstrapConfig()
  },

  onShow(options) {
    cacheInviteCode(options)
  },

  async _loadBootstrapConfig() {
    try {
      const cached = wx.getStorageSync('bootstrapConfig')
      if (cached) {
        this.globalData.bootstrapConfig = cached
      }
      const res = await http.get('/api/v1/miniapp/config/bootstrap')
      if (res && res.config) {
        this.globalData.bootstrapConfig = res.config
        wx.setStorageSync('bootstrapConfig', res.config)
      }
    } catch (err) {
      console.error('加载 bootstrap 配置失败', err)
    }
  },

  globalData: {
    version: '1.0.0',
    bootstrapConfig: null
  }
})

