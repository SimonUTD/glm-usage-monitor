<template>
  <div class="sync-page">

    <el-card class="sync-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon><Refresh /></el-icon>
          <span>数据同步</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" class="sync-tabs" @tab-change="handleTabChange">
        <!-- 增量同步 -->
        <el-tab-pane label="增量同步" name="incremental">
          <div>
            <div class="description">
              <el-alert
                type="info"
                :closable="false"
                show-icon
                class="description-alert"
              >
                <template #title>
                  <div class="alert-content">
                    <span>增量同步仅获取比本地数据库更新的数据，适用于日常数据更新</span>
                  </div>
                </template>
              </el-alert>
            </div>

            <div class="sync-form">
              <el-form label-position="top">
                <el-form-item>
                  <div class="sync-config-row">
                    <!-- 当前月份显示 -->
                    <div class="current-month-display">
                      <el-icon class="calendar-icon"><Calendar /></el-icon>
                      <span class="month-text">{{ getCurrentMonth() }}</span>
                      <el-tag type="info" size="small" effect="plain">同步月份</el-tag>
                    </div>

                    <!-- 自动同步配置 -->
                    <AutoSyncConfig
                      :auto-sync-config="autoSyncConfig"
                      @update:config="handleAutoSyncConfigUpdate"
                      @loading-change="handleAutoSyncLoadingChange"
                    />
                  </div>
                </el-form-item>

                <el-form-item>
                  <el-button
                    type="primary"
                    @click="handleIncrementalSync"
                    :loading="incrementalSyncing"
                    :disabled="incrementalSyncing"
                    class="sync-button"
                  >
                    <el-icon v-if="!incrementalSyncing"><Upload /></el-icon>
                    {{ incrementalSyncing ? '同步中...' : '开始增量同步' }}
                  </el-button>
                </el-form-item>
              </el-form>
            </div>

            <!-- 增量同步进度条 -->
            <SyncProgress
              v-if="incrementalSyncing"
              :progress="incrementalProgress"
              progress-type="incremental"
            />

            <!-- 增量同步历史记录 -->
            <SyncHistory
              ref="incrementalHistoryRef"
              sync-type="incremental"
            />
          </div>
        </el-tab-pane>

        <!-- 全量同步 -->
        <el-tab-pane label="全量同步" name="full">
          <div>
            <div class="description">
              <el-alert
                type="warning"
                :closable="false"
                show-icon
                class="description-alert"
              >
                <template #title>
                  <div class="alert-content">
                    <span>全量同步将清空现有数据并重新同步，风险较高！请确认您已备份重要数据。</span>
                  </div>
                </template>
              </el-alert>
            </div>

            <div class="sync-form">
              <el-form :model="fullForm" label-position="top">
                <el-form-item label="账单月份">
                  <el-date-picker
                    v-model="fullForm.billingMonth"
                    type="month"
                    placeholder="选择月份"
                    format="YYYY-MM"
                    value-format="YYYY-MM"
                    class="month-picker"
                  />
                </el-form-item>

                <el-form-item>
                  <el-button
                    type="primary"
                    @click="handleFullSync"
                    :loading="fullSyncing"
                    :disabled="fullSyncing"
                    class="sync-button full-sync-button"
                  >
                    <el-icon v-if="!fullSyncing"><Upload /></el-icon>
                    {{ fullSyncing ? '同步中...' : '开始全量同步' }}
                  </el-button>
                </el-form-item>
              </el-form>
            </div>

            <!-- 进度条 - 仅全量同步时显示 -->
            <SyncProgress
              v-if="fullSyncing"
              :progress="progress"
              progress-type="full"
            />

            <!-- 全量同步历史记录 -->
            <SyncHistory
              ref="fullHistoryRef"
              sync-type="full"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

  </div>
</template>

<script setup>
import { ref, onMounted, onActivated, onUnmounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Upload, Loading, WarningFilled, Clock, Calendar, Switch, Timer } from '@element-plus/icons-vue'
import api from '@/api'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import timezone from 'dayjs/plugin/timezone'
import SyncProgress from '@/components/SyncProgress.vue'
import SyncHistory from '@/components/SyncHistory.vue'
import AutoSyncConfig from '@/components/AutoSyncConfig.vue'
import { getCurrentMonth } from '@/utils/formatters'

// 扩展 dayjs
dayjs.extend(utc)
dayjs.extend(timezone)

// 设置默认时区
dayjs.tz.setDefault('Asia/Shanghai')

