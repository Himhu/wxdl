const authApi = require('../../api/auth')

const downloadFile = (url) => new Promise((resolve, reject) => {
  wx.downloadFile({
    url,
    success: resolve,
    fail: reject
  })
})

const getSetting = () => new Promise((resolve, reject) => {
  wx.getSetting({
    success: resolve,
    fail: reject
  })
})

const saveImageToPhotosAlbum = (filePath) => new Promise((resolve, reject) => {
  wx.saveImageToPhotosAlbum({
    filePath,
    success: resolve,
    fail: reject
  })
})

const nextTick = () => new Promise((resolve) => wx.nextTick(resolve))

Page({
  data: {
    inviteCode: '',
    displayInviteCode: '------',
    miniProgramCodeUrl: '',
    sharePath: '',
    posterUrl: '',
    posterWidth: 640,
    posterHeight: 800
  },

  async onLoad() {
    try {
      const res = await authApi.getCurrentInvite()
      const inviteCode = res.invite ? res.invite.code : ''
      this.setData({
        inviteCode,
        displayInviteCode: inviteCode || '------',
        sharePath: res.sharePath || '',
        miniProgramCodeUrl: authApi.getMiniProgramCodeUrl()
      })

      if (inviteCode) {
        await this.buildPoster()
      }
    } catch (err) {
      wx.showToast({ title: err.message || '获取邀请码失败', icon: 'none' })
    }
  },

  onCopyInviteCode() {
    if (!this.data.inviteCode) return
    wx.setClipboardData({ data: this.data.inviteCode })
  },

  drawRoundRect(ctx, x, y, width, height, radius) {
    ctx.beginPath()
    ctx.moveTo(x + radius, y)
    ctx.lineTo(x + width - radius, y)
    ctx.quadraticCurveTo(x + width, y, x + width, y + radius)
    ctx.lineTo(x + width, y + height - radius)
    ctx.quadraticCurveTo(x + width, y + height, x + width - radius, y + height)
    ctx.lineTo(x + radius, y + height)
    ctx.quadraticCurveTo(x, y + height, x, y + height - radius)
    ctx.lineTo(x, y + radius)
    ctx.quadraticCurveTo(x, y, x + radius, y)
    ctx.closePath()
  },

  async buildPoster() {
    if (this.data.posterUrl || !this.data.miniProgramCodeUrl) {
      return this.data.posterUrl
    }

    await nextTick()
    const downloadRes = await downloadFile(this.data.miniProgramCodeUrl)
    if (downloadRes.statusCode !== 200) {
      console.error('小程序码下载失败', {
        statusCode: downloadRes.statusCode,
        miniProgramCodeUrl: this.data.miniProgramCodeUrl,
        tempFilePath: downloadRes.tempFilePath,
      })
      throw new Error(`小程序码下载失败(${downloadRes.statusCode})`)
    }

    const { posterWidth, posterHeight, displayInviteCode } = this.data
    const query = wx.createSelectorQuery()

    return new Promise((resolve, reject) => {
      query.select('#posterCanvas').fields({ node: true, size: true }).exec((res) => {
        const canvasRes = res && res[0]
        if (!canvasRes || !canvasRes.node) {
          reject(new Error('画布初始化失败'))
          return
        }

        const canvas = canvasRes.node
        const ctx = canvas.getContext('2d')
        const dpr = wx.getWindowInfo ? wx.getWindowInfo().pixelRatio : 2

        canvas.width = posterWidth * dpr
        canvas.height = posterHeight * dpr
        ctx.scale(dpr, dpr)

        ctx.fillStyle = '#F3F6FA'
        ctx.fillRect(0, 0, posterWidth, posterHeight)

        const outerPadding = 32
        const cardRadius = 30
        const cardX = outerPadding
        const cardY = outerPadding
        const cardWidth = posterWidth - outerPadding * 2
        const cardHeight = posterHeight - outerPadding * 2

        this.drawRoundRect(ctx, cardX, cardY, cardWidth, cardHeight, cardRadius)
        ctx.fillStyle = '#FFFFFF'
        ctx.fill()

        const headerHeight = 150
        this.drawRoundRect(ctx, cardX, cardY, cardWidth, headerHeight, cardRadius)
        ctx.fillStyle = '#EAF1FF'
        ctx.fill()

        ctx.textAlign = 'left'
        ctx.fillStyle = '#1F2937'
        ctx.font = 'bold 36px sans-serif'
        ctx.fillText('邀请代理', 72, 102)

        ctx.fillStyle = '#5B6472'
        ctx.font = '22px sans-serif'
        ctx.fillText('扫码进入小程序，登录后自动关联邀请关系', 72, 142)

        ctx.fillStyle = '#111827'
        ctx.font = 'bold 26px sans-serif'
        ctx.fillText('邀请码', 72, 236)

        ctx.fillStyle = '#2F6BFF'
        ctx.font = 'bold 54px sans-serif'
        ctx.fillText(displayInviteCode || '------', 72, 306)

        ctx.fillStyle = '#6B7280'
        ctx.font = '21px sans-serif'
        ctx.fillText('保存海报到相册，发给好友后可直接扫码加入。', 72, 352)

        const img = canvas.createImage()
        img.onload = () => {
          const codeSize = 260
          const codeX = (posterWidth - codeSize) / 2
          const codeY = 410
          ctx.drawImage(img, codeX, codeY, codeSize, codeSize)

          ctx.fillStyle = '#4B5563'
          ctx.font = '20px sans-serif'
          ctx.textAlign = 'center'
          ctx.fillText('长按识别小程序码', posterWidth / 2, 708)

          ctx.fillStyle = '#9CA3AF'
          ctx.font = '18px sans-serif'
          ctx.fillText('加入后将自动绑定邀请关系', posterWidth / 2, 740)
          ctx.textAlign = 'left'

          wx.canvasToTempFilePath({
            canvas,
            x: 0,
            y: 0,
            width: posterWidth,
            height: posterHeight,
            destWidth: posterWidth * 2,
            destHeight: posterHeight * 2,
            fileType: 'png',
            success: ({ tempFilePath }) => {
              this.setData({ posterUrl: tempFilePath })
              resolve(tempFilePath)
            },
            fail: reject
          }, this)
        }
        img.onerror = reject
        img.src = downloadRes.tempFilePath
      })
    })
  },

  async onSavePoster() {
    try {
      wx.showLoading({ title: '海报生成中...' })
      const setting = await getSetting()
      if (setting.authSetting['scope.writePhotosAlbum'] === false) {
        wx.hideLoading()
        wx.showModal({
          title: '需要授权',
          content: '请授权保存到相册',
          success: (res) => {
            if (res.confirm) wx.openSetting()
          }
        })
        return
      }

      const posterUrl = await this.buildPoster()
      if (!posterUrl) {
        wx.hideLoading()
        wx.showToast({ title: '海报未生成', icon: 'none' })
        return
      }

      await saveImageToPhotosAlbum(posterUrl)
      wx.hideLoading()
      wx.showToast({ title: '保存成功', icon: 'success' })
    } catch (err) {
      wx.hideLoading()
      console.error('保存邀请海报失败', err)
      const msg = (err && err.errMsg) || (err && err.message) || ''
      if (msg.includes('auth deny') || msg.includes('auth denied')) {
        wx.showModal({
          title: '需要授权',
          content: '请在设置中开启保存到相册权限',
          success: (res) => {
            if (res.confirm) wx.openSetting()
          }
        })
        return
      }
      if (!msg.includes('cancel')) {
        wx.showToast({ title: '保存失败', icon: 'none' })
      }
    }
  },

  onShareAppMessage() {
    return {
      title: '邀请你加入代理平台',
      path: this.data.sharePath || `/pages/common/login/index?inviteCode=${this.data.inviteCode}`
    }
  }
})
