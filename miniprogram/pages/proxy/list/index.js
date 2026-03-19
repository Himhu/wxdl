const config = require('../../../config/index')

Page({
  data: {
    proxyList: [],
    loading: true,
    hasMore: false
  },

  onLoad() {
    this._loadList()
  },

  onPullDownRefresh() {
    this._loadList(() => wx.stopPullDownRefresh())
  },

  onReachBottom() {
    if (this.data.hasMore) {
      // TODO: 分页加载
    }
  },

  _loadList(cb) {
    if (this.data.loading && this.data.proxyList.length) return
    this.setData({ loading: true })
    setTimeout(() => {
      this.setData({
        loading: false,
        proxyList: [
          { id: 1, name: '张三', level: 1, levelName: '一级代理', totalCards: 156, balance: 3200 },
          { id: 2, name: '李四', level: 2, levelName: '二级代理', totalCards: 89, balance: 1500 },
          { id: 3, name: '王五', level: 2, levelName: '二级代理', totalCards: 42, balance: 800 }
        ]
      })
      cb && cb()
    }, 500)
  }
})
