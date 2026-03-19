Page({
  data: {
    balance: 1280,
    amounts: [100, 200, 500, 1000, 2000, 5000],
    selectedIndex: -1,
    customAmount: ''
  },

  onSelectAmount(e) {
    const index = e.currentTarget.dataset.index
    this.setData({ selectedIndex: index, customAmount: '' })
  },

  onCustomInput(e) {
    this.setData({ customAmount: e.detail.value, selectedIndex: -1 })
  },

  onSubmit() {
    const amount = this.data.selectedIndex >= 0
      ? this.data.amounts[this.data.selectedIndex]
      : Number(this.data.customAmount)

    if (!amount || amount <= 0) {
      return wx.showToast({ title: '请选择充值金额', icon: 'none' })
    }
    wx.showToast({ title: '功能开发中', icon: 'none' })
  }
})
