import request from './request'

/**
 * 代理商相关 API
 */

// 创建下级代理
export function createAgent(data) {
  return request({
    url: '/agents',
    method: 'POST',
    data
  })
}

// 查询下级代理列表
export function getAgents(params) {
  return request({
    url: '/agents',
    method: 'GET',
    params
  })
}

// 查询代理详情
export function getAgentDetail(id) {
  return request({
    url: `/agents/${id}`,
    method: 'GET'
  })
}

// 更新代理信息
export function updateAgent(id, data) {
  return request({
    url: `/agents/${id}`,
    method: 'PUT',
    data
  })
}

// 禁用/启用代理
export function updateAgentStatus(id, status) {
  return request({
    url: `/agents/${id}/status`,
    method: 'PUT',
    data: { status }
  })
}
