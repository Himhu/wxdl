import request from './request'

/**
 * 卡密相关 API
 */

// 查询卡密列表
export function getCards(params) {
  return request({
    url: '/cards',
    method: 'GET',
    params
  })
}

// 查询卡密详情
export function getCardDetail(id) {
  return request({
    url: `/cards/${id}`,
    method: 'GET'
  })
}

// 销毁卡密
export function destroyCard(id) {
  return request({
    url: `/cards/${id}`,
    method: 'DELETE'
  })
}

// 卡密统计
export function getCardStats() {
  return request({
    url: '/cards/stats',
    method: 'GET'
  })
}
