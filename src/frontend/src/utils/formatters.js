/**
 * 格式化工具函数
 * 提取公共的格式化逻辑，避免在组件中重复代码
 */

/**
 * 格式化数字显示
 * @param {number} num - 要格式化的数字
 * @returns {string} 格式化后的字符串
 */
export const formatNumber = (num) => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'k'
  }
  return num.toString()
}

/**
 * 格式化调用次数
 * @param {number} value - 调用次数
 * @returns {string} 格式化后的字符串
 */
export const formatCallCount = (value) => {
  if (value >= 1000) {
    return (value / 1000).toFixed(1) + 'k'
  }
  return value.toString()
}

/**
 * 格式化Token数量
 * @param {number} value - Token数量
 * @returns {string} 格式化后的字符串
 */
export const formatTokenNumber = (value) => {
  if (value >= 1000000) {
    return (value / 1000000).toFixed(1) + 'M'
  }
  if (value >= 1000) {
    return (value / 1000).toFixed(1) + 'k'
  }
  return value.toString()
}

/**
 * 格式化金额显示
 * @param {number} value - 金额
 * @returns {string} 格式化后的金额字符串
 */
export const formatCurrency = (value) => {
  return '¥' + value.toFixed(2)
}

/**
 * 格式化统计数据
 * @param {number} value - 统计值
 * @param {string} type - 数据类型 ('call', 'token', 'cost')
 * @returns {string} 格式化后的字符串
 */
export const formatStatsValue = (value, type) => {
  if (type === 'cost') {
    return formatCurrency(value)
  }
  if (type === 'token') {
    return formatTokenNumber(value)
  }
  if (type === 'call') {
    return formatCallCount(value)
  }
  return value.toLocaleString()
}

/**
 * 获取当前月份字符串
 * @returns {string} 当前月份 (YYYY-MM格式)
 */
export const getCurrentMonth = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  return `${year}-${month}`
}

/**
 * 格式化时间显示
 * @param {string} timeStr - 时间字符串
 * @returns {string} 格式化后的时间字符串
 */
export const formatTime = (timeStr) => {
  if (!timeStr) return '--'
  // 后端已返回本地时间格式，直接使用
  return timeStr
}

/**
 * 获取进度条颜色
 * @param {number} percentage - 百分比
 * @returns {string} 颜色值
 */
export const getProgressColor = (percentage) => {
  if (percentage >= 90) return '#E74C3C'      // 红色 (危险)
  if (percentage >= 70) return '#F39C12'      // 橙色 (警告)
  if (percentage >= 50) return '#4D6782'      // 主色 (正常)
  return '#A8C686'                            // 绿色 (良好)
}

/**
 * 获取进度条动画时长
 * @param {number} percentage - 百分比
 * @returns {number} 动画时长(秒)
 */
export const getProgressDuration = (percentage) => {
  if (percentage >= 90) return 14      // 很快
  if (percentage >= 70) return 16      // 快
  if (percentage >= 50) return 18      // 正常
  return 20                            // 慢
}

/**
 * 防抖函数
 * @param {Function} fn - 要防抖的函数
 * @param {number} delay - 延迟时间(毫秒)
 * @returns {Function} 防抖后的函数
 */
export const debounce = (fn, delay = 300) => {
  let timer = null
  return function(...args) {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => fn.apply(this, args), delay)
  }
}

/**
 * 节流函数
 * @param {Function} fn - 要节流的函数
 * @param {number} delay - 延迟时间(毫秒)
 * @returns {Function} 节流后的函数
 */
export const throttle = (fn, delay = 300) => {
  let lastCall = 0
  return function(...args) {
    const now = Date.now()
    if (now - lastCall >= delay) {
      lastCall = now
      return fn.apply(this, args)
    }
  }
}

/**
 * 检查元素是否在视窗内
 * @param {Element} element - DOM元素
 * @returns {boolean} 是否可见
 */
export const isElementInViewport = (element) => {
  const rect = element.getBoundingClientRect()
  return (
    rect.top >= 0 &&
    rect.left >= 0 &&
    rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
    rect.right <= (window.innerWidth || document.documentElement.clientWidth)
  )
}

/**
 * 产品名称常量
 */
export const PRODUCT_NAMES = [
  'glm-4.5-air 0-32k 0-0.2k',
  'glm-4.5-air 0-32k 0.2k+',
  'glm-4.5-air 32-128k',
  'glm-4.6 0-32k 0-0.2k',
  'glm-4.6 0-32k 0.2k+',
  'glm-4.6 32-200k'
]