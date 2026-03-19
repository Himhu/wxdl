const userStore = require('./store/user')

App({
  onLaunch() {
    userStore.restoreLoginState()
    console.log('小程序启动')
  },

  globalData: {
    version: '1.0.0'
  }
})
