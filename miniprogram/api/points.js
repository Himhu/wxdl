import request from './request'

/**
 * 积分相关 API
 */

// 查询积分余额
export function getBalance() {
  return request({
    url: '/points/balance',
    method: 'GET'
  })
}

// 申请充值（向上级申请）
export function applyRecharge(data) {
  return request({
    url: '/points/recharge/apply',
    method: 'POST',
    data
  })
}

// 查看待审批的充值申请（上级查看）
export function getPendingRecharges(params) {
  return request({
    url: '/points/recharge/pending',
    method: 'GET',
    params
  })
}

// 审批通过
export function approveRecharge(id) {
  return request({
    url: `/points/recharge/approve/${id}`,
    method: 'POST'
  })
}

// 审批拒绝
export function rejectRecharge(id, data) {
  return request({
    url: `/points/recharge/reject/${id}`,
    method: 'POST',
    data
  })
}

// 查看充值历史
export function getRechargeHistory(params) {
  return request({
    url: '/points/recharge/history',
    method: 'GET',
    params
  })
}

// 查询积分流水
export function getPointsRecords(params) {
  return request({
    url: '/points/records',
    method: 'GET',
    params
  })
}

// 查询积分统计
export function getPointsStats() {
  return request({
    url: '/points/stats',
    method: 'GET'
  })
}
