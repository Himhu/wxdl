const config = require('../../../config/index')

Page({
  data: {
    records: [],
    loading: true,
    hasMore: false
  },

  onLoad() {
    this._loadRecords()
  },

  onPullDownRefresh() {
    this._loadRecords(() => wx.stopPullDownRefresh())
  },

  onReachBottom() {
    if (this.data.hasMore) {
      // TODO: 分页加载
    }
  },

  _loadRecords(cb) {
    if (this.data.loading && this.data.records.length) return
    this.setData({ loading: true })
    setTimeout(() => {
      this.setData({
        loading: false,
        records: [
          { id: 1, amount: 500, status: 1, createTime: '2024-01-15 14:30', remark: '微信支付' },
          { id: 2, amount: 1000, status: 0, createTime: '2024-01-14 10:20', remark: '支付宝转账' },
          { id: 3, amount: 200, status: -1, createTime: '2024-01-13 09:15', remark: '银行转账' }
        ]
      })
      cb && cb()
    }, 500)
  }
})
