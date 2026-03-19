const config = require('../../../config/index')

Page({
  data: {
    cardInfo: {},
    infoList: []
  },

  onLoad(options) {
    const cardInfo = {
      id: options.id || '1',
      cardKey: 'ABCD-EFGH-IJKL-MNOP',
      status: 0,
      createTime: '2024-01-15 14:30',
      usedTime: '',
      siteName: '站点A'
    }

    this.setData({
      cardInfo,
      infoList: [
        { label: '创建时间', value: cardInfo.createTime },
        { label: '使用时间', value: cardInfo.usedTime || '未使用' },
        { label: '所属站点', value: cardInfo.siteName }
      ]
    })
  },

  onCopyKey() {
    const key = this.data.cardInfo.cardKey
    if (!key) return wx.showToast({ title: '卡密为空', icon: 'none' })
    wx.setClipboardData({
      data: key,
      success: () => wx.showToast({ title: '已复制', icon: 'success' })
    })
  }
})
