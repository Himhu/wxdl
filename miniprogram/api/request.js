const config = require('../config/index')
const userStore = require('../store/user')
const stationStore = require('../store/station')

// 请求拦截器 — 从 mobx store 读取，避免每次请求调用 wx.getStorageSync
const requestInterceptor = (options) => {
  const token = userStore.token
  const currentSite = stationStore.currentSite || config.SITES[0].id

  options.header = {
    'Content-Type': 'application/json',
    'Authorization': token ? `Bearer ${token}` : '',
    'X-Site-Id': currentSite,
    ...options.header
  }

  return options
}

// 响应拦截器
const responseInterceptor = (response) => {
  const { statusCode, data } = response

  if (statusCode === 200) {
    if (data.code === 0 || data.success) {
      return Promise.resolve(data.data || data)
    } else {
      wx.showToast({
        title: data.message || '请求失败',
        icon: 'none'
      })
      return Promise.reject(data)
    }
  } else if (statusCode === 401) {
    wx.showToast({
      title: '登录已过期',
      icon: 'none'
    })
    userStore.clearLoginState()
    wx.reLaunch({
      url: '/pages/common/login/index'
    })
    return Promise.reject({ message: '未授权' })
  } else {
    wx.showToast({
      title: '网络错误',
      icon: 'none'
    })
    return Promise.reject({ message: '网络错误' })
  }
}

// 封装请求方法
const request = (options) => {
  return new Promise((resolve, reject) => {
    const requestOptions = requestInterceptor({
      url: config.API_BASE_URL + options.url,
      method: options.method || 'GET',
      data: options.data || {},
      header: options.header || {},
      timeout: options.timeout || 30000
    })

    wx.request({
      ...requestOptions,
      success: (res) => {
        responseInterceptor(res)
          .then(resolve)
          .catch(reject)
      },
      fail: (err) => {
        wx.showToast({
          title: '网络请求失败',
          icon: 'none'
        })
        reject(err)
      }
    })
  })
}

module.exports = {
  get: (url, data, options = {}) => request({ url, data, method: 'GET', ...options }),
  post: (url, data, options = {}) => request({ url, data, method: 'POST', ...options }),
  put: (url, data, options = {}) => request({ url, data, method: 'PUT', ...options }),
  delete: (url, data, options = {}) => request({ url, data, method: 'DELETE', ...options })
}
