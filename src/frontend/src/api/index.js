// Wails API 集成 - 替换 HTTP 调用为直接的 Wails 方法调用
import {
  StartSync,
  GetSyncStatusAsync,
  SyncBills,
  GetBills,
  GetBillsByDateRange,
  GetBillByID,
  DeleteBill,
  GetStats,
  GetSyncStatus,
  GetSyncHistory,
  SaveToken,
  GetToken,
  DeleteToken,
  ValidateToken,
  ValidateSavedToken,
  GetAllTokens,
  GetConfig,
  SetConfig,
  GetAllConfigs,
  GetHourlyUsage,
  GetModelDistribution,
  GetRecentUsage,
  GetUsageTrend,
  CheckAPIConnectivity,
  GetDatabaseInfo,
  SyncRecentMonths,
  ForceResetSyncStatus,
  GetApiUsageProgress,
  GetTokenUsageProgress,
  GetTotalCostProgress,
  GetAutoSyncConfig,
  SaveAutoSyncConfig,
  TriggerAutoSync,
  StopAutoSync,
  GetAutoSyncStatus,
  GetProductNames,
  GetBillsCount,
  GetCurrentMembershipTier
} from '../../wailsjs/go/main/App'

/**
 * Wails API 包装器 - 将原来的 HTTP API 调用转换为 Wails 方法调用
 * 保持与原有 API 接口的兼容性，最小化对前端组件的影响
 */

// 错误处理函数 - 增强版，使用新的错误处理器
import { handleError, ErrorTypes, withRetry } from '../utils/errorHandler'

const handleWailsError = (error, context = {}) => {
  // 使用新的错误处理器处理错误
  const errorInfo = handleError(error, {
    operation: 'wails_api_call',
    ...context
  })

  return {
    success: false,
    message: errorInfo.message,
    data: null,
    errorInfo: errorInfo // 保留完整的错误信息供调试使用
  }
}

// 成功响应包装器 - 将 Go 返回值转换为前端期望的格式
const handleWailsSuccess = (data, message = '操作成功') => {
  return {
    success: true,
    message,
    data
  }
}

