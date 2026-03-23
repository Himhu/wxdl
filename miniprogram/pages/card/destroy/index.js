const cardApi = require('../../../api/card')
const userStore = require('../../../store/user')

function normalizeCard(detail) {
  const cost = Number(detail.cost)
  return {
    ...detail,
    hasCost: cost > 0,
    costDisplay: cost > 0 ? detail.cost : '',
    remainingYuan: detail.externalRemaining !== undefined ? Math.round(detail.externalRemaining / 500000 * 100) / 100 : 0
  }
}

Page({
  data: {
    keyword: '',
    viewState: 'idle',
    card: null,
    refundPreview: '',
    destroying: false,
    destroyResult: null
  },

  onLoad() {
    const userInfo = userStore.userInfo || {}
    if (userInfo.role !== 'agent') {
      wx.showModal({
        title: '权限不足',
        content: '销毁卡密仅代理可用',
        showCancel: false,
        success: () => wx.navigateBack()
      })
      return
    }
  },

  onKeywordInput(e) {
    this.setData({ keyword: e.detail.value })
  },

  async onSearch() {
    const keyword = this.data.keyword.trim()
    if (!keyword) {
      wx.showToast({ title: '请输入兑换码', icon: 'none' })
      return
    }
    if (this.data.viewState === 'searching') return

    this.setData({ viewState: 'searching', card: null, destroyResult: null })

    try {
      const res = await cardApi.getCards({ keyword, page: 1, pageSize: 20 })
      const matched = (res.list || []).find(item =>
        String(item.cardKey).trim() === keyword
      )

      if (!matched) {
        this.setData({ viewState: 'not-found' })
        return
      }

      const detail = await cardApi.getCardDetail(matched.id)
      const card = normalizeCard(detail)
      const cost = Number(card.cost)

      let refund = ''
      if (card.status === 2 && card.estimatedRefund !== undefined) {
        refund = card.estimatedRefund
      } else {
        refund = cost > 0 ? card.cost : String(card.quota || 0)
      }

      this.setData({ viewState: 'found', card, refundPreview: refund })
    } catch (err) {
      console.error('查询卡密失败', err)
      this.setData({ viewState: 'idle' })
      wx.showToast({ title: '查询失败，请重试', icon: 'none' })
    }
  },

  onCopyKey() {
    const key = this.data.card && this.data.card.cardKey
    if (key) wx.setClipboardData({ data: key })
  },

  onDestroyTap() {
    const card = this.data.card
    if (!card || card.status === 3 || this.data.destroying) return

    wx.showModal({
      title: '确认销毁',
      content: '销毁后将返还 ' + this.data.refundPreview + ' 积分，且不可恢复，是否继续？',
      confirmColor: '#ff4d4f',
      success: (res) => {
        if (res.confirm) this._doDestroy(card.id)
      }
    })
  },

  async _doDestroy(id) {
    this.setData({ destroying: true })
    try {
      const result = await cardApi.destroyCard(id)
      this.setData({
        viewState: 'success',
        destroying: false,
        destroyResult: {
          refundedPoints: result.refundedPoints,
          balanceAfter: result.balanceAfter
        }
      })
    } catch (err) {
      this.setData({ destroying: false })
      wx.showToast({ title: err.message || '销毁失败', icon: 'none' })
    }
  },

  onResetSearch() {
    this.setData({
      keyword: '',
      viewState: 'idle',
      card: null,
      refundPreview: '',
      destroyResult: null
    })
  },

  onBack() {
    wx.navigateBack()
  }
})
