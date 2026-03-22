const transferApi = require('../../api/transfer')
const userStore = require('../../store/user')

Page({
  data: {
    username: '',
    password: '',
    queryLoading: false,
    confirmLoading: false,
    result: null,
    transferred: false
  },

  onUsernameInput(e) {
    this.setData({ username: e.detail.value, result: null, transferred: false })
  },

  onPasswordInput(e) {
    this.setData({ password: e.detail.value, result: null, transferred: false })
  },

  async onQuery() {
    const { username, password } = this.data
    if (!username || !password) {
      return wx.showToast({ title: '请输入账号和密码', icon: 'none' })
    }

    this.setData({ queryLoading: true, result: null, transferred: false })

    try {
      const res = await transferApi.fetchLegacyBalance(username, password)
      this.setData({ result: res })
      wx.showToast({ title: '查询成功', icon: 'success' })
    } catch (err) {
      wx.showToast({ title: err.message || '查询失败', icon: 'none' })
    } finally {
      this.setData({ queryLoading: false })
    }
  },

  onConfirmTransfer() {
    const { result, confirmLoading, transferred } = this.data
    if (!result || confirmLoading || transferred) {
      return
    }

    wx.showModal({
      title: '确认转移',
      content: `确认将旧站账号「${result.legacyUsername}」的余额转移到新站吗？转移成功后不可重复操作。`,
      cancelText: '再想想',
      confirmText: '确认转移',
      success: (res) => {
        if (res.confirm) {
          this._confirmTransfer()
        }
      }
    })
  },

  async _confirmTransfer() {
    const { username, password, confirmLoading, transferred } = this.data
    if (confirmLoading || transferred) {
      return
    }
    this.setData({ confirmLoading: true })

    try {
      const res = await transferApi.confirmLegacyTransfer(username, password)
      if (res.token && res.userInfo) {
        userStore.setLoginState({ token: res.token, userInfo: res.userInfo })
      }
      this.setData({ transferred: true, result: { ...this.data.result, ...res, balance: res.newBalance || this.data.result.balance } })
      wx.showToast({ title: '转移成功', icon: 'success' })
    } catch (err) {
      wx.showToast({ title: err.message || '转移失败', icon: 'none' })
    } finally {
      this.setData({ confirmLoading: false })
    }
  }
})