const activeTab = ref('incremental')

// 全量同步表单
const fullForm = ref({
  billingMonth: ''
})

// 自动同步配置
const autoSyncConfig = ref({
  enabled: false,
  frequency_seconds: 10,
  next_sync_time: null
})
const autoSyncLoading = ref(false)

// 组件引用
const incrementalHistoryRef = ref(null)
const fullHistoryRef = ref(null)

// 状态
const incrementalSyncing = ref(false)
const fullSyncing = ref(false)
const progress = ref({
  percentage: 0,
  current: 0,
  total: 0,
  stage: 'idle'
})
let progressTimer = null

// 增量同步进度数据
const incrementalProgress = ref({
  percentage: 0,
  current: 0,
  total: 0,
  stage: '正在准备增量同步...',
  syncedCount: 0
})

// 同步历史记录已移至SyncHistory组件

// 增量同步
const handleIncrementalSync = async () => {
  // 检查数据库中是否有配置Token
  try {
    const tokenResult = await api.getToken()
    if (!tokenResult.success || !tokenResult.data || !tokenResult.data.token) {
      ElMessage.error('请先在设置页面配置 API Token')
      return
    }
  } catch (error) {
    ElMessage.error('检查 Token 配置失败')
    return
  }

  // 检查是否有基础数据
  try {
    const countResult = await api.getBillsCount()
    if (!countResult.success) {
      throw new Error(countResult.message || '校验失败')
    }

    const { total, hasData } = countResult.data

    if (!hasData) {
      ElMessage.warning({
        message: '无法进行增量同步：系统中暂无账单数据\n\n请先进行一次"全量同步"来获取基础数据',
        duration: 5000
      })
      return
    }
  } catch (error) {
    ElMessage.error('检查数据失败：' + error.message)
    return
  }

  incrementalSyncing.value = true

  try {
    // 使用新的异步同步方法
    const result = await api.startSync(getCurrentMonth())

    if (result.success) {
      // 重置增量同步进度状态
      incrementalProgress.value = {
        percentage: 0,
        current: 0,
        total: 0,
        stage: result.message || '增量同步任务已启动，正在准备...',
        syncedCount: 0
      }

      ElMessage.success(result.message || '增量数据同步已启动')

      // 启动异步轮询来跟踪同步状态
      startIncrementalSyncPolling()
    } else {
      ElMessage.error(result.message || '同步启动失败')
      incrementalSyncing.value = false
    }
  } catch (error) {
    incrementalSyncing.value = false
    ElMessage.error('同步失败：' + error.message)
  }
}

// 全量同步
const handleFullSync = async () => {
  if (!fullForm.value.billingMonth) {
    ElMessage.warning('请选择账单月份')
    return
  }

  // 检查数据库中是否有配置Token
  try {
    const tokenResult = await api.getToken()
    if (!tokenResult.success || !tokenResult.data || !tokenResult.data.token) {
      ElMessage.error('请先在设置页面配置 API Token')
      return
    }
  } catch (error) {
    ElMessage.error('检查 Token 配置失败')
    return
  }

  // 先检查是否已经有同步在进行
  const statusResult = await api.getSyncStatus()
  if (statusResult.success && statusResult.data.syncing) {
    // 已经有同步在进行，恢复前端状态
    ElMessage.info('检测到已有同步正在进行，已恢复同步状态')
    fullSyncing.value = true
    if (statusResult.data.progress) {
      progress.value = statusResult.data.progress
    }
    if (!progressTimer) {
      startProgressPolling()
    }
    return
  }

  try {
    await ElMessageBox.confirm(
      '全量同步将清空现有数据并重新同步，风险较高！\n请确认您已备份重要数据。',
      '全量同步确认',
      {
        confirmButtonText: '确认执行',
        cancelButtonText: '取消',
        type: 'warning',
        center: true,
        customClass: 'full-sync-confirm-dialog',
        confirmButtonClass: 'confirm-danger-btn',
        cancelButtonClass: 'cancel-btn'
      }
    )
  } catch {
    ElMessage.info('已取消同步')
    return
  }

  fullSyncing.value = true
  progress.value = {
    percentage: 0,
    current: 0,
    total: 0,
    stage: 'idle'
  }

  startProgressPolling()

  try {
    // 使用新的异步同步方法
    const result = await api.startSync(fullForm.value.billingMonth)

    if (result.success) {
      ElMessage.success(result.message || '全量数据同步已启动')
      // 进度轮询已经在之前启动
    } else {
      fullSyncing.value = false
      stopProgressPolling()
      ElMessage.error(result.message || '同步启动失败')
    }
  } catch (error) {
    fullSyncing.value = false
    stopProgressPolling()
    ElMessage.error('同步失败：' + error.message)
  }
}

