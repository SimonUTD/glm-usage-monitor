// 错误处理工具函数
import { ElMessage, ElNotification } from 'element-plus'

// 错误类型枚举
export const ErrorTypes = {
  NETWORK: 'network',
  API: 'api',
  VALIDATION: 'validation',
  AUTH: 'auth',
  PERMISSION: 'permission',
  TIMEOUT: 'timeout',
  UNKNOWN: 'unknown'
}

// 错误级别枚举
export const ErrorLevels = {
  INFO: 'info',
  WARNING: 'warning',
  ERROR: 'error',
  SUCCESS: 'success'
}

// 错误处理配置
const errorConfig = {
  // 是否显示详细错误信息
  showDetails: true,
  // 是否记录错误日志
  logErrors: true,
  // 错误消息显示时长（毫秒）
  messageDuration: 5000,
  // 是否启用错误重试
  enableRetry: true,
  // 最大重试次数
  maxRetries: 3,
  // 重试延迟（毫秒）
  retryDelay: 1000
}

/**
 * 标准化错误对象
 * @param {Error|Object} error - 原始错误对象
 * @param {string} type - 错误类型
 * @param {string} level - 错误级别
 * @param {string} context - 错误上下文
 * @returns {Object} 标准化的错误对象
 */
export const createError = (error, type = ErrorTypes.UNKNOWN, level = ErrorLevels.ERROR, context = '') => {
  const standardError = {
    message: '',
    type,
    level,
    context,
    timestamp: new Date().toISOString(),
    originalError: error,
    code: null,
    details: null
  }

  // 根据错误类型提取信息
  if (error) {
    if (typeof error === 'string') {
      standardError.message = error
    } else if (error.response) {
      // HTTP 错误
      standardError.message = error.response.data?.message || error.response.statusText || '请求失败'
      standardError.code = error.response.status
      standardError.details = error.response.data
      standardError.type = ErrorTypes.API
    } else if (error.request) {
      // 网络错误
      standardError.message = '网络连接失败，请检查网络设置'
      standardError.type = ErrorTypes.NETWORK
    } else if (error.message) {
      // JavaScript 错误
      standardError.message = error.message
      standardError.code = error.code
      standardError.details = error.details
    }
  }

  return standardError
}

/**
 * 显示错误消息
 * @param {Object} error - 标准化的错误对象
 * @param {Object} options - 显示选项
 */
export const showError = (error, options = {}) => {
  const config = { ...errorConfig, ...options }
  
  if (!error.message) return

  // 根据错误级别选择显示方式
  switch (error.level) {
    case ErrorLevels.SUCCESS:
      ElMessage.success({
        message: error.message,
        duration: config.messageDuration,
        showClose: true
      })
      break
    case ErrorLevels.WARNING:
      ElMessage.warning({
        message: error.message,
        duration: config.messageDuration,
        showClose: true
      })
      break
    case ErrorLevels.INFO:
      ElMessage.info({
        message: error.message,
        duration: config.messageDuration,
        showClose: true
      })
      break
    case ErrorLevels.ERROR:
    default:
      if (config.showDetails && error.details) {
        ElNotification.error({
          title: '错误详情',
          message: `
            <div>
              <p><strong>错误信息:</strong> ${error.message}</p>
              <p><strong>错误类型:</strong> ${error.type}</p>
              <p><strong>错误上下文:</strong> ${error.context}</p>
              <p><strong>错误代码:</strong> ${error.code || 'N/A'}</p>
              <p><strong>时间:</strong> ${new Date(error.timestamp).toLocaleString()}</p>
            </div>
          `,
          duration: config.messageDuration,
          dangerouslyUseHTMLString: true
        })
      } else {
        ElMessage.error({
          message: error.message,
          duration: config.messageDuration,
          showClose: true
        })
      }
      break
  }
}

/**
 * 记录错误日志
 * @param {Object} error - 标准化的错误对象
 */
export const logError = (error) => {
  if (!errorConfig.logErrors) return

  const logData = {
    message: error.message,
    type: error.type,
    level: error.level,
    context: error.context,
    timestamp: error.timestamp,
    code: error.code,
    details: error.details,
    userAgent: navigator.userAgent,
    url: window.location.href
  }

  // 根据错误级别选择日志方法
  switch (error.level) {
    case ErrorLevels.ERROR:
      console.error('Error:', logData)
      break
    case ErrorLevels.WARNING:
      console.warn('Warning:', logData)
      break
    case ErrorLevels.INFO:
      console.info('Info:', logData)
      break
    default:
      console.log('Log:', logData)
  }

  // 可以在这里添加远程日志上报逻辑
  // sendErrorToServer(logData)
}

/**
 * 处理错误的完整流程
 * @param {Error|Object} error - 原始错误对象
 * @param {string} type - 错误类型
 * @param {string} level - 错误级别
 * @param {string} context - 错误上下文
 * @param {Object} options - 处理选项
 */
