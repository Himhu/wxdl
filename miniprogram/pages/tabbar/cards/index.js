const cardApi = require('../../../api/card')
const stationStore = require('../../../store/station')

Page({
  data: {
    cardList: [],
    loading: false,
    page: 1,
    hasMore: true,
    showCreateModal: false,
    inputQuota: '',
    createLoading: false
  },

  onLoad() {
    this.lastLoadedSite = stationStore.currentSite
    this.loadCardList()
  },

  onShow() {
    if (this.lastLoadedSite && this.lastLoadedSite !== stationStore.currentSite) {
      this.lastLoadedSite = stationStore.currentSite
      this.setData({ page: 1, hasMore: true, cardList: [] })
      this.loadCardList()
    }
  },

  onPullDownRefresh() {
    this.setData({ page: 1, hasMore: true })
    this.loadCardList().then(() => wx.stopPullDownRefresh())
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.setData({ page: this.data.page + 1 })
      this.loadCardList()
    }
  },

  async loadCardList() {
    if (this.data.loading) return
    this.setData({ loading: true })
    try {
      const res = await cardApi.getCards({ page: this.data.page, pageSize: 20 })
      const cardList = this.data.page === 1 ? res.list : [...this.data.cardList, ...res.list]
      this.setData({ cardList, hasMore: res.hasMore })
    } catch (err) {
      console.error('加载卡密列表失败', err)
    } finally {
      this.setData({ loading: false })
    }
  },

  onShowCreate() {
    this.setData({ showCreateModal: true, inputQuota: '' })
  },

  onCloseCreate() {
    this.setData({ showCreateModal: false, inputQuota: '' })
  },

  onQuotaInput(e) {
    this.setData({ inputQuota: e.detail.value })
  },

  async onConfirmCreate() {
    if (this.data.createLoading) return
    const quota = parseInt(this.data.inputQuota)
    if (!quota || quota <= 0) {
      wx.showToast({ title: '请输入有效额度', icon: 'none' })
      return
    }

    this.setData({ createLoading: true })
    try {
      const result = await cardApi.createCard({ quota })
      wx.showToast({ title: '创建成功', icon: 'success' })
      this.setData({ showCreateModal: false, inputQuota: '', page: 1, hasMore: true })
      this.loadCardList()

      setTimeout(() => {
        if (result.card && result.card.cardKey) {
          wx.showModal({
            title: '兑换码已创建',
            content: result.card.cardKey,
            confirmText: '复制',
            success: (r) => {
              if (r.confirm) {
                wx.setClipboardData({ data: result.card.cardKey })
              }
            }
          })
        }
      }, 1500)
    } catch (err) {
      wx.showToast({ title: err.message || '创建失败', icon: 'none' })
    } finally {
      this.setData({ createLoading: false })
    }
  },

  onCopyKey(e) {
    const key = e.currentTarget.dataset.key
    if (key) wx.setClipboardData({ data: key })
  },

  onCardDetail(e) {
    const { id } = e.currentTarget.dataset
    wx.navigateTo({ url: '/pages/card/detail/index?id=' + id })
  }
})
