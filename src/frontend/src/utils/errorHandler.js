// 前端错误处理工具
import { ElMessage, ElNotification } from 'element-plus'

// 错误类型常量
export const ErrorTypes = {
  NETWORK: 'NETWORK_ERROR',
  API: 'API_ERROR',
  DATABASE: 'DATABASE_ERROR',
  VALIDATION: 'VALIDATION_ERROR',
  AUTH: 'AUTH_ERROR',
  SYNC: 'SYNC_ERROR',
  INTERNAL: 'INTERNAL_ERROR',
  NOT_FOUND: 'NOT_FOUND_ERROR'
}

// 错误级别常量
export const ErrorLevels = {
  INFO: 'info',
  WARNING: 'warning',
  ERROR: 'error',
  CRITICAL: 'critical'
}

// 用户友好的错误消息映射
const ErrorMessages = {
  [ErrorTypes.NETWORK]: {
    title: '网络连接错误',
    default: '网络连接异常，请检查网络设置后重试',
    actions: ['检查网络连接', '稍后重试']
  },
  [ErrorTypes.API]: {
    title: 'API调用失败',
    default: '服务接口调用失败，请稍后重试',
    actions: ['检查API配置', '联系技术支持']
  },
  [ErrorTypes.DATABASE]: {
    title: '数据库错误',
    default: '数据操作失败，请重试或联系技术支持',
    actions: ['重启应用', '联系技术支持']
  },
  [ErrorTypes.VALIDATION]: {
    title: '参数验证错误',
    default: '输入参数有误，请检查后重试',
    actions: ['检查输入内容', '查看帮助文档']
  },
  [ErrorTypes.AUTH]: {
    title: '认证失败',
    default: '身份验证失败，请检查登录状态',
    actions: ['重新登录', '检查权限']
  },
  [ErrorTypes.SYNC]: {
    title: '数据同步失败',
    default: '数据同步过程中出现错误，请重试',
    actions: ['检查网络连接', '查看同步历史']
  },
  [ErrorTypes.INTERNAL]: {
    title: '系统内部错误',
    default: '系统出现未知错误，请联系技术支持',
    actions: ['重启应用', '联系技术支持']
  },
  [ErrorTypes.NOT_FOUND]: {
    title: '资源未找到',
    default: '请求的资源不存在，请检查后重试',
    actions: ['刷新页面', '检查参数']
  }
}

// 错误恢复策略
export const ErrorRecoveryActions = {
  RETRY: 'retry',
  REFRESH: 'refresh',
  RESTART: 'restart',
  CONTACT_SUPPORT: 'contact_support',
  CHECK_CONFIG: 'check_config',
  IGNORE: 'ignore'
}

/**
 * 前端错误处理器类
 */
export class FrontendErrorHandler {
  constructor(options = {}) {
    this.options = {
      enableNotification: options.enableNotification ?? true,
      enableConsoleLog: options.enableConsoleLog ?? true,
      enableErrorReporting: options.enableErrorReporting ?? false,
      maxRetries: options.maxRetries ?? 3,
      retryDelay: options.retryDelay ?? 1000,
      ...options
    }

    this.errorQueue = []
    this.retryCount = new Map()
  }

  /**
   * 处理错误的主要入口方法
   * @param {Error|Object} error - 错误对象
   * @param {Object} context - 错误上下文信息
   * @param {Object} options - 处理选项
   */
  handleError(error, context = {}, options = {}) {
    const errorInfo = this.parseError(error, context)

    // 记录错误日志
    if (this.options.enableConsoleLog) {
      this.logError(errorInfo)
    }

    // 添加到错误队列（用于重试等）
    this.errorQueue.push(errorInfo)

    // 显示用户通知
    if (this.options.enableNotification && options.showNotification !== false) {
      this.showErrorNotification(errorInfo, options)
    }

    // 错误上报（如果启用）
    if (this.options.enableErrorReporting) {
      this.reportError(errorInfo)
    }

    return errorInfo
  }

  /**
   * 解析错误对象
   * @param {Error|Object} error - 原始错误
   * @param {Object} context - 上下文信息
   * @returns {Object} 解析后的错误信息
   */
  parseError(error, context = {}) {
    let errorInfo = {
      id: this.generateErrorId(),
      timestamp: new Date().toISOString(),
      type: ErrorTypes.INTERNAL,
      level: ErrorLevels.ERROR,
      message: '未知错误',
      details: '',
      code: '',
      context: context,
      stack: null,
      retryable: false,
      originalError: error
    }

    if (typeof error === 'string') {
      errorInfo.message = error
    } else if (error && typeof error === 'object') {
      // 处理Go后端返回的结构化错误
      if (error.type && error.code) {
        errorInfo.type = error.type
        errorInfo.code = error.code
        errorInfo.message = error.message || error.details || errorInfo.message
        errorInfo.details = error.details || ''
        errorInfo.context = { ...errorInfo.context, ...error.context }
      } else {
        // 处理普通JavaScript错误
        errorInfo.message = error.message || errorInfo.message
        errorInfo.stack = error.stack
        errorInfo.code = error.code || ''
      }

      // 根据错误内容推断错误类型
      errorInfo.type = this.inferErrorType(errorInfo)
      errorInfo.level = this.inferErrorLevel(errorInfo)
      errorInfo.retryable = this.isRetryable(errorInfo)
    }

    return errorInfo
  }

