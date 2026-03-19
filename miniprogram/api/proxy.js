const request = require('./request')

// 代理相关接口
module.exports = {
  // 登录
  login(data) {
    return request.post('/api/auth/login', data)
  },

  // 获取用户信息
  getUserInfo() {
    return request.get('/api/auth/userinfo')
  },

  // 获取下级代理列表
  getSubAgents(params) {
    return request.get('/api/proxy/list', params)
  },

  // 获取操作日志
  getOperationLogs(params) {
    return request.get('/api/audit/logs', params)
  }
}
