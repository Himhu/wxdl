const { storeBindingsBehavior } = require('mobx-miniprogram-bindings')
const userStore = require('../../../store/user')
const authApi = require('../../../api/auth')
const config = require('../../../config/index')

Page({
  behaviors: [storeBindingsBehavior],

  storeBindings: {
    store: userStore,
    fields: ['isLogin'],
    actions: ['setLoginState']
  },

  data: {
    wechatLoading: false,
    avatarUrl: '',
    avatarPermanentUrl: '',
    avatarUploading: false,
    nickname: ''
  },

  onShow() {
    if (this.data.isLogin) {
      wx.switchTab({ url: '/pages/tabbar/home/index' })
    }
  },

  onChooseAvatar(e) {
    const { avatarUrl } = e.detail || {}
    if (avatarUrl) {
      this.setData({ avatarUrl, avatarPermanentUrl: '' })
      this._silentUploadAvatar(avatarUrl)
    }
  },

  _silentUploadAvatar(tempFilePath) {
    if (!tempFilePath) return
    this.setData({ avatarUploading: true })
    wx.uploadFile({
      url: config.API_BASE_URL + '/api/v1/miniapp/upload-avatar',
      filePath: tempFilePath,
      name: 'file',
      success: (res) => {
        try {
          const data = JSON.parse(res.data)
          if (data.code === 0 && data.data && data.data.url) {
            this.setData({ avatarPermanentUrl: data.data.url })
          }
        } catch (e) {
          console.error('头像上传解析失败', e)
        }
      },
      fail: (err) => console.error('头像上传失败', err),
      complete: () => this.setData({ avatarUploading: false })
    })
  },

  onNicknameInput(e) {
    this.setData({ nickname: e.detail.value || '' })
  },

  onNicknameBlur(e) {
    this.setData({ nickname: e.detail.value || '' })
  },

  getWxLoginCode() {
    return new Promise((resolve, reject) => {
      wx.login({
        success: (res) => {
          if (res.code) {
            resolve(res.code)
            return
          }
          reject(new Error('未获取到微信登录凭证'))
        },
        fail: () => reject(new Error('wx.login 调用失败'))
      })
    })
  },

  async onWechatLogin() {
    if (this.data.wechatLoading) return

    const nickname = (this.data.nickname || '').trim()
    const avatarUrl = this.data.avatarUrl || ''

    if (!avatarUrl) {
      wx.showToast({ title: '请先选择头像', icon: 'none' })
      return
    }
    if (!nickname) {
      wx.showToast({ title: '请先填写昵称', icon: 'none' })
      return
    }

    if (this.data.avatarUploading) {
      wx.showToast({ title: '头像上传中，请稍候', icon: 'none' })
      return
    }

    this.setData({ wechatLoading: true })

    try {
      const code = await this.getWxLoginCode()
      const pendingInviteCode = wx.getStorageSync('pendingInviteCode') || ''
      const finalAvatarUrl = this.data.avatarPermanentUrl || avatarUrl

      const res = await authApi.wechatLogin({
        code,
        userInfo: {
          nickName: nickname,
          avatarUrl: finalAvatarUrl
        },
        inviteCode: pendingInviteCode
      })

      this.setLoginState({
        token: res.token,
        userInfo: res.userInfo
      })

      if (pendingInviteCode && res.userInfo.role !== 'agent') {
        try {
          await authApi.applyAgent({ inviteCode: pendingInviteCode })
          wx.removeStorageSync('pendingInviteCode')
        } catch (e) {
          console.error('自动提交代理申请失败', e)
        }
      }

      const isAgent = res.userInfo.role === 'agent'
      wx.showToast({
        title: isAgent ? '欢迎回来' : '登录成功',
        icon: 'success'
      })

      setTimeout(() => {
        wx.switchTab({ url: '/pages/tabbar/home/index' })
      }, 1500)

    } catch (err) {
      wx.showToast({
        title: err.message || '微信登录失败',
        icon: 'none'
      })
      console.error('登录失败', err)
    } finally {
      this.setData({ wechatLoading: false })
    }
  }
})
