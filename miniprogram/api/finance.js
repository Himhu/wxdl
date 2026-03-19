const request = require('./request')

// 财务相关接口
module.exports = {
  // 提交充值申请
  submitRecharge(data) {
    return request.post('/api/points/recharge', data)
  },

  // 获取充值记录
  getRechargeRecords(params) {
    return request.get('/api/points/ledger', params)
  },

  // 获取积分余额
  getBalance(agentId) {
    return request.get(`/api/points/balance/${agentId}`)
  }
}