export default {
  // 新的异步账单同步方法
  async startSync(billingMonth) {
    return withRetry(async () => {
      try {
        // 直接传递 billingMonth 字符串给后端
        const result = await StartSync(billingMonth)
        if (result.success) {
          return handleWailsSuccess(result, result.message || '同步任务已启动')
        } else {
          return handleWailsError(new Error(result.message || '同步启动失败'), {
            operation: 'startSync',
            billingMonth,
            apiResponse: result
          })
        }
      } catch (error) {
        return handleWailsError(error, {
          operation: 'startSync',
          billingMonth,
          errorPhase: 'wails_api_call'
        })
      }
    }, {
      maxRetries: 2,
      retryDelay: 1000
    })
  },

  // 获取异步同步状态
  async getSyncStatusAsync() {
    try {
      const result = await GetSyncStatusAsync()
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 账单同步（修复参数类型 - FRONTEND_02）
  async syncBills(billingMonth, syncType = 'full') {
    return withRetry(async () => {
      try {
        // 直接传递billingMonth字符串和syncType给后端
        // 后端期望 (billingMonth: string, syncType: string) - 参数类型已正确
        const result = await SyncBills(billingMonth, syncType)
        return handleWailsSuccess(result, '同步已启动')
      } catch (error) {
        return handleWailsError(error, {
          operation: 'syncBills',
          billingMonth,
          syncType
        })
      }
    }, {
      maxRetries: 2,
      retryDelay: 1000
    })
  },

  // 获取账单列表
  async getBills(params = {}) {
    return withRetry(async () => {
      try {
        // 将查询参数转换为 Wails API 期望的格式
        const { page = 1, pageSize = 20, startDate, endDate, model } = params

        if (startDate && endDate) {
          // 如果有日期范围，使用 GetBillsByDateRange
          // 将日期字符串转换为 Date 对象
          const startDateTime = new Date(startDate)
          const endDateTime = new Date(endDate)
          const result = await GetBillsByDateRange(startDateTime, endDateTime, page, pageSize)
          return handleWailsSuccess(result)
        } else {
          // 否则使用 GetBills - 修复参数结构
          const billParams = {
            page_num: page,
            page_size: pageSize,
            model_name: model ? model : undefined
          }
          const result = await GetBills(billParams)
          return handleWailsSuccess(result)
        }
      } catch (error) {
        return handleWailsError(error)
      }
    }, {
      maxRetries: 2,
      retryDelay: 1000
    })
  },

  // 根据ID获取账单详情
  async getBillByID(id) {
    try {
      const result = await GetBillByID(id)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 删除账单
  async deleteBill(id) {
    try {
      await DeleteBill(id)
      return handleWailsSuccess(null, '账单删除成功')
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 获取统计数据（添加period参数支持 - FRONTEND_02）
  async getStats(period = '5h') {
    try {
      // 将 period 参数转换为 startDate 和 endDate
      const now = new Date()
      let startDate

      switch (period) {
        case '1h':
          startDate = new Date(now.getTime() - 60 * 60 * 1000)
          break
        case '5h':
          startDate = new Date(now.getTime() - 5 * 60 * 60 * 1000)
          break
        case '24h':
          startDate = new Date(now.getTime() - 24 * 60 * 60 * 1000)
          break
        case '7d':
          startDate = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
          break
        default:
          startDate = new Date(now.getTime() - 5 * 60 * 60 * 1000)
      }

      // GetStats 现在期望 (startDate, endDate, period)
      const result = await GetStats(startDate, now, period)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error, {
        operation: 'getStats',
        period
      })
    }
  },

  // 获取每小时使用量
  async getHourlyUsage(hours = 5) {
    try {
      const result = await GetHourlyUsage(hours)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 获取模型使用分布
  async getProductDistribution(hours = 5) {
    try {
      const endDate = new Date()
      const startDate = new Date(Date.now() - hours * 60 * 60 * 1000)

      const result = await GetModelDistribution(startDate, endDate)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 获取同步状态（新的异步方法）
  async getSyncStatus() {
    try {
      const result = await GetSyncStatusAsync()
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 获取同步历史（修复参数顺序）
  async getSyncHistory(syncType, page = 1, limit = 10) {
    try {
      // 后端期望 (syncType, pageNum, pageSize)
      const result = await GetSyncHistory(syncType, page, limit)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error, {
        operation: 'getSyncHistory',
        syncType,
        page,
        limit
      })
    }
  },

  // Token 管理（修复参数传递 - FRONTEND_02）
  async saveToken(token, tokenName = 'default') {
    return withRetry(async () => {
      try {
        // 后端期望 (tokenName, tokenValue) - 修复参数顺序
        await SaveToken(tokenName, token)
        return handleWailsSuccess(null, 'Token 保存成功')
      } catch (error) {
        return handleWailsError(error, {
          operation: 'saveToken',
          tokenName
        })
      }
    }, {
      maxRetries: 2,
      retryDelay: 1000
    })
  },

  async getToken() {
    return withRetry(async () => {
      try {
        const result = await GetToken()
        // 修复返回值处理 - 统一返回格式
        if (result && result.token_value) {
          return handleWailsSuccess({ token: result.token_value }, 'Token获取成功')
        } else {
          return handleWailsSuccess({ token: null }, '未找到Token')
        }
      } catch (error) {
        return handleWailsError(error)
      }
    }, {
      maxRetries: 2,
      retryDelay: 1000
    })
  },

  async getAllTokens() {
    try {
      const result = await GetAllTokens()
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async deleteToken(id) {
    try {
      await DeleteToken(id)
      return handleWailsSuccess(null, 'Token 删除成功')
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async verifyToken(token) {
    try {
      const result = await ValidateToken(token)
      return handleWailsSuccess(result, 'Token 验证完成')
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async validateSavedToken() {
    try {
      const result = await ValidateSavedToken()
      return handleWailsSuccess(result, '保存的Token验证完成')
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 配置管理
  async getAutoSyncConfig() {
    try {
      const result = await GetConfig('auto_sync')
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async saveAutoSyncConfig(config) {
    try {
      await SetConfig('auto_sync', config, '自动同步配置')
      return handleWailsSuccess(null, '配置保存成功')
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 统计数据 API - 映射到 Wails 方法
  async getDayApiUsage() {
    try {
      const result = await GetRecentUsage(1)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getDayTokenUsage() {
    try {
      const result = await GetRecentUsage(1)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getWeekApiUsage() {
    try {
      const result = await GetRecentUsage(7)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getWeekTokenUsage() {
    try {
      const result = await GetRecentUsage(7)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getMonthApiUsage() {
    try {
      const result = await GetRecentUsage(30)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getMonthTokenUsage() {
    try {
      const result = await GetRecentUsage(30)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getDailyUsage(days = 7) {
    try {
      const result = await GetRecentUsage(days)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getMonthlyUsage() {
    try {
      const result = await GetRecentUsage(30)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 使用趋势
  async getUsageTrend(days = 7) {
    try {
      const result = await GetUsageTrend(days)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 数据库信息
  async getDatabaseInfo() {
    try {
      const result = await GetDatabaseInfo()
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // API 连接检查
  async checkAPIConnectivity() {
    try {
      const result = await CheckAPIConnectivity()
      return handleWailsSuccess(result, '连接检查完成')
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 同步最近几个月
  async syncRecentMonths(months = 3) {
    try {
      const result = await SyncRecentMonths(months)
      return handleWailsSuccess(result, `${months}个月数据同步已启动`)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 获取产品名称列表（修复方案A：修改后端返回格式）
  async getProducts() {
    try {
      // 直接调用新的GetProductNames方法
      const result = await GetProductNames()
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getBillsCount() {
    try {
      const result = await GetBillsCount()
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 获取当前会员等级
  async getCurrentMembershipTier() {
    try {
      const result = await GetCurrentMembershipTier()
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getApiUsageProgress() {
    try {
      const result = await GetApiUsageProgress()
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getTokenUsageProgress() {
    try {
      const result = await GetTokenUsageProgress()
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getTotalCostProgress() {
    try {
      const result = await GetTotalCostProgress()
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getDayTotalCost() {
    try {
      const result = await GetRecentUsage(1)
      return handleWailsSuccess(result?.cost?.total || 0)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getWeekTotalCost() {
    try {
      const result = await GetRecentUsage(7)
      return handleWailsSuccess(result?.cost?.total || 0)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getMonthTotalCost() {
    try {
      const result = await GetRecentUsage(30)
      return handleWailsSuccess(result?.cost?.total || 0)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 重置同步状态（用于解决 "sync already in progress" 问题）
  async resetSyncStatus() {
    try {
      const result = await ForceResetSyncStatus()
      return handleWailsSuccess(result, result.message || '同步状态已重置')
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // ========== 自动同步相关方法 ==========

  // 触发一次自动同步
  async triggerAutoSync() {
    try {
      const result = await TriggerAutoSync()
      return handleWailsSuccess(result, result.message || '自动同步已触发')
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 停止自动同步
  async stopAutoSync() {
    try {
      const result = await StopAutoSync()
      return handleWailsSuccess(result, result.message || '自动同步已停止')
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 获取自动同步状态
  async getAutoSyncStatus() {
    try {
      const result = await GetAutoSyncStatus()
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  }
}
