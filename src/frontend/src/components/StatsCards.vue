<template>
  <div class="stats-cards-container">
    <!-- 当月统计卡片 -->
    <div class="stats-section">
      <h3 class="section-title">当月统计</h3>
      <div class="stats-grid">
        <StatCardVertical
          title="API 调用次数"
          :value="formatNumber(monthlyApiUsage.value)"
          :progress="apiUsageProgress"
          :loading="monthlyApiUsageLoading"
          icon="api"
          color="#4D6782"
        />
        <StatCardVertical
          title="Token 使用量"
          :value="formatNumber(monthlyTokenUsage.value)"
          :progress="tokenUsageProgress"
          :loading="monthlyTokenUsageLoading"
          icon="token"
          color="#4D6782"
        />
        <StatCardVertical
          title="总费用"
          :value="'¥' + monthlyTotalCost.value.toFixed(2)"
          :progress="totalCostProgress"
          :loading="monthlyTotalCostLoading"
          icon="cost"
          color="#4D6782"
        />
      </div>
    </div>

    <!-- 近1天统计卡片 -->
    <div class="stats-section">
      <h3 class="section-title">近1天统计</h3>
      <div class="stats-grid">
        <StatCardVertical
          title="API 调用次数"
          :value="formatNumber(dayApiUsage.value)"
          :loading="dayApiUsageLoading"
          icon="api"
          color="#A8C686"
        />
        <StatCardVertical
          title="Token 使用量"
          :value="formatNumber(dayTokenUsage.value)"
          :loading="dayTokenUsageLoading"
          icon="token"
          color="#A8C686"
        />
        <StatCardVertical
          title="总费用"
          :value="'¥' + dayTotalCost.value.toFixed(2)"
          :loading="dayTotalCostLoading"
          icon="cost"
          color="#A8C686"
        />
      </div>
    </div>

    <!-- 近7天统计卡片 -->
    <div class="stats-section">
      <h3 class="section-title">近7天统计</h3>
      <div class="stats-grid">
        <StatCardVertical
          title="API 调用次数"
          :value="formatNumber(weekApiUsage.value)"
          :loading="weekApiUsageLoading"
          icon="api"
          color="#B8A99A"
        />
        <StatCardVertical
          title="Token 使用量"
          :value="formatNumber(weekTokenUsage.value)"
          :loading="weekTokenUsageLoading"
          icon="token"
          color="#B8A99A"
        />
        <StatCardVertical
          title="总费用"
          :value="'¥' + weekTotalCost.value.toFixed(2)"
          :loading="weekTotalCostLoading"
          icon="cost"
          color="#B8A99A"
        />
      </div>
    </div>

    <!-- 近30天统计卡片 -->
    <div class="stats-section">
      <h3 class="section-title">近30天统计</h3>
      <div class="stats-grid">
        <StatCardVertical
          title="API 调用次数"
          :value="formatNumber(monthApiUsage.value)"
          :loading="monthApiUsageLoading"
          icon="api"
          color="#C57272"
        />
        <StatCardVertical
          title="Token 使用量"
          :value="formatNumber(monthTokenUsage.value)"
          :loading="monthTokenUsageLoading"
          icon="token"
          color="#C57272"
        />
        <StatCardVertical
          title="总费用"
          :value="'¥' + monthTotalCost.value.toFixed(2)"
          :loading="monthTotalCostLoading"
          icon="cost"
          color="#C57272"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import StatCardVertical from '@/components/StatCardVertical.vue'
import { formatNumber } from '@/utils/formatters'

// 当月统计数据
const monthlyApiUsage = ref(0)
const monthlyTokenUsage = ref(0)
const monthlyTotalCost = ref(0)
const monthlyApiUsageLoading = ref(false)
const monthlyTokenUsageLoading = ref(false)
const monthlyTotalCostLoading = ref(false)

// 进度数据
const apiUsageProgress = ref({
  percentage: 0,
  current: 0,
  total: 0,
  color: '#4D6782'
})
const tokenUsageProgress = ref({
  percentage: 0,
  current: 0,
  total: 0,
  color: '#4D6782'
})
const totalCostProgress = ref({
  percentage: 0,
  current: 0,
  total: 0,
  color: '#4D6782'
})

// 近1天统计数据
const dayApiUsage = ref(0)
const dayTokenUsage = ref(0)
const dayTotalCost = ref(0)
const dayApiUsageLoading = ref(false)
const dayTokenUsageLoading = ref(false)
const dayTotalCostLoading = ref(false)

// 近7天统计数据
const weekApiUsage = ref(0)
const weekTokenUsage = ref(0)
const weekTotalCost = ref(0)
const weekApiUsageLoading = ref(false)
const weekTokenUsageLoading = ref(false)
const weekTotalCostLoading = ref(false)

// 近30天统计数据
const monthApiUsage = ref(0)
const monthTokenUsage = ref(0)
const monthTotalCost = ref(0)
const monthApiUsageLoading = ref(false)
const monthTokenUsageLoading = ref(false)
const monthTotalCostLoading = ref(false)

// 暴露数据和方法给父组件
defineExpose({
  // 当月数据
  monthlyApiUsage,
  monthlyTokenUsage,
  monthlyTotalCost,
  monthlyApiUsageLoading,
  monthlyTokenUsageLoading,
  monthlyTotalCostLoading,
  apiUsageProgress,
  tokenUsageProgress,
  totalCostProgress,
  
  // 近1天数据
  dayApiUsage,
  dayTokenUsage,
  dayTotalCost,
  dayApiUsageLoading,
  dayTokenUsageLoading,
  dayTotalCostLoading,
  
  // 近7天数据
  weekApiUsage,
  weekTokenUsage,
  weekTotalCost,
  weekApiUsageLoading,
  weekTokenUsageLoading,
  weekTotalCostLoading,
  
  // 近30天数据
  monthApiUsage,
  monthTokenUsage,
  monthTotalCost,
  monthApiUsageLoading,
  monthTokenUsageLoading,
  monthTotalCostLoading
})
</script>

<style scoped>
.stats-cards-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.stats-section {
  background: #FFFFFF;
  border: 1px solid rgba(77, 103, 130, 0.12);
  border-radius: 8px;
  padding: 20px;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.stats-section:hover {
  border-color: rgba(77, 103, 130, 0.2);
  box-shadow: 0 2px 8px rgba(77, 103, 130, 0.1);
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: #4D6782;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}

.section-title::before {
  content: '';
  width: 4px;
  height: 20px;
  background: #4D6782;
  border-radius: 2px;
  margin-right: 12px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 16px;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .stats-section {
    padding: 16px;
  }
}
</style>