// 增量同步状态轮询
let incrementalProgressTimer = null
let incrementalRetryCount = 0
const INCREMENTAL_MAX_RETRIES = 5
const INCREMENTAL_TIMEOUT = 600000 // 10分钟超时

const startIncrementalSyncPolling = () => {
  incrementalRetryCount = 0
  const startTime = Date.now()

  incrementalProgressTimer = setInterval(async () => {
    try {
      // 检查超时
      if (Date.now() - startTime > INCREMENTAL_TIMEOUT) {
        incrementalSyncing.value = false
        stopIncrementalSyncPolling()
        ElMessage.error('增量同步超时，请检查网络连接或重新尝试')
        return
      }

      const result = await api.getSyncStatus()
      if (result.success) {
        const syncData = result.data

        // 更新增量同步进度信息
        if (syncData.current_page !== undefined && syncData.total_pages !== undefined) {
          incrementalProgress.value = {
            percentage: syncData.progress || 0,
            current: syncData.current_page,
            total: syncData.total_pages,
            stage: syncData.message || '正在同步增量数据...',
            syncedCount: syncData.synced_count || 0
          }
        }

        // 重置重试计数
        incrementalRetryCount = 0

        // 检查同步是否完成
        if (!syncData.syncing) {
          // 同步完成
          incrementalSyncing.value = false
          stopIncrementalSyncPolling()

          // 显示同步结果
          if (syncData.status === 'completed' || syncData.status === 'success') {
            ElMessage.success('增量数据同步完成')
          } else if (syncData.status === 'failed') {
            ElMessage.error('同步失败：' + (syncData.message || '未知错误'))
          } else {
            ElMessage.success('增量数据同步完成')
          }

          // 重新加载历史记录
          if (incrementalHistoryRef.value) {
            incrementalHistoryRef.value.loadHistory()
          }
        }
      }
    } catch (error) {
      console.error('获取增量同步进度失败：', error)
      incrementalRetryCount++
      
      // 增加错误重试机制
      if (incrementalRetryCount >= INCREMENTAL_MAX_RETRIES) {
        incrementalSyncing.value = false
        stopIncrementalSyncPolling()
        ElMessage.error('连续获取同步状态失败，请检查网络连接')
      } else {
        incrementalProgress.value.stage = `获取进度信息失败，正在重试... (${incrementalRetryCount}/${INCREMENTAL_MAX_RETRIES})`
      }
    }
  }, 1000)
}

const stopIncrementalSyncPolling = () => {
  if (incrementalProgressTimer) {
    clearInterval(incrementalProgressTimer)
    incrementalProgressTimer = null
    incrementalRetryCount = 0
  }
}

// 全量同步状态轮询变量
let fullSyncRetryCount = 0
const FULL_SYNC_MAX_RETRIES = 5
const FULL_SYNC_TIMEOUT = 600000 // 10分钟超时

// 开始轮询同步进度
const startProgressPolling = () => {
  fullSyncRetryCount = 0
  const startTime = Date.now()

  progressTimer = setInterval(async () => {
    try {
      // 检查超时
      if (Date.now() - startTime > FULL_SYNC_TIMEOUT) {
        fullSyncing.value = false
        stopProgressPolling()
        ElMessage.error('同步超时，请检查网络连接或重新尝试')
        return
      }

      const result = await api.getSyncStatus()
      if (result.success) {
        const syncData = result.data

        // 更新进度信息
        if (syncData.current_page !== undefined && syncData.total_pages !== undefined) {
          progress.value = {
            percentage: syncData.progress || 0,
            current: syncData.current_page,
            total: syncData.total_pages,
            stage: syncData.message || '正在同步数据...',
            syncedCount: syncData.synced_count || 0
          }
        }

        // 重置重试计数
        fullSyncRetryCount = 0

        // 检查同步是否完成
        if (!syncData.syncing) {
          fullSyncing.value = false
          stopProgressPolling()

          // 显示同步结果
          if (syncData.status === 'completed' || syncData.status === 'success') {
            ElMessage.success('全量数据同步完成')
          } else if (syncData.status === 'failed') {
            ElMessage.error('同步失败：' + (syncData.message || '未知错误'))
          } else {
            ElMessage.success('数据同步完成')
          }

          // 重新加载历史记录
          if (fullHistoryRef.value) {
            fullHistoryRef.value.loadHistory()
          }
        }
      }
    } catch (error) {
      console.error('获取同步进度失败：', error)
      fullSyncRetryCount++
      
      // 增加错误重试机制
      if (fullSyncRetryCount >= FULL_SYNC_MAX_RETRIES) {
        fullSyncing.value = false
        stopProgressPolling()
        ElMessage.error('连续获取同步状态失败，请检查网络连接')
      } else {
        progress.value.stage = `获取进度信息失败，正在重试... (${fullSyncRetryCount}/${FULL_SYNC_MAX_RETRIES})`
      }
    }
  }, 1000)
}

