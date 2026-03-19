const config = require('../../config/index')

Component({
  properties: {
    type: {
      type: String,
      value: 'card' // card | recharge
    },
    status: {
      type: Number,
      value: 0
    }
  },

  data: {
    statusInfo: {}
  },

  observers: {
    'type, status': function(type, status) {
      let statusMap = {}
      if (type === 'card') {
        statusMap = config.CARD_STATUS
      } else if (type === 'recharge') {
        statusMap = config.RECHARGE_STATUS
      }

      const statusInfo = Object.values(statusMap).find(item => item.value === status) || {}
      this.setData({ statusInfo })
    }
  }
})
