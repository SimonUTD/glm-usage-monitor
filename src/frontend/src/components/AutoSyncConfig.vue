<template>
  <div class="auto-sync-config-compact">
    <div class="auto-sync-switch">
      <span class="config-label">
        <el-icon><Switch /></el-icon>
        启用自动同步
      </span>
      <el-switch
        :model-value="autoSyncConfig.enabled"
        :loading="loading"
        @change="handleToggle"
      />
    </div>

    <div v-if="autoSyncConfig.enabled" class="auto-sync-frequency">
      <span class="frequency-label">同步频率</span>
      <el-select
        :model-value="autoSyncConfig.frequency_seconds"
        @change="handleFrequencyChange"
        :loading="loading"
        class="frequency-select-compact"
        size="small"
      >
        <el-option :value="5" label="5秒" />
        <el-option :value="10" label="10秒" />
        <el-option :value="60" label="1分钟" />
        <el-option :value="300" label="5分钟" />
      </el-select>
    </div>

    <div v-if="autoSyncConfig.enabled" class="auto-sync-status">
      <el-tag
        :type="autoSyncConfig.enabled ? 'success' : 'info'"
        size="small"
        effect="plain"
      >
        <el-icon v-if="autoSyncConfig.enabled">
          <Loading class="rotating" />
        </el-icon>
        {{ autoSyncConfig.enabled ? '自动同步中' : '已停止' }}
      </el-tag>
    </div>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { Switch, Loading } from '@element-plus/icons-vue'
import api from '@/api'
import { getCurrentMonth } from '@/utils/formatters'

const props = defineProps({
  autoSyncConfig: {
    type: Object,
    required: true,
    default: () => ({
      enabled: false,
      frequency_seconds: 10,
      next_sync_time: null
    })
  }
})

const emit = defineEmits(['update:config', 'loading-change'])

import { ref } from 'vue'

const loading = ref(false)

// 处理自动同步开关
const handleToggle = async (enabled) => {
  if (!enabled) {
    // 关闭自动同步
    try {
      loading.value = true
      emit('loading-change', true)
      const result = await api.stopAutoSync()
      if (result.success) {
        emit('update:config', {
          ...props.autoSyncConfig,
          enabled: false,
          next_sync_time: null
        })
        ElMessage.success('自动同步已停止')
      }
    } catch (error) {
      ElMessage.error('停止自动同步失败：' + error.message)
    } finally {
      loading.value = false
      emit('loading-change', false)
    }
  } else {
    // 开启自动同步 - 需要校验是否有基础数据
    loading.value = true
    emit('loading-change', true)

    try {
      // 1. 检查是否有基础数据
      const countResult = await api.getBillsCount()
      if (!countResult.success) {
        throw new Error(countResult.message || '校验失败')
      }

      const { total, hasData } = countResult.data

      // 2. 如果没有数据，阻止开启并提示
      if (!hasData) {
        ElMessage.warning({
          message: '无法开启自动同步：系统中暂无账单数据\n\n请先进行一次"全量同步"或"增量同步"来获取基础数据',
          duration: 5000
        })
        return
      }

      // 3. 有数据，继续开启自动同步
      const result = await api.saveAutoSyncConfig({
        enabled: true,
        frequency_seconds: props.autoSyncConfig.frequency_seconds
      })

      if (result.success) {
        emit('update:config', {
          ...props.autoSyncConfig,
          enabled: true,
          ...result.data
        })
        ElMessage.success('自动同步已开启')
      }

    } catch (error) {
      ElMessage.error('开启自动同步失败：' + error.message)
    } finally {
      loading.value = false
      emit('loading-change', false)
    }
  }
}

// 处理频率变化
const handleFrequencyChange = async (frequency) => {
  if (!props.autoSyncConfig.enabled) return

  try {
    loading.value = true
    emit('loading-change', true)
    const result = await api.saveAutoSyncConfig({
      enabled: true,
      frequency_seconds: frequency
    })
    if (result.success) {
      emit('update:config', {
        ...props.autoSyncConfig,
        frequency_seconds: frequency,
        ...result.data
      })
      ElMessage.success('频率已更新')
    }
  } catch (error) {
    ElMessage.error('更新频率失败：' + error.message)
  } finally {
    loading.value = false
    emit('loading-change', false)
  }
}
</script>

<style scoped>
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

/* 响应式设计 */
@media (max-width: 768px) {
  .auto-sync-config-compact {
    width: 100%;
  }
}
</style>