  /**
   * 根据错误内容推断错误类型
   */
  inferErrorType(errorInfo) {
    const { message, type, code } = errorInfo

    // 如果已经明确指定了类型，直接使用
    if (type && Object.values(ErrorTypes).includes(type)) {
      return type
    }

    const lowerMessage = message.toLowerCase()

    if (lowerMessage.includes('network') || lowerMessage.includes('connection') ||
        lowerMessage.includes('timeout') || lowerMessage.includes('fetch')) {
      return ErrorTypes.NETWORK
    }

    if (lowerMessage.includes('unauthorized') || lowerMessage.includes('token') ||
        lowerMessage.includes('认证') || lowerMessage.includes('权限')) {
      return ErrorTypes.AUTH
    }

    if (lowerMessage.includes('database') || lowerMessage.includes('db') ||
        lowerMessage.includes('数据库') || lowerMessage.includes('查询')) {
      return ErrorTypes.DATABASE
    }

    if (lowerMessage.includes('validation') || lowerMessage.includes('invalid') ||
        lowerMessage.includes('验证') || lowerMessage.includes('参数')) {
      return ErrorTypes.VALIDATION
    }

    if (lowerMessage.includes('sync') || lowerMessage.includes('同步') ||
        code && code.toString().includes('SYNC')) {
      return ErrorTypes.SYNC
    }

    if (lowerMessage.includes('not found') || lowerMessage.includes('不存在') ||
        lowerMessage.includes('未找到')) {
      return ErrorTypes.NOT_FOUND
    }

    if (lowerMessage.includes('api') || code && code.toString().includes('API')) {
      return ErrorTypes.API
    }

    return ErrorTypes.INTERNAL
  }

  /**
   * 推断错误级别
   */
  inferErrorLevel(errorInfo) {
    const { type, message } = errorInfo

    switch (type) {
      case ErrorTypes.NETWORK:
      case ErrorTypes.SYNC:
        return ErrorLevels.WARNING
      case ErrorTypes.AUTH:
      case ErrorTypes.DATABASE:
      case ErrorTypes.INTERNAL:
        return ErrorLevels.ERROR
      case ErrorTypes.VALIDATION:
      case ErrorTypes.NOT_FOUND:
        return ErrorLevels.INFO
      default:
        return ErrorLevels.ERROR
    }
  }

  /**
   * 判断错误是否可重试
   */
  isRetryable(errorInfo) {
    const { type, code, message } = errorInfo

    // 网络错误通常可重试
    if (type === ErrorTypes.NETWORK) {
      return true
    }

    // API超时和限流错误可重试
    if (type === ErrorTypes.API) {
      return code?.includes('TIMEOUT') || code?.includes('RATE_LIMIT') ||
             message.toLowerCase().includes('timeout')
    }

    // 同步错误通常可重试
    if (type === ErrorTypes.SYNC) {
      return !code?.includes('INVALID_TOKEN') && !code?.includes('NO_TOKEN')
    }

    return false
  }

  /**
   * 显示错误通知
   */
  showErrorNotification(errorInfo, options = {}) {
    const errorTypeConfig = ErrorMessages[errorInfo.type] || ErrorMessages[ErrorTypes.INTERNAL]

    const notificationConfig = {
      title: errorTypeConfig.title,
      message: this.formatErrorMessage(errorInfo),
      type: errorInfo.level === ErrorLevels.CRITICAL ? 'error' : errorInfo.level,
      duration: errorInfo.level === ErrorLevels.CRITICAL ? 0 : 5000,
      showClose: true,
      ...options.notification
    }

    // 根据错误级别选择通知方式
    if (errorInfo.level === ErrorLevels.CRITICAL || errorInfo.type === ErrorTypes.AUTH) {
      // 严重错误使用Notification，需要用户确认
      ElNotification({
        ...notificationConfig,
        dangerouslyUseHTMLString: true,
        customClass: 'error-notification'
      })
    } else {
      // 普通错误使用Message
      ElMessage({
        ...notificationConfig,
        grouping: true,
        customClass: 'error-message'
      })
    }
  }

