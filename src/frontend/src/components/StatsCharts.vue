<template>
  <div class="stats-charts-container">
    <!-- 24小时调用趋势 -->
    <div class="chart-section">
      <h3 class="section-title">24小时调用趋势</h3>
      <div class="chart-card chart-row">
        <UnifiedHourlyChart 
          :data="hourlyData" 
          @chart-click="handleChartClick" 
        />
      </div>
    </div>

    <!-- 产品分布饼图 -->
    <div class="chart-section">
      <h3 class="section-title">产品使用分布</h3>
      <div class="charts-grid">
        <div class="chart-card">
          <ProductPieEChart :data="productDistribution" />
        </div>
        <div class="chart-card">
          <ProductPieChart :data="productDistribution" />
        </div>
      </div>
    </div>

    <!-- 产品使用量柱状图 -->
    <div class="chart-section">
      <h3 class="section-title">产品使用量对比</h3>
      <div class="chart-card">
        <ProductBarEChart :data="productDistribution" />
      </div>
    </div>

    <!-- 时间段统计 -->
    <div class="time-sections">
      <!-- 近1天统计 -->
      <div class="time-section">
        <h3 class="section-title">近1天统计</h3>
        <div class="chart-card chart-row-other hourly-chart-row">
          <UnifiedHourlyChart 
            :data="hourlyDataDay" 
            title="每天调用次数和 Token 数量" 
            @chart-click="handleChartClick" 
          />
        </div>
      </div>

      <!-- 近7天统计 -->
      <div class="time-section">
        <h3 class="section-title">近7天统计</h3>
        <div class="chart-card chart-row-other weekly-chart-row">
          <UnifiedHourlyChart 
            :data="hourlyDataWeek" 
            title="每天调用次数和 Token 数量" 
            @chart-click="handleChartClick" 
          />
        </div>
      </div>

      <!-- 近30天统计 -->
      <div class="time-section">
        <h3 class="section-title">近30天统计</h3>
        <div class="chart-card chart-row-other monthly-chart-row">
          <UnifiedHourlyChart 
            :data="hourlyDataMonth" 
            title="每天调用次数和 Token 数量" 
            @chart-click="handleChartClick" 
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import UnifiedHourlyChart from '@/components/UnifiedHourlyChart.vue'
import ProductPieEChart from '@/components/ProductPieEChart.vue'
import ProductPieChart from '@/components/ProductPieChart.vue'
import ProductBarEChart from '@/components/ProductBarEChart.vue'

// 图表数据
const hourlyData = ref([])
const hourlyDataDay = ref([])
const hourlyDataWeek = ref([])
const hourlyDataMonth = ref([])
const productDistribution = ref([])

// 图表点击事件
const handleChartClick = (params) => {
  console.log('图表点击:', params)
  // 可以在这里添加图表点击后的处理逻辑
}

// 暴露数据给父组件
defineExpose({
  hourlyData,
  hourlyDataDay,
  hourlyDataWeek,
  hourlyDataMonth,
  productDistribution
})
</script>

<style scoped>
.stats-charts-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.chart-section, .time-section {
  background: #FFFFFF;
  border: 1px solid rgba(77, 103, 130, 0.12);
  border-radius: 8px;
  padding: 20px;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.chart-section:hover, .time-section:hover {
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

.chart-card {
  background: #FFFFFF;
  border: 1px solid rgba(77, 103, 130, 0.12);
  border-radius: 8px;
  padding: 16px;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
  min-height: 300px;
}

.chart-card:hover {
  border-color: rgba(77, 103, 130, 0.2);
  box-shadow: 0 2px 8px rgba(77, 103, 130, 0.1);
}

.chart-row {
  grid-column: 1 / -1;
}

.chart-row-other {
  grid-column: 1 / -1;
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 16px;
}

.time-sections {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

@media (max-width: 768px) {
  .charts-grid {
    grid-template-columns: 1fr;
  }
  
  .chart-section, .time-section {
    padding: 16px;
  }
  
  .chart-card {
    padding: 12px;
    min-height: 250px;
  }
}
</style>