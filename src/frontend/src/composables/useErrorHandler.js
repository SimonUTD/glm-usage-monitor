import { ref, computed } from 'vue'
import { defaultErrorHandler, ErrorLevels, ErrorTypes } from '@/utils/errorHandler'
import { ElMessageBox } from 'element-plus'

/**
 * 错误处理 Composable
 * 提供响应式的错误处理功能
 */
export function useErrorHandler(options = {}) {
  // 错误状态
  const currentError = ref(null)
  const errorHistory = ref([])
  const isLoading = ref(false)
  const retryCount = ref(0)

  // 配置选项
  const config = {
    maxRetries: 3,
    showDetailedErrors: false,
    autoRetry: false,
    ...options
  }

  // 计算属性
  const hasError = computed(() => currentError.value !== null)
  const errorCount = computed(() => errorHistory.value.length)
  const canRetry = computed(() => {
    return hasError.value &&
           currentError.value?.retryable &&
           retryCount.value < config.maxRetries
  })
  const errorStats = computed(() => {
    return defaultErrorHandler.getErrorStats()
  })

  /**
   * 处理错误
   * @param {Error|Object} error - 错误对象
   * @param {Object} context - 错误上下文
   * @param {Object} options - 处理选项
   */
  const handleError = (error, context = {}, options = {}) => {
    const errorInfo = defaultErrorHandler.handleError(error, context, options)

    currentError.value = errorInfo
    errorHistory.value.unshift(errorInfo)

    // 限制历史记录数量
    if (errorHistory.value.length > 100) {
      errorHistory.value = errorHistory.value.slice(0, 100)
    }

    return errorInfo
  }

  /**
   * 清除当前错误
   */
  const clearError = () => {
    currentError.value = null
    retryCount.value = 0
  }

  /**
   * 重试操作
   * @param {Function} retryFunction - 重试函数
   */
  const retry = async (retryFunction) => {
    if (!canRetry.value || !retryFunction) {
      return false
    }

    isLoading.value = true
    retryCount.value++

    try {
      const result = await retryFunction()
      clearError()
      return result
    } catch (error) {
      // 处理重试失败
      const errorInfo = handleError(error, {
        operation: 'retry',
        attempt: retryCount.value,
        maxAttempts: config.maxRetries
      })

      if (retryCount.value >= config.maxRetries) {
        // 达到最大重试次数，询问用户
        await showMaxRetriesDialog(errorInfo)
      }

      return false
    } finally {
      isLoading.value = false
    }
  }

  /**
   * 安全执行函数
   * @param {Function} fn - 要执行的函数
   * @param {Object} context - 上下文信息
   * @param {Function} onError - 错误回调
   */
  const safeExecute = async (fn, context = {}, onError = null) => {
    isLoading.value = true
    clearError()

    try {
      return await fn()
    } catch (error) {
      const errorInfo = handleError(error, context)

      if (onError) {
        onError(errorInfo)
      }

      // 如果错误可重试且启用了自动重试
      if (errorInfo.retryable && config.autoRetry) {
        return await retry(() => fn())
      }

      throw error
    } finally {
      isLoading.value = false
    }
  }

  /**
   * 显示最大重试次数对话框
   */
  const showMaxRetriesDialog = async (errorInfo) => {
    try {
      await ElMessageBox.confirm(
        `操作已重试 ${config.maxRetries} 次仍然失败。\n错误信息：${errorInfo.message}`,
        '重试次数已达上限',
        {
          confirmButtonText: '联系技术支持',
          cancelButtonText: '稍后再试',
          type: 'warning',
          dangerouslyUseHTMLString: true
        }
      )

      // 用户选择联系技术支持
      window.open('mailto:support@example.com?subject=应用错误报告&body=' +
                  encodeURIComponent(JSON.stringify(errorInfo, null, 2)))
    } catch {
      // 用户选择稍后再试，不做任何操作
    }
  }

  /**
   * 获取错误建议
   * @param {Object} errorInfo - 错误信息
   */
  const getErrorSuggestions = (errorInfo) => {
    const suggestions = []

    switch (errorInfo.type) {
      case ErrorTypes.NETWORK:
        suggestions.push(
          '检查网络连接是否正常',
          '尝试切换网络环境',
          '稍后再试'
        )
        break
      case ErrorTypes.AUTH:
        suggestions.push(
          '重新登录应用',
          '检查API Token配置',
          '联系管理员获取权限'
        )
        break
      case ErrorTypes.SYNC:
        suggestions.push(
          '检查网络连接',
          '查看同步历史记录',
          '尝试手动同步'
        )
        break
      case ErrorTypes.DATABASE:
        suggestions.push(
          '重启应用程序',
          '检查数据库文件权限',
          '联系技术支持'
        )
        break
      case ErrorTypes.VALIDATION:
        suggestions.push(
          '检查输入参数格式',
          '参考文档要求',
          '使用默认值重试'
        )
        break
      default:
        suggestions.push(
          '重启应用程序',
          '查看错误详情',
          '联系技术支持'
        )
    }

    return suggestions
  }

  /**
   * 格式化错误详情
   * @param {Object} errorInfo - 错误信息
   */
  const formatErrorDetails = (errorInfo) => {
    if (!config.showDetailedErrors) {
      return null
    }

    return {
      id: errorInfo.id,
      timestamp: errorInfo.timestamp,
      type: errorInfo.type,
      level: errorInfo.level,
      code: errorInfo.code,
      context: errorInfo.context,
      suggestions: getErrorSuggestions(errorInfo)
    }
  }

  /**
   * 导出错误日志
   */
  const exportErrorLog = () => {
    const logData = {
      exportTime: new Date().toISOString(),
      errorStats: errorStats.value,
      currentError: currentError.value,
      errorHistory: errorHistory.value.slice(0, 10), // 只导出最近10个错误
      appInfo: {
        userAgent: navigator.userAgent,
        url: window.location.href,
        timestamp: new Date().toISOString()
      }
    }

    const blob = new Blob([JSON.stringify(logData, null, 2)], {
      type: 'application/json'
    })

    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `error-log-${new Date().toISOString().split('T')[0]}.json`
    link.click()
    URL.revokeObjectURL(url)
  }

  /**
   * 重置错误状态
   */
  const reset = () => {
    clearError()
    errorHistory.value = []
    retryCount.value = 0
  }

  return {
    // 状态
    currentError,
    errorHistory,
    isLoading,
    retryCount,
    hasError,
    errorCount,
    canRetry,
    errorStats,

    // 方法
    handleError,
    clearError,
    retry,
    safeExecute,
    getErrorSuggestions,
    formatErrorDetails,
    exportErrorLog,
    reset
  }
}

/**
 * API 调用错误处理 Composable
 * 专门用于处理API调用中的错误
 */
export function useApiErrorHandler() {
  const { handleError, clearError, safeExecute } = useErrorHandler()

  /**
   * 安全的API调用
   * @param {Function} apiCall - API调用函数
   * @param {Object} context - API调用上下文
   */
  const safeApiCall = async (apiCall, context = {}) => {
    return safeExecute(apiCall, {
      operation: 'api_call',
      ...context
    })
  }

  /**
   * 处理API响应错误
   * @param {Object} response - API响应
   * @param {Object} context - 响应上下文
   */
  const handleApiResponse = (response, context = {}) => {
    if (!response.success) {
      throw new Error(response.message || 'API调用失败')
    }
    return response.data
  }

  /**
   * 完整的API调用处理
   * @param {Function} apiCall - API调用函数
   * @param {Object} context - 调用上下文
   */
  const executeApiCall = async (apiCall, context = {}) => {
    return safeApiCall(async () => {
      const response = await apiCall()
      return handleApiResponse(response, context)
    }, context)
  }

  return {
    handleError,
    clearError,
    safeApiCall,
    handleApiResponse,
    executeApiCall
  }
}

export default useErrorHandler