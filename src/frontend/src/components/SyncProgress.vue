<template>
  <div class="progress-container">
    <div class="progress-header">
      <span class="progress-stage">{{ progress.stage }}</span>
      <span class="progress-percentage">
        {{ progress.total > 0 && progress.current > 0 ? Math.floor((progress.current / progress.total) * 100) : progress.percentage }}%
      </span>
    </div>
    <el-progress
      :percentage="progress.total > 0 ? Math.floor((progress.current / progress.total) * 100) : progress.percentage"
      :stroke-width="10"
      :color="progressColor"
      :show-text="false"
    />
    <div class="progress-details">
      <span v-if="progress.total > 0">
        {{ progressType === 'incremental' ? `第 ${progress.current}/${progress.total} 页` : `${progress.current} / ${progress.total} 条记录` }}
        (已处理 {{ progress.syncedCount || 0 }} 条记录)
      </span>
      <span v-else>
        {{ progress.percentage }}%
      </span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  progress: {
    type: Object,
    required: true,
    default: () => ({
      percentage: 0,
      current: 0,
      total: 0,
      stage: '处理中...',
      syncedCount: 0
    })
  },
  progressType: {
    type: String,
    default: 'full', // 'full' or 'incremental'
    validator: (value) => ['full', 'incremental'].includes(value)
  }
})

const progressColor = computed(() => {
  return props.progressType === 'incremental' ? '#4D6782' : '#C57272'
})
</script>

<style scoped>
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
</style>