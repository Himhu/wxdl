const http = require('./request')

module.exports = {
  fetchLegacyBalance(username, password) {
    return http.post('/api/v1/user/data-transfer/legacy/balance', {
      username,
      password
    })
  },

  confirmLegacyTransfer(username, password) {
    return http.post('/api/v1/user/data-transfer/legacy/confirm', {
      username,
      password
    })
  }
}
