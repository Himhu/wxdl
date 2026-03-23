const request = require('./request')

// 查询卡密列表
function getCards(params) {
  return request.get('/api/v1/cards', params)
}

// 查询卡密详情
function getCardDetail(id) {
  return request.get('/api/v1/cards/' + id)
}

// 销毁卡密
function destroyCard(id) {
  return request.delete('/api/v1/cards/' + id)
}

// 卡密统计
function getCardStats() {
  return request.get('/api/v1/cards/stats')
}

// 获取创建兑换码选项（面值列表）
function getCreateOptions() {
  return request.get('/api/v1/cards/create-options')
}

// 创建兑换码
function createCard(data) {
  return request.post('/api/v1/cards', data)
}

module.exports = {
  getCards,
  getCardDetail,
  destroyCard,
  getCardStats,
  getCreateOptions,
  createCard
}