// 停止轮询同步进度
const stopProgressPolling = () => {
  if (progressTimer) {
    clearInterval(progressTimer)
    progressTimer = null
    fullSyncRetryCount = 0
  }
}

// 获取阶段文本
const getStageText = (stage) => {
  const stageMap = {
    'idle': '等待中',
    'clearing': '正在清空数据',
    'fetching': '正在获取数据',
    'saving': '正在保存数据',
    'completed': '同步完成'
  }
  return stageMap[stage] || '处理中'
}

// 处理自动同步配置更新
const handleAutoSyncConfigUpdate = (newConfig) => {
  autoSyncConfig.value = newConfig
}

// 处理自动同步加载状态变化
const handleAutoSyncLoadingChange = (loading) => {
  autoSyncLoading.value = loading
}

// 获取当前月份已从formatters导入

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return '--'
  // 后端已返回本地时间格式，直接使用
  return timeStr
}

// 加载自动同步配置
const loadAutoSyncConfig = async () => {
  try {
    const result = await api.getAutoSyncConfig()
    if (result.success) {
      autoSyncConfig.value = {
        ...result.data
      }
    }
  } catch (error) {
    console.error('加载自动同步配置失败:', error)
  }
}

// 恢复全量同步状态
const restoreFullSyncStatus = async () => {
  try {
    const result = await api.getSyncStatus()
    if (result.success && result.data.syncing) {
      // 后台正在同步，恢复前端状态
      fullSyncing.value = true

      const syncData = result.data
      if (syncData.progress) {
        progress.value = {
          percentage: syncData.progress || 0,
          current: syncData.syncedItems || 0,
          total: syncData.totalItems || 0,
          stage: syncData.message || 'processing'
        }
      }

      // 启动轮询
      if (!progressTimer) {
        startProgressPolling()
      }
    }
  } catch (error) {
    console.error('恢复同步状态失败：', error)
  }
}

// 监听标签页切换
const handleTabChange = (tabName) => {
  if (tabName === 'full') {
    // 切换到全量同步tab时，检查是否需要恢复状态
    restoreFullSyncStatus()
  }
}

onMounted(() => {
  const now = new Date()
  const currentMonth = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  fullForm.value.billingMonth = currentMonth

  // 加载自动同步配置
  loadAutoSyncConfig()


  // 加载同步历史记录
  loadSyncHistory()

  // 页面加载时也检查一次
  restoreFullSyncStatus()
})

onActivated(() => {
  // 页面重新激活时也检查状态
  restoreFullSyncStatus()
})

onUnmounted(() => {
  // 组件卸载时清理定时器
  stopProgressPolling()
  stopIncrementalSyncPolling()
})
</script>

<style scoped>
.sync-page {
  animation: fadeIn 0.3s ease;
}

