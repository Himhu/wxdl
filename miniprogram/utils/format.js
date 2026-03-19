// 日期格式化
function formatDate(date, format = 'YYYY-MM-DD HH:mm:ss') {
  if (!date) return ''

  const d = new Date(date)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hour = String(d.getHours()).padStart(2, '0')
  const minute = String(d.getMinutes()).padStart(2, '0')
  const second = String(d.getSeconds()).padStart(2, '0')

  return format
    .replace('YYYY', year)
    .replace('MM', month)
    .replace('DD', day)
    .replace('HH', hour)
    .replace('mm', minute)
    .replace('ss', second)
}

// 金额格式化
function formatAmount(amount) {
  if (amount === null || amount === undefined) return '0.00'
  return Number(amount).toFixed(2)
}

// 卡密格式化（隐藏中间部分）
function formatCardKey(cardKey) {
  if (!cardKey || cardKey.length < 8) return cardKey
  const start = cardKey.substring(0, 4)
  const end = cardKey.substring(cardKey.length - 4)
  return `${start}****${end}`
}

module.exports = {
  formatDate,
  formatAmount,
  formatCardKey
}
