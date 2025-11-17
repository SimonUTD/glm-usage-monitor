import { ref } from 'vue'

/**
 * 通用API调用包装器Composable
 * 用于消除重复的API调用模式，遵循DRY原则
 * 更新支持Wails API的错误处理模式
 *
 * @param {Function} apiCall - API调用函数
 * @param {Object} options - 配置选项
 * @param {*} options.defaultValue - 失败时的默认值
 * @param {string} options.errorMessage - 自定义错误消息前缀
 * @param {Function} options.dataTransformer - 数据转换函数
 * @param {boolean} options.isWailsCall - 是否为Wails API调用，默认为true
 * @returns {Object} { data, loading, error, execute }
 */
export function useApiCall(apiCall, options = {}) {
  const {
    defaultValue = null,
    errorMessage = '',
    dataTransformer = null,
    isWailsCall = true
  } = options

  const data = ref(null)
  const loading = ref(false)
  const error = ref(null)

  /**
   * 执行API调用
   * @param {...any} args - 传递给API调用的参数
   * @returns {Promise<any>} API调用结果
   */
  const execute = async (...args) => {
    loading.value = true
    error.value = null

    try {
      const result = await apiCall(...args)

      // Wails API 调用已经包装了响应格式，直接检查 success 字段
      if (isWailsCall) {
        if (result && result.success) {
          // 如果有数据转换函数，应用转换
          data.value = dataTransformer ? dataTransformer(result.data) : result.data
        } else {
          // API调用失败，使用默认值
          data.value = defaultValue
          const errorMsg = result?.message || '操作失败'
          error.value = new Error(errorMsg)
          console.error(`${errorMessage}操作失败:`, errorMsg)
        }
      } else {
        // 非 Wails 调用的兼容处理
        if (result && result.success && result.data !== undefined) {
          data.value = dataTransformer ? dataTransformer(result.data) : result.data
        } else {
          data.value = defaultValue
          console.error(`${errorMessage}API返回失败:`, result?.message || '未知错误')
        }
      }

      return result
    } catch (err) {
      // 捕获异常，设置默认值和错误信息
      data.value = defaultValue
      error.value = err
      console.error(`${errorMessage}API调用异常:`, err)

      // 返回标准化的错误响应
      return {
        success: false,
        message: err.message || '操作失败',
        error: err
      }
    } finally {
      loading.value = false
    }
  }

  return {
    data,
    loading,
    error,
    execute
  }
}

/**
 * 简单的单值API调用包装器
 * 用于获取单个数值类型的API响应，优化支持Wails API
 *
 * @param {Function} apiCall - API调用函数
 * @param {Object} options - 配置选项
 * @param {*} options.defaultValue - 失败时的默认值，默认为0
 * @param {string} options.errorMessage - 错误消息前缀
 * @param {string} options.dataKey - 从响应数据中提取的键名，默认为'used'
 * @param {boolean} options.isWailsCall - 是否为Wails API调用
 * @returns {Function} 返回一个异步函数，直接执行API调用
 */
export function createSimpleApiHandler(apiCall, options = {}) {
  const {
    defaultValue = 0,
    errorMessage = '',
    dataKey = 'used',
    isWailsCall = true
  } = options

  return async (...args) => {
    try {
      const result = await apiCall(...args)

      if (isWailsCall) {
        // Wails API 调用已经标准化响应格式
        if (result && result.success && result.data) {
          return typeof result.data === 'object'
            ? result.data[dataKey] || defaultValue
            : result.data || defaultValue
        }
      } else {
        // 传统 HTTP API 格式
        if (result && result.success && result.data) {
          return result.data[dataKey] || defaultValue
        }
      }

      return defaultValue
    } catch (error) {
      console.error(`${errorMessage}失败:`, error)
      return defaultValue
    }
  }
}

/**
 * 对象型API调用包装器
 * 用于获取包含多个字段的API响应，优化支持Wails API
 *
 * @param {Function} apiCall - API调用函数
 * @param {Object} options - 配置选项
 * @param {Object} options.defaultValue - 失败时的默认值对象
 * @param {string} options.errorMessage - 错误消息前缀
 * @param {boolean} options.isWailsCall - 是否为Wails API调用
 * @returns {Function} 返回一个异步函数，直接执行API调用
 */
export function createObjectApiHandler(apiCall, options = {}) {
  const {
    defaultValue = {},
    errorMessage = '',
    isWailsCall = true
  } = options

  return async (...args) => {
    try {
      const result = await apiCall(...args)

      if (isWailsCall) {
        // Wails API 调用
        if (result && result.success) {
          return result.data || defaultValue
        }
      } else {
        // 传统 HTTP API 格式
        if (result && result.success && result.data) {
          return result.data
        }
      }

      return defaultValue
    } catch (error) {
      console.error(`${errorMessage}失败:`, error)
      return defaultValue
    }
  }
}

/**
 * 产品分布数据API调用包装器
 * 专门处理产品分布数据格式转换，优化支持Wails API
 *
 * @param {Function} apiCall - API调用函数
 * @param {string} errorMessage - 错误消息前缀
 * @param {boolean} options.isWailsCall - 是否为Wails API调用
 * @returns {Function} 返回一个异步函数，执行API调用并转换数据格式
 */
export function createProductDistributionHandler(apiCall, errorMessage = '', options = {}) {
  const { isWailsCall = true } = options

  return async (...args) => {
    try {
      const result = await apiCall(...args)

      let data = null
      if (isWailsCall) {
        data = result?.success ? result.data : null
      } else {
        data = result?.success ? result.data : null
      }

      if (data && data.length > 0) {
        // 兼容不同的数据格式
        const productNames = data.map(item =>
          item.productName || item.model || item.name || '未知'
        )
        const productValues = data.map(item =>
          item.totalUsage || item.usage || item.count || 0
        )

        return {
          productNames,
          productValues
        }
      } else {
        // 如果没有数据，使用空数组
        return {
          productNames: [],
          productValues: []
        }
      }
    } catch (error) {
      console.error(`${errorMessage}失败:`, error)
      return {
        productNames: [],
        productValues: []
      }
    }
  }
}

/**
 * 时间序列数据API调用包装器
 * 用于处理图表时间序列数据，优化支持Wails API
 *
 * @param {Function} apiCall - API调用函数
 * @param {string} errorMessage - 错误消息前缀
 * @param {boolean} options.isWailsCall - 是否为Wails API调用
 * @returns {Function} 返回一个异步函数，执行API调用并转换数据格式
 */
export function createTimeSeriesHandler(apiCall, errorMessage = '', options = {}) {
  const { isWailsCall = true } = options

  return async (...args) => {
    try {
      const result = await apiCall(...args)

      let data = null
      if (isWailsCall) {
        data = result?.success ? result.data : null
      } else {
        data = result?.success ? result.data : null
      }

      if (data) {
        return {
          labels: data.labels || data.time || data.dates || [],
          callCountData: data.callCountData || data.calls || data.counts || [],
          tokenData: data.tokenData || data.tokens || []
        }
      } else {
        return {
          labels: [],
          callCountData: [],
          tokenData: []
        }
      }
    } catch (error) {
      console.error(`${errorMessage}失败:`, error)
      return {
        labels: [],
        callCountData: [],
        tokenData: []
      }
    }
  }
}