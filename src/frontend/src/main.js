import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import router from './router'
import App from './App.vue'
import './style.css'

// 导入错误处理器
import { ErrorHandlerPlugin, errorStyles } from '@/utils/errorHandler'

// 全局注册ECharts组件
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart, BarChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent
} from 'echarts/components'
import VChart from 'vue-echarts'

use([
  CanvasRenderer,
  LineChart,
  PieChart,
  BarChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent
])

const app = createApp(App)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(ElementPlus, { locale: zhCn })
app.use(router)

// 使用错误处理器插件
app.use(ErrorHandlerPlugin, {
  enableNotification: true,
  enableConsoleLog: true,
  enableErrorReporting: false, // 可以在生产环境启用
  maxRetries: 3,
  retryDelay: 1000
})

// 添加错误处理样式
const style = document.createElement('style')
style.textContent = errorStyles
document.head.appendChild(style)

app.mount('#app')
