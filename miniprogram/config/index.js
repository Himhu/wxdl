// 全局配置
module.exports = {
  // API 基础地址（后续对接时修改）
  API_BASE_URL: 'http://192.168.1.20:8080',

  // 卡密状态枚举
  CARD_STATUS: {
    UNUSED: { value: 1, label: '未使用', color: '#2F57C8', bgColor: '#EDF3FF', borderColor: '#C9DBFF' },
    USED: { value: 2, label: '已使用', color: '#067647', bgColor: '#ECFDF3', borderColor: '#ABEFC6' },
    DESTROYED: { value: 3, label: '已销毁', color: '#B42318', bgColor: '#FEF2F2', borderColor: '#FDA29B' }
  },

  // 充值状态枚举
  RECHARGE_STATUS: {
    PENDING: { value: 0, label: '审核中', color: '#B54708', bgColor: '#FFF7E8', borderColor: '#F7D79B' },
    APPROVED: { value: 1, label: '已通过', color: '#067647', bgColor: '#ECFDF3', borderColor: '#ABEFC6' },
    REJECTED: { value: -1, label: '已驳回', color: '#B42318', bgColor: '#FEF2F2', borderColor: '#FDA29B' }
  },

  // 用户角色枚举
  USER_ROLE: {
    USER: { value: 'user', label: '普通用户' },
    AGENT: { value: 'agent', label: '代理商' }
  },
}
