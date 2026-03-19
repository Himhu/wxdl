const { storeBindingsBehavior } = require('mobx-miniprogram-bindings')
const stationStore = require('../../store/station')

Component({
  behaviors: [storeBindingsBehavior],

  storeBindings: {
    store: stationStore,
    fields: ['currentSite', 'currentSiteInfo', 'sites'],
    actions: ['switchSite']
  },

  methods: {
    onSwitchTap() {
      const items = this.data.sites.map(site => site.name)
      wx.showActionSheet({
        itemList: items,
        success: (res) => {
          const selectedSite = this.data.sites[res.tapIndex]
          this.switchSite(selectedSite.id)
        }
      })
    }
  }
})
