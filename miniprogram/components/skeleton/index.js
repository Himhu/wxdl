Component({
  properties: {
    loading: { type: Boolean, value: true },
    count: { type: Number, value: 3 }
  },
  data: { rows: [] },
  observers: {
    count(val) {
      this.setData({ rows: new Array(val).fill(0) })
    }
  },
  lifetimes: {
    attached() {
      this.setData({ rows: new Array(this.data.count).fill(0) })
    }
  }
})
