const cardApi = require('../../../api/card')

Page({
  data: {
    amount: '',
    isValid: false,
    isCreating: false,
    successData: null
  },

  handleInput(e) {
    let value = e.detail.value.replace(/\D/g, '')
    if (value.length > 1 && value.startsWith('0')) {
      value = value.replace(/^0+/, '')
    }
    const num = parseInt(value, 10)
    this.setData({ amount: value, isValid: !isNaN(num) && num > 0 })
  },

  async handleCreate() {
    if (!this.data.isValid || this.data.isCreating) return
    const quota = parseInt(this.data.amount, 10)

    this.setData({ isCreating: true })
    wx.showLoading({ title: '创建中...', mask: true })

    try {
      const res = await cardApi.createCard({ quota })
      if (res && res.card) {
        this.setData({ successData: res.card })
      } else {
        throw new Error('返回数据格式错误')
      }
    } catch (err) {
      wx.showToast({ title: err.message || '创建失败', icon: 'none' })
    } finally {
      wx.hideLoading()
      this.setData({ isCreating: false })
    }
  },

  handleCopy() {
    if (!this.data.successData) return
    wx.setClipboardData({
      data: this.data.successData.cardKey,
      success: () => wx.showToast({ title: '复制成功', icon: 'success' })
    })
  },

  handleContinue() {
    this.setData({ successData: null, amount: '', isValid: false })
  }
})
