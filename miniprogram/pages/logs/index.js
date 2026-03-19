const proxyApi = require('../../api/proxy')
const stationStore = require('../../store/station')

Page({
  data: {
    logs: [],
    loading: false,
    page: 1,
    hasMore: true
  },

  onLoad() {
    this.lastLoadedSite = stationStore.currentSite
    this.loadLogs()
  },

  onShow() {
    if (this.lastLoadedSite && this.lastLoadedSite !== stationStore.currentSite) {
      this.lastLoadedSite = stationStore.currentSite
      this.setData({ page: 1, hasMore: true, logs: [] })
      this.loadLogs()
    }
  },

  onPullDownRefresh() {
    this.setData({ page: 1, hasMore: true })
    this.loadLogs().then(() => {
      wx.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.setData({ page: this.data.page + 1 })
      this.loadLogs()
    }
  },

  async loadLogs() {
    if (this.data.loading) return

    this.setData({ loading: true })

    try {
      const res = await proxyApi.getOperationLogs({
        page: this.data.page,
        pageSize: 20
      })

      const logs = this.data.page === 1 ? res.list : [...this.data.logs, ...res.list]

      this.setData({
        logs,
        hasMore: res.hasMore
      })
    } catch (err) {
      console.error('加载日志失败', err)
    } finally {
      this.setData({ loading: false })
    }
  }
})
