const userStore = require('./store/user')

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
  },

  onShow(options) {
    cacheInviteCode(options)
  },

  globalData: {
    version: '1.0.0'
  }
})

