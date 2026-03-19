const cardApi = require('../../../api/card')
const stationStore = require('../../../store/station')

Page({
  data: {
    cardList: [],
    loading: false,
    page: 1,
    hasMore: true
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
    this.loadCardList().then(() => {
      wx.stopPullDownRefresh()
    })
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
      const res = await cardApi.getCards({
        page: this.data.page,
        pageSize: 20
      })

      const cardList = this.data.page === 1 ? res.list : [...this.data.cardList, ...res.list]

      this.setData({
        cardList,
        hasMore: res.hasMore
      })
    } catch (err) {
      console.error('加载卡密列表失败', err)
    } finally {
      this.setData({ loading: false })
    }
  },

  onCardDetail(e) {
    const { id } = e.currentTarget.dataset
    wx.navigateTo({
      url: `/pages/card/detail/index?id=${id}`
    })
  }
})
