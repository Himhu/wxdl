const { storeBindingsBehavior } = require('mobx-miniprogram-bindings')
const userStore = require('../../../store/user')
const authApi = require('../../../api/auth')

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
    nickname: ''
  },

  onShow() {
    if (this.data.isLogin) {
      wx.switchTab({ url: '/pages/tabbar/home/index' })
    }
  },

  // 用户选择头像
  onChooseAvatar(e) {
    const { avatarUrl } = e.detail || {}
    if (avatarUrl) {
      this.setData({ avatarUrl })
    }
  },

  // 昵称输入
  onNicknameInput(e) {
    this.setData({ nickname: e.detail.value || '' })
  },

  // 昵称失焦（兼容键盘上方快捷填入）
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

    this.setData({ wechatLoading: true })

    try {
      const code = await this.getWxLoginCode()

      const res = await authApi.wechatLogin({
        code,
        userInfo: {
          nickName: nickname,
          avatarUrl: avatarUrl
        }
      })

      this.setLoginState({
        token: res.token,
        userInfo: res.userInfo
      })

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
