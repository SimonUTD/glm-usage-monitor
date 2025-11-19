<template>
  <div class="history-section" :class="{ 'full-sync-history': syncType === 'full' }">
    <div class="history-header" :class="{ 'full-sync-header': syncType === 'full' }">
      <div class="history-title">
        <el-icon><Clock /></el-icon>
        <span>同步历史记录</span>
      </div>
      <el-button
        type="text"
        :loading="refreshing"
        @click="handleRefresh"
        class="refresh-button"
        :disabled="refreshing"
      >
        <el-icon><Refresh /></el-icon>
      </el-button>
    </div>
    <el-table
      :data="history"
      style="width: 100%"
      size="small"
      :empty-text="'暂无历史记录'"
      v-if="true"
      :class="{ 'full-sync-table': syncType === 'full' }"
    >
      <el-table-column prop="sync_time" label="同步时间" width="180">
        <template #default="scope">
          {{ dayjs(scope.row.sync_time).tz('Asia/Shanghai').format('YYYY-MM-DD HH:mm:ss') }}
        </template>
      </el-table-column>
      <el-table-column prop="billing_month" label="账单月份" width="120" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.status === '成功' ? 'success' : 'danger'" size="small">
            {{ scope.row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="synced_count" label="成功数量" width="100" />
      <el-table-column prop="failed_count" label="失败数量" width="100" />
      <el-table-column prop="total_count" label="总数量" width="100" />
      <el-table-column prop="message" label="消息" />
    </el-table>

    <!-- 分页 -->
    <div class="pagination-container">
      <el-pagination
        :current-page="pagination.currentPage"
        :page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Clock } from '@element-plus/icons-vue'
import api from '@/api'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import timezone from 'dayjs/plugin/timezone'

// 扩展 dayjs
dayjs.extend(utc)
dayjs.extend(timezone)

// 设置默认时区
dayjs.tz.setDefault('Asia/Shanghai')

const props = defineProps({
  syncType: {
    type: String,
    required: true,
    validator: (value) => ['incremental', 'full'].includes(value)
  }
})

const history = ref([])
const refreshing = ref(false)
const pagination = ref({
  currentPage: 1,
  pageSize: 10,
  total: 0
})

// 加载同步历史记录
const loadHistory = async () => {
  try {
    const { currentPage, pageSize } = pagination.value
    const result = await api.getSyncHistory(props.syncType, pageSize, currentPage)
    if (result.success) {
      history.value = result.data
      // 如果后端返回总数，更新分页信息
      if (result.total !== undefined) {
        pagination.value.total = result.total
      }
    }
  } catch (error) {
    console.error(`加载${props.syncType === 'full' ? '全量' : '增量'}同步历史失败:`, error)
  }
}

// 处理分页变化
const handlePageChange = (page) => {
  pagination.value.currentPage = page
  loadHistory()
}

// 处理每页显示数量变化
const handleSizeChange = (size) => {
  pagination.value.pageSize = size
  pagination.value.currentPage = 1 // 重置到第一页
  loadHistory()
}

// 刷新历史记录
const handleRefresh = async () => {
  refreshing.value = true
  try {
    await loadHistory()
    ElMessage.success('历史记录已刷新')
  } catch (error) {
    ElMessage.error('刷新失败：' + error.message)
  } finally {
    refreshing.value = false
  }
}

// 暴露方法给父组件
defineExpose({
  loadHistory
})

onMounted(() => {
  loadHistory()
})
</script>

<style scoped>
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

/* 分页样式 */
.pagination-container {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.pagination-container :deep(.el-pagination) {
  --el-pagination-bg-color: rgba(255, 255, 255, 0.8);
  --el-pagination-text-color: #4D6782;
  --el-pagination-border-radius: 8px;
}

.pagination-container :deep(.el-pagination .el-pager li) {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(77, 103, 130, 0.2);
  color: #4D6782;
  margin: 0 2px;
  border-radius: 6px;
}

.pagination-container :deep(.el-pagination .el-pager li:hover) {
  color: #4D6782;
  background: rgba(77, 103, 130, 0.1);
}

.pagination-container :deep(.el-pagination .el-pager li.is-active) {
  background: #4D6782;
  color: #FFFFFF;
  border-color: #4D6782;
}

.pagination-container :deep(.el-pagination .btn-prev),
.pagination-container :deep(.el-pagination .btn-next) {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(77, 103, 130, 0.2);
  color: #4D6782;
  border-radius: 6px;
}

.pagination-container :deep(.el-pagination .btn-prev:hover),
.pagination-container :deep(.el-pagination .btn-next:hover) {
  color: #4D6782;
  background: rgba(77, 103, 130, 0.1);
}

.pagination-container :deep(.el-pagination .el-select .el-input .el-input__inner) {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(77, 103, 130, 0.2);
  color: #4D6782;
  border-radius: 6px;
}

.pagination-container :deep(.el-pagination .el-input__inner:hover) {
  border-color: rgba(77, 103, 130, 0.4);
}

.pagination-container :deep(.el-pagination .el-input__inner:focus) {
  border-color: #4D6782;
  box-shadow: 0 0 0 2px rgba(77, 103, 130, 0.1);
}

.pagination-container :deep(.el-pagination__editor) {
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(77, 103, 130, 0.2);
  color: #4D6782;
  border-radius: 6px;
}

/* 全量同步分页样式 */
.full-sync-history .pagination-container :deep(.el-pagination) {
  --el-pagination-text-color: #B55A5A;
}

.full-sync-history .pagination-container :deep(.el-pagination .el-pager li) {
  border: 1px solid rgba(197, 114, 114, 0.2);
  color: #B55A5A;
}

.full-sync-history .pagination-container :deep(.el-pagination .el-pager li:hover) {
  color: #B55A5A;
  background: rgba(197, 114, 114, 0.1);
}

.full-sync-history .pagination-container :deep(.el-pagination .el-pager li.is-active) {
  background: #C57272;
  color: #FFFFFF;
  border-color: #C57272;
}

.full-sync-history .pagination-container :deep(.el-pagination .btn-prev),
.full-sync-history .pagination-container :deep(.el-pagination .btn-next) {
  border: 1px solid rgba(197, 114, 114, 0.2);
  color: #B55A5A;
}

.full-sync-history .pagination-container :deep(.el-pagination .btn-prev:hover),
.full-sync-history .pagination-container :deep(.el-pagination .btn-next:hover) {
  color: #B55A5A;
  background: rgba(197, 114, 114, 0.1);
}

.full-sync-history .pagination-container :deep(.el-pagination .el-select .el-input .el-input__inner) {
  border: 1px solid rgba(197, 114, 114, 0.2);
  color: #B55A5A;
}

.full-sync-history .pagination-container :deep(.el-pagination .el-input__inner:hover) {
  border-color: rgba(197, 114, 114, 0.4);
}

.full-sync-history .pagination-container :deep(.el-pagination .el-input__inner:focus) {
  border-color: #C57272;
  box-shadow: 0 0 0 2px rgba(197, 114, 114, 0.1);
}

.full-sync-history .pagination-container :deep(.el-pagination__editor) {
  border: 1px solid rgba(197, 114, 114, 0.2);
  color: #B55A5A;
}
</style>