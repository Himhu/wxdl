const { observable, action } = require('mobx-miniprogram')
const config = require('../config/index')

// 站点状态管理
const stationStore = observable({
  // 当前选中的站点
  currentSite: wx.getStorageSync('currentSite') || config.SITES[0].id,

  // 站点列表
  sites: config.SITES,

  // 获取当前站点信息
  get currentSiteInfo() {
    return this.sites.find(site => site.id === this.currentSite) || this.sites[0]
  },

  // 切换站点
  switchSite: action(function(siteId) {
    this.currentSite = siteId
    wx.setStorageSync('currentSite', siteId)

    // 切换站点后提示用户
    wx.showToast({
      title: `已切换到${this.currentSiteInfo.name}`,
      icon: 'success'
    })
  })
})

module.exports = stationStore