  /**
   * 格式化错误消息
   */
  formatErrorMessage(errorInfo) {
    const errorTypeConfig = ErrorMessages[errorInfo.type] || ErrorMessages[ErrorTypes.INTERNAL]

    let message = errorInfo.message || errorTypeConfig.default

    // 如果有详细信息，添加到消息中
    if (errorInfo.details && errorInfo.details !== errorInfo.message) {
      message += `<br/><small>详细信息: ${errorInfo.details}</small>`
    }

    // 如果有错误代码，添加到消息中
    if (errorInfo.code) {
      message += `<br/><small>错误代码: ${errorInfo.code}</small>`
    }

    // 添加建议的操作
    if (errorTypeConfig.actions && errorTypeConfig.actions.length > 0) {
      const actions = errorTypeConfig.actions.map(action => `<span class="error-action">${action}</span>`).join(' | ')
      message += `<br/><small>建议: ${actions}</small>`
    }

    return message
  }

  /**
   * 记录错误日志
   */
  logError(errorInfo) {
    const logData = {
      id: errorInfo.id,
      type: errorInfo.type,
      level: errorInfo.level,
      message: errorInfo.message,
      code: errorInfo.code,
      context: errorInfo.context,
      timestamp: errorInfo.timestamp
    }

    switch (errorInfo.level) {
      case ErrorLevels.CRITICAL:
        console.error('🚨 CRITICAL ERROR:', logData)
        break
      case ErrorLevels.ERROR:
        console.error('❌ ERROR:', logData)
        break
      case ErrorLevels.WARNING:
        console.warn('⚠️ WARNING:', logData)
        break
      case ErrorLevels.INFO:
        console.info('ℹ️ INFO:', logData)
        break
    }

    if (errorInfo.stack) {
      console.groupCollapsed(`${errorInfo.type} Stack Trace`)
      console.error(errorInfo.stack)
      console.groupEnd()
    }
  }

  /**
   * 错误上报
   */
  reportError(errorInfo) {
    // 这里可以集成错误监控服务，如Sentry
    try {
      const reportData = {
        ...errorInfo,
        userAgent: navigator.userAgent,
        url: window.location.href,
        timestamp: new Date().toISOString()
      }

      // 示例：发送到错误收集服务
      // fetch('/api/errors', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify(reportData)
      // }).catch(() => {
      //   // 忽略上报失败，避免无限循环
      // })

      console.log('Error reported:', reportData)
    } catch (e) {
      console.warn('Failed to report error:', e)
    }
  }

  /**
   * 重试机制
   */
  async retry(errorId, retryFunction) {
    const errorInfo = this.errorQueue.find(err => err.id === errorId)
    if (!errorInfo || !errorInfo.retryable) {
      return false
    }

    const currentRetries = this.retryCount.get(errorId) || 0
    if (currentRetries >= this.options.maxRetries) {
      return false
    }

    try {
      this.retryCount.set(errorId, currentRetries + 1)

      // 延迟重试
      await this.delay(this.options.retryDelay * Math.pow(2, currentRetries))

      const result = await retryFunction()
      this.retryCount.delete(errorId)
      return result
    } catch (error) {
      return this.retry(errorId, retryFunction)
    }
  }

  /**
   * 延迟函数
   */
  delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms))
  }

  /**
   * 生成错误ID
   */
  generateErrorId() {
    return 'err_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
  }

  /**
   * 清理错误队列
   */
  clearErrors() {
    this.errorQueue = []
    this.retryCount.clear()
  }

  /**
   * 获取错误统计
   */
  getErrorStats() {
    const stats = {
      total: this.errorQueue.length,
      byType: {},
      byLevel: {}
    }

    this.errorQueue.forEach(error => {
      stats.byType[error.type] = (stats.byType[error.type] || 0) + 1
      stats.byLevel[error.level] = (stats.byLevel[error.level] || 0) + 1
    })

    return stats
  }
}

// 创建默认错误处理器实例
export const defaultErrorHandler = new FrontendErrorHandler()

// 便捷方法
export const handleError = (error, context, options) => {
  return defaultErrorHandler.handleError(error, context, options)
}

export const retryError = (errorId, retryFunction) => {
  return defaultErrorHandler.retry(errorId, retryFunction)
}

// Vue插件形式的错误处理器
export const ErrorHandlerPlugin = {
  install(app, options = {}) {
    const errorHandler = new FrontendErrorHandler(options)

    // 全局属性
    app.config.globalProperties.$errorHandler = errorHandler
    app.config.globalProperties.$handleError = handleError

    // 全局错误处理
    app.config.errorHandler = (err, instance, info) => {
      errorHandler.handleError(err, {
        component: instance?.$options?.name || 'Unknown',
        info: info
      })
    }

    // 未捕获的Promise错误
    window.addEventListener('unhandledrejection', (event) => {
      errorHandler.handleError(event.reason, {
        type: 'unhandledrejection'
      })
    })
  }
}

// CSS样式（需要在main.js中导入或在组件中使用）
export const errorStyles = `
.error-notification {
  border-left: 4px solid #f56c6c;
}

.error-message {
  border-left: 4px solid #e6a23c;
}

.error-action {
  color: #409eff;
  cursor: pointer;
  font-weight: 500;
}

.error-action:hover {
  text-decoration: underline;
}
`