.sync-card {
  background: #FFFFFF;
  border: 1px solid rgba(77, 103, 130, 0.12);
  border-radius: 8px;
  margin-bottom: 24px;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.sync-card:hover {
  border-color: rgba(77, 103, 130, 0.2);
  box-shadow: 0 2px 8px rgba(77, 103, 130, 0.1);
}

.card-header {
  display: flex;
  align-items: center;
  font-weight: 600;
  font-size: 18px;
  color: #4D6782;
  background: rgba(77, 103, 130, 0.05);
  padding: 20px;
  border-bottom: 1px solid rgba(77, 103, 130, 0.12);
}

.card-header .el-icon {
  margin-right: 10px;
  font-size: 20px;
  color: #4D6782;
}

.sync-tabs {
  padding: 0 20px;
}

.description {
  margin-bottom: 20px;
}

.description-alert {
  border-radius: 8px;
}

.alert-content {
  display: flex;
  align-items: center;
  color: #5A5A5A;
}

.alert-content .el-icon {
  margin-right: 10px;
  font-size: 18px;
}

.sync-form {
  margin-bottom: 20px;
}

.month-picker {
  width: 100%;
  max-width: 250px;
}

@media (min-width: 768px) {
  .month-picker {
    width: 250px;
  }
}

.month-picker :deep(.el-input__wrapper) {
  background: #F5F5F5;
  border: 1px solid #D0D0D0;
  box-shadow: none;
  width: 100%;
  box-sizing: border-box;
}

.month-picker :deep(.el-input__wrapper):hover,
.month-picker :deep(.el-input__wrapper).is-focus {
  border-color: #9DB4C0;
  box-shadow: 0 0 0 2px rgba(157, 180, 192, 0.1);
}

.current-month-display {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(77, 103, 130, 0.08);
  border: 1px solid rgba(77, 103, 130, 0.2);
  border-radius: 6px;
  font-size: 13px;
  color: #4D6782;
  font-weight: 500;
}

.calendar-icon {
  font-size: 16px;
  color: #4D6782;
}

.month-text {
  font-weight: 600;
  color: #4D6782;
  letter-spacing: 0.3px;
}

/* 同步配置行 - 水平布局 */
.sync-config-row {
  display: flex;
  gap: 16px;
  align-items: stretch;
}

/* 自动同步配置 - 紧凑版（与月份显示保持一致的颜色） */
.auto-sync-config-compact {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 8px 12px;
  background: rgba(77, 103, 130, 0.08);  /* 与 current-month-display 保持一致 */
  border: 1px solid rgba(77, 103, 130, 0.2);  /* 与 current-month-display 保持一致 */
  border-radius: 6px;
  min-height: 32px;
  flex: 1;
}

.auto-sync-switch {
  display: flex;
  align-items: center;
  gap: 8px;
}

.auto-sync-frequency {
  display: flex;
  align-items: center;
  gap: 8px;
}

.auto-sync-status {
  display: flex;
  align-items: center;
  margin-left: 8px;
}

/* 配置标签 - 紧凑版 */
.config-label {
  display: flex;
  align-items: center;
  font-size: 12px;
  color: #4D6782;
  font-weight: 500;
  padding: 0;
}

.config-label .el-icon {
  margin-right: 6px;
  font-size: 14px;
  color: #4D6782;
}

/* 频率标签 */
.frequency-label {
  font-size: 12px;
  color: #4D6782;
  font-weight: 500;
  white-space: nowrap;
}

.frequency-select-compact {
  width: 120px;
}

.frequency-select-compact :deep(.el-input__wrapper) {
  background: #F5F5F5;
  border: 1px solid rgba(77, 103, 130, 0.2);
  border-radius: 6px;
  height: 28px;
  line-height: 28px;
}

.frequency-select-compact :deep(.el-input__inner) {
  height: 26px;
  line-height: 26px;
  font-size: 12px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .sync-config-row {
    flex-direction: column;
    gap: 12px;
  }

  .auto-sync-config-compact {
    width: 100%;
  }
}

.sync-button {
  height: 36px;
  padding: 0 20px;
  font-size: 14px;
  border-radius: 6px;
  transition: all 0.2s ease;
  background: #4D6782;
  border: none;
  color: #FFFFFF;
  font-weight: 500;
}

.sync-button:hover {
  background: #3d5568;
  box-shadow: 0 1px 4px rgba(77, 103, 130, 0.15);
}

.full-sync-button {
  background: #C57272;
  border: none;
}

.full-sync-button:hover {
  background: #B55A5A;
  box-shadow: 0 1px 4px rgba(197, 114, 114, 0.25);
}

.sync-alert {
  background: rgba(77, 103, 130, 0.08);
  border-color: rgba(77, 103, 130, 0.2);
  border-radius: 12px;
  margin-bottom: 20px;
}

.progress-container {
  margin: 20px 0;
  padding: 20px;
  background: #F5F5F5;
  border-radius: 12px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.progress-stage {
  font-size: 15px;
  color: #5A5A5A;
  font-weight: 500;
}

.progress-percentage {
  font-size: 15px;
  color: #4D6782;
  font-weight: 600;
}

.progress-details {
  margin-top: 8px;
  font-size: 13px;
  color: #8E8E8E;
  text-align: center;
}

.sync-alert-content {
  display: flex;
  align-items: center;
  color: #5A5A5A;
}

.rotating {
  animation: rotate 1s linear infinite;
  margin-right: 10px;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.help-card {
  background: rgba(168, 198, 134, 0.08);
  border: 1px solid rgba(168, 198, 134, 0.2);
  border-radius: 12px;
}

.help-card .card-header {
  background: rgba(168, 198, 134, 0.05);
  border-bottom: 1px solid rgba(168, 198, 134, 0.2);
}

.help-content {
  padding: 20px 0;
}

.help-item {
  display: flex;
  align-items: center;
  padding: 14px 0;
  color: #5A5A5A;
  font-size: 15px;
  transition: all 0.2s ease;
}

.help-item:hover {
  color: #A8C686;
  transform: translateX(5px);
}

.help-icon {
  margin-right: 12px;
  color: #A8C686;
  font-size: 18px;
  flex-shrink: 0;
}

/* Tabs样式 */
.sync-tabs :deep(.el-tabs__header) {
  margin: 0 0 20px 0;
}

.sync-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: rgba(77, 103, 130, 0.12);
}

.sync-tabs :deep(.el-tabs__item) {
  font-size: 16px;
  font-weight: 500;
  color: #8E8E8E;
  padding: 0 30px;
  height: 50px;
  line-height: 50px;
}

.sync-tabs :deep(.el-tabs__item.is-active) {
  color: #4D6782;
}

.sync-tabs :deep(.el-tabs__active-bar) {
  background-color: #4D6782;
}

/* 全量同步确认对话框样式 - 使用:global确保样式在全局作用域生效 */
:global(.el-overlay-dialog.full-sync-confirm-dialog) {
  border-radius: 12px !important;
  padding: 24px 0 !important;
  box-shadow: 0 8px 32px rgba(77, 103, 130, 0.15) !important;
  border: none !important;
}

:global(.full-sync-confirm-dialog .el-message-box__header) {
  background: rgba(184, 169, 154, 0.08) !important;
  border-bottom: 1px solid rgba(184, 169, 154, 0.2) !important;
  padding: 20px 24px 16px 24px !important;
}

:global(.full-sync-confirm-dialog .el-message-box__title) {
  font-size: 18px !important;
  font-weight: 600 !important;
  color: #5A5A5A !important;
  display: flex !important;
  align-items: center !important;
  gap: 10px !important;
}

:global(.full-sync-confirm-dialog .el-message-box__content) {
  padding: 28px 24px 24px 24px !important;
}

:global(.full-sync-confirm-dialog .el-message-box__message) {
  color: #5A5A5A !important;
  font-size: 15px !important;
  line-height: 1.7 !important;
  text-align: center !important;
  white-space: pre-line !important;
}

:global(.full-sync-confirm-dialog .el-message-box__btns) {
  padding: 0 24px 24px 24px !important;
  display: flex !important;
  gap: 12px !important;
  justify-content: center !important;
}

:global(.full-sync-confirm-dialog .confirm-danger-btn) {
  background: #D4A5A5 !important;
  border: 1px solid #D4A5A5 !important;
  color: #FFFFFF !important;
  padding: 10px 28px !important;
  font-size: 15px !important;
  font-weight: 500 !important;
  border-radius: 8px !important;
  transition: all 0.2s ease !important;
}

:global(.full-sync-confirm-dialog .confirm-danger-btn:hover) {
  background: #c49494 !important;
  border-color: #c49494 !important;
  box-shadow: 0 2px 8px rgba(212, 165, 165, 0.3) !important;
  transform: translateY(-1px) !important;
}

:global(.full-sync-confirm-dialog .confirm-danger-btn:active) {
  transform: translateY(0) !important;
}

:global(.full-sync-confirm-dialog .cancel-btn) {
  background: #FFFFFF !important;
  border: 1px solid #b8a99a !important;
  color: #8b7b6f !important;
  padding: 10px 28px !important;
  font-size: 15px !important;
  font-weight: 500 !important;
  border-radius: 8px !important;
  transition: all 0.2s ease !important;
}

:global(.full-sync-confirm-dialog .cancel-btn:hover) {
  background: rgba(184, 169, 154, 0.1) !important;
  border-color: #a89988 !important;
  color: #7a6b5f !important;
  transform: translateY(-1px) !important;
}

:global(.full-sync-confirm-dialog .cancel-btn:active) {
  transform: translateY(0) !important;
}

:global(.full-sync-confirm-dialog .el-message-box__status.el-icon-warning) {
  color: #b8a99a !important;
  font-size: 24px !important;
  margin-right: 8px !important;
}

:global(.full-sync-confirm-dialog .el-message-box__headerbtn) {
  top: 18px !important;
  right: 20px !important;
  width: 28px !important;
  height: 28px !important;
  border-radius: 6px !important;
  transition: all 0.2s ease !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
}

:global(.full-sync-confirm-dialog .el-message-box__headerbtn:hover) {
  background: rgba(184, 169, 154, 0.15) !important;
  transform: rotate(90deg) !important;
}

:global(.full-sync-confirm-dialog .el-message-box__close) {
  color: #b8a99a !important;
  font-size: 16px !important;
  font-weight: 600 !important;
  transition: all 0.2s ease !important;
}

:global(.full-sync-confirm-dialog .el-message-box__headerbtn:hover .el-message-box__close) {
  color: #8b7b6f !important;
  transform: scale(1.1) !important;
}

/* 历史记录区域样式 */
.history-section {
  margin-top: 30px;
  padding: 20px;
  background: rgba(77, 103, 130, 0.05);
  border: 1px solid rgba(77, 103, 130, 0.12);
  border-radius: 8px;
}

.history-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  font-size: 16px;
  color: #4D6782;
  margin-bottom: 16px;
}

.history-title {
  display: flex;
  align-items: center;
}

.history-title .el-icon {
  margin-right: 8px;
  font-size: 18px;
  color: #4D6782;
}

.refresh-button {
  padding: 8px;
  border-radius: 8px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  background: transparent;
}

.refresh-button::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(77, 103, 130, 0.08);
  opacity: 0;
  transition: opacity 0.3s ease;
  border-radius: 8px;
}

