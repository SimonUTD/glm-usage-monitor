// Wails API 集成 - 替换 HTTP 调用为直接的 Wails 方法调用
import {
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
  GetTotalCostProgress
} from '../../wailsjs/go/main/App'

/**
 * Wails API 包装器 - 将原来的 HTTP API 调用转换为 Wails 方法调用
 * 保持与原有 API 接口的兼容性，最小化对前端组件的影响
 */

// 错误处理函数 - 将 Go 错误转换为前端期望的格式
const handleWailsError = (error) => {
  if (typeof error === 'string') {
    return {
      success: false,
      message: error,
      data: null
    }
  }
  return {
    success: false,
    message: error?.message || '操作失败',
    data: null
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
  // 账单同步
  async syncBills(billingMonth, type = 'full') {
    try {
      // 从 billingMonth 解析年份和月份，格式为 "YYYY-MM"
      const [year, month] = billingMonth.split('-').map(Number)
      // Wails SyncBills 方法返回同步进度和状态
      const result = await SyncBills(year, month)
      return handleWailsSuccess(result, '同步已启动')
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 获取账单列表
  async getBills(params = {}) {
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
        // 否则使用 GetBills
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

  // 获取统计数据
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

      // GetStats 期望 time.Time 指针，传 Date 对象
      const result = await GetStats(startDate, now)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
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

  // 获取同步状态
  async getSyncStatus() {
    try {
      const result = await GetSyncStatus()
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 获取同步历史
  async getSyncHistory(syncType, limit = 10, page = 1) {
    try {
      const result = await GetSyncHistory(syncType, page, limit)
      return handleWailsSuccess(result)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // Token 管理
  async saveToken(token, description = '') {
    try {
      // 后端期望 (tokenName, tokenValue)
      await SaveToken('default', token)
      return handleWailsSuccess(null, 'Token 保存成功')
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getToken() {
    try {
      const result = await GetToken()
      // 将 APIToken 对象转换为前端期望的格式
      if (result) {
        return handleWailsSuccess({ token: result.token_value }, 'Token获取成功')
      } else {
        return handleWailsSuccess(null, '未找到Token')
      }
    } catch (error) {
      return handleWailsError(error)
    }
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

  // 以下方法是为了兼容性保留的包装器，可能需要根据实际需求调整
  async getProducts() {
    try {
      const startDate = new Date(Date.now() - 24*60*60*1000)
      const endDate = new Date()
      const result = await GetModelDistribution(startDate, endDate)
      return handleWailsSuccess(result?.models || [])
    } catch (error) {
      return handleWailsError(error)
    }
  },

  async getBillsCount() {
    try {
      const result = await GetBills({ page: 1, pageSize: 1 })
      return handleWailsSuccess(result?.total || 0)
    } catch (error) {
      return handleWailsError(error)
    }
  },

  // 占位符方法 - 需要根据实际业务逻辑实现
  async getCurrentMembershipTier() {
    try {
      const startDate = new Date(Date.now() - 24*60*60*1000)
      const endDate = new Date()
      const result = await GetStats(startDate, endDate)
      return handleWailsSuccess(result?.membership_info?.tier_name || 'free')
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
  }
}