export const handleError = (error, type = ErrorTypes.UNKNOWN, level = ErrorLevels.ERROR, context = '', options = {}) => {
  const standardError = createError(error, type, level, context)
  
  // 记录日志
  logError(standardError)
  
  // 显示错误消息
  showError(standardError, options)
  
  return standardError
}

/**
 * 带重试机制的异步函数执行器
 * @param {Function} asyncFn - 异步函数
 * @param {Object} options - 重试选项
 * @returns {Promise} 执行结果
 */
export const withRetry = async (asyncFn, options = {}) => {
  const config = { ...errorConfig, ...options }
  let lastError = null
  
  for (let attempt = 1; attempt <= config.maxRetries; attempt++) {
    try {
      return await asyncFn()
    } catch (error) {
      lastError = error
      
      if (attempt === config.maxRetries) {
        // 最后一次尝试失败，处理错误
        throw handleError(error, ErrorTypes.UNKNOWN, ErrorLevels.ERROR, `重试${config.maxRetries}次后失败`)
      }
      
      // 等待后重试
      await new Promise(resolve => setTimeout(resolve, config.retryDelay))
      
      // 显示重试信息
      ElMessage.info({
        message: `操作失败，正在重试 (${attempt}/${config.maxRetries})...`,
        duration: 2000,
        showClose: false
      })
    }
  }
  
  throw lastError
}

/**
 * API 错误处理器
 * @param {Object} error - API 错误对象
 * @param {string} context - API 上下文
 */
export const handleApiError = (error, context = 'API调用') => {
  let type = ErrorTypes.API
  let level = ErrorLevels.ERROR
  let message = 'API调用失败'

  // 根据状态码确定错误类型和级别
  if (error.response) {
    const status = error.response.status
    
    switch (status) {
      case 400:
        type = ErrorTypes.VALIDATION
        message = '请求参数错误'
        break
      case 401:
        type = ErrorTypes.AUTH
        message = '未授权，请重新登录'
        break
      case 403:
        type = ErrorTypes.PERMISSION
        message = '权限不足'
        break
      case 404:
        message = '请求的资源不存在'
        break
      case 408:
        type = ErrorTypes.TIMEOUT
        message = '请求超时'
        break
      case 429:
        message = '请求过于频繁，请稍后重试'
        break
      case 500:
        message = '服务器内部错误'
        break
      case 502:
        message = '网关错误'
        break
      case 503:
        message = '服务暂时不可用'
        break
      default:
        message = `请求失败 (${status})`
    }
  } else if (error.request) {
    type = ErrorTypes.NETWORK
    message = '网络连接失败'
  }

  // 使用服务器返回的错误消息（如果有）
  if (error.response?.data?.message) {
    message = error.response.data.message
  }

  return handleError(error, type, level, context)
}

/**
 * 网络错误处理器
 * @param {Object} error - 网络错误对象
 * @param {string} context - 网络请求上下文
 */
export const handleNetworkError = (error, context = '网络请求') => {
  return handleError(error, ErrorTypes.NETWORK, ErrorLevels.ERROR, context)
}

/**
 * 验证错误处理器
 * @param {Object} error - 验证错误对象
 * @param {string} context - 验证上下文
 */
export const handleValidationError = (error, context = '数据验证') => {
  return handleError(error, ErrorTypes.VALIDATION, ErrorLevels.WARNING, context)
}

/**
 * 权限错误处理器
 * @param {Object} error - 权限错误对象
 * @param {string} context - 权限检查上下文
 */
export const handlePermissionError = (error, context = '权限检查') => {
  const standardError = handleError(error, ErrorTypes.PERMISSION, ErrorLevels.ERROR, context)
  
  // 权限错误可能需要跳转到登录页面
  if (error.response?.status === 401) {
    // 可以在这里添加跳转到登录页面的逻辑
    // router.push('/login')
  }
  
  return standardError
}

/**
 * 全局错误处理器
 * @param {Error} error - 全局错误对象
 */
export const handleGlobalError = (error) => {
  return handleError(error, ErrorTypes.UNKNOWN, ErrorLevels.ERROR, '全局错误')
}

// 设置全局错误处理
if (typeof window !== 'undefined') {
  window.addEventListener('error', (event) => {
    handleGlobalError(event.error || new Error(event.message))
  })
  
  window.addEventListener('unhandledrejection', (event) => {
    handleGlobalError(event.reason || new Error('Unhandled Promise Rejection'))
  })
}

export default {
  ErrorTypes,
  ErrorLevels,
  createError,
  showError,
  logError,
  handleError,
  withRetry,
  handleApiError,
  handleNetworkError,
  handleValidationError,
  handlePermissionError,
  handleGlobalError
}