.refresh-button::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 0;
  height: 0;
  background: rgba(77, 103, 130, 0.15);
  border-radius: 50%;
  transform: translate(-50%, -50%);
  transition: width 0.4s ease, height 0.4s ease, opacity 0.3s ease;
  opacity: 0;
}

.refresh-button:hover::before {
  opacity: 1;
}

.refresh-button:hover::after {
  width: 100%;
  height: 100%;
  opacity: 1;
}

.refresh-button:hover {
  box-shadow: 0 2px 8px rgba(77, 103, 130, 0.15);
}

.refresh-button .el-icon {
  font-size: 16px;
  color: #4D6782;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: block;
  position: relative;
  z-index: 1;
}

.refresh-button:hover .el-icon {
  color: #4D6782;
  transform: rotate(180deg) scale(1.1);
  filter: drop-shadow(0 0 2px rgba(77, 103, 130, 0.3));
}

.refresh-button:active::after {
  width: 0;
  height: 0;
  opacity: 0;
}

.refresh-button:active .el-icon {
  transform: rotate(360deg) scale(0.95);
}

.refresh-button:active {
  box-shadow: 0 1px 4px rgba(77, 103, 130, 0.1);
}

.refresh-button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.refresh-button:disabled::before,
.refresh-button:disabled::after {
  display: none;
}

.refresh-button:disabled .el-icon {
  animation: none;
  transform: none;
  filter: none;
}

.history-section :deep(.el-table) {
  background: transparent;
}

.history-section :deep(.el-table th) {
  background: rgba(77, 103, 130, 0.08);
  color: #4D6782;
  font-weight: 600;
}

.history-section :deep(.el-table td) {
  background: rgba(255, 255, 255, 0.5);
}

.history-section :deep(.el-table tr:hover > td) {
  background: rgba(77, 103, 130, 0.05);
}

/* 全量同步历史记录 - 红色系样式 */
.full-sync-history {
  background: rgba(197, 114, 114, 0.12);
  border: 1px solid rgba(197, 114, 114, 0.3);
}

.full-sync-header {
  color: #B55A5A;
}

.full-sync-header .el-icon {
  color: #B55A5A;
}

.full-sync-header .refresh-button::before {
  background: rgba(197, 114, 114, 0.08);
}

.full-sync-header .refresh-button::after {
  background: rgba(197, 114, 114, 0.15);
}

.full-sync-header .refresh-button:hover {
  box-shadow: 0 2px 8px rgba(197, 114, 114, 0.2);
}

.full-sync-header .refresh-button .el-icon {
  color: #C57272;
}

.full-sync-header .refresh-button:hover .el-icon {
  color: #C57272;
  filter: drop-shadow(0 0 2px rgba(197, 114, 114, 0.3));
}

