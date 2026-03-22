const authApi = require('../../../api/auth')

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

  _loadList(cb) {
    this.setData({ loading: true })
    authApi.getCurrentInvite()
      .then((res) => {
        const records = (res.records || []).map((item) => {
          const name = item.applicant.nickname || '微信用户'
          return {
            id: item.id,
            name,
            avatar: item.applicant.avatar || '',
            avatarText: name.slice(0, 1) || '?',
            status: item.status,
            statusText: item.status === 'approved' ? '已开通' : item.status === 'rejected' ? '已驳回' : '待审核',
            createdAt: item.createdAt,
            inviteCode: item.inviteCode || '-',
            displayTime: item.createdAt || '-'
          }
        })
        this.setData({
          loading: false,
          proxyList: records,
          hasMore: false
        })
        cb && cb()
      })
      .catch((err) => {
        this.setData({ loading: false })
        wx.showToast({ title: err.message || '加载失败', icon: 'none' })
        cb && cb()
      })
  }
})