.full-sync-header .refresh-button:active {
  box-shadow: 0 1px 4px rgba(197, 114, 114, 0.15);
}

/* 旧的自动同步配置样式已移除，改用新的紧凑版布局 */

.full-sync-table :deep(.el-table th) {
  background: rgba(197, 114, 114, 0.18);
  color: #B55A5A;
  font-weight: 600;
}

.full-sync-table :deep(.el-table td) {
  background: rgba(255, 255, 255, 0.8);
}

.full-sync-table :deep(.el-table tr:hover > td) {
  background: rgba(197, 114, 114, 0.15);
}

/* 增量同步分页样式 - 蓝色系 */
.history-section .pagination-container :deep(.el-pagination) {
  --el-pagination-bg-color: rgba(255, 255, 255, 0.8);
  --el-pagination-text-color: #4D6782;
  --el-pagination-border-radius: 8px;
}

.history-section .pagination-container :deep(.el-pagination .el-pager li) {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(77, 103, 130, 0.2);
  color: #4D6782;
  margin: 0 2px;
  border-radius: 6px;
}

.history-section .pagination-container :deep(.el-pagination .el-pager li:hover) {
  color: #4D6782;
  background: rgba(77, 103, 130, 0.1);
}

.history-section .pagination-container :deep(.el-pagination .el-pager li.is-active) {
  background: #4D6782;
  color: #FFFFFF;
  border-color: #4D6782;
}

.history-section .pagination-container :deep(.el-pagination .btn-prev),
.history-section .pagination-container :deep(.el-pagination .btn-next) {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(77, 103, 130, 0.2);
  color: #4D6782;
  border-radius: 6px;
}

.history-section .pagination-container :deep(.el-pagination .btn-prev:hover),
.history-section .pagination-container :deep(.el-pagination .btn-next:hover) {
  color: #4D6782;
  background: rgba(77, 103, 130, 0.1);
}

.history-section .pagination-container :deep(.el-pagination .el-select .el-input .el-input__inner) {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(77, 103, 130, 0.2);
  color: #4D6782;
  border-radius: 6px;
}

.history-section .pagination-container :deep(.el-pagination .el-input__inner:hover) {
  border-color: rgba(77, 103, 130, 0.4);
}

.history-section .pagination-container :deep(.el-pagination .el-input__inner:focus) {
  border-color: #4D6782;
  box-shadow: 0 0 0 2px rgba(77, 103, 130, 0.1);
}

.history-section .pagination-container :deep(.el-pagination__editor) {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(77, 103, 130, 0.2);
  color: #4D6782;
  border-radius: 6px;
}

/* 全量同步分页样式 */
.pagination-container {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.pagination-container :deep(.el-pagination) {
  --el-pagination-bg-color: rgba(255, 255, 255, 0.8);
  --el-pagination-text-color: #B55A5A;
  --el-pagination-border-radius: 8px;
}

.pagination-container :deep(.el-pagination .el-pager li) {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(197, 114, 114, 0.2);
  color: #B55A5A;
  margin: 0 2px;
  border-radius: 6px;
}

.pagination-container :deep(.el-pagination .el-pager li:hover) {
  color: #B55A5A;
  background: rgba(197, 114, 114, 0.1);
}

.pagination-container :deep(.el-pagination .el-pager li.is-active) {
  background: #C57272;
  color: #FFFFFF;
  border-color: #C57272;
}

.pagination-container :deep(.el-pagination .btn-prev),
.pagination-container :deep(.el-pagination .btn-next) {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(197, 114, 114, 0.2);
  color: #B55A5A;
  border-radius: 6px;
}

.pagination-container :deep(.el-pagination .btn-prev:hover),
.pagination-container :deep(.el-pagination .btn-next:hover) {
  color: #B55A5A;
  background: rgba(197, 114, 114, 0.1);
}

.pagination-container :deep(.el-pagination .el-select .el-input .el-input__inner) {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(197, 114, 114, 0.2);
  color: #B55A5A;
  border-radius: 6px;
}

.pagination-container :deep(.el-pagination .el-input__inner:hover) {
  border-color: rgba(197, 114, 114, 0.4);
}

.pagination-container :deep(.el-pagination .el-input__inner:focus) {
  border-color: #C57272;
  box-shadow: 0 0 0 2px rgba(197, 114, 114, 0.1);
}

.pagination-container :deep(.el-pagination__editor) {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(197, 114, 114, 0.2);
  color: #B55A5A;
  border-radius: 6px;
}
</style>
