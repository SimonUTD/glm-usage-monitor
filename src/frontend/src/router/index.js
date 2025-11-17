import { createRouter, createWebHistory } from 'vue-router'
import Settings from '../views/Settings.vue'
import Bills from '../views/Bills.vue'
import Stats from '../views/Stats.vue'
import Sync from '../views/Sync.vue'
import Onboarding from '../views/Onboarding.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Onboarding
  },
  {
    path: '/onboarding',
    name: 'Onboarding',
    component: Onboarding
  },
  {
    path: '/stats',
    name: 'Stats',
    component: Stats,
    meta: { requiresAuth: true }
  },
  {
    path: '/bills',
    name: 'Bills',
    component: Bills,
    meta: { requiresAuth: true }
  },
  {
    path: '/sync',
    name: 'Sync',
    component: Sync,
    meta: { requiresAuth: true }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: Settings,
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

import api from '../api'

router.beforeEach(async (to, from, next) => {
  // 访问根路径时，检查并验证数据库中的Token
  if (to.path === '/') {
    try {
      const result = await api.getToken()
      if (result.success && result.data && result.data.token_value) {
        // 数据库中有Token，验证Token是否有效
        const validation = await api.validateSavedToken()
        if (validation.success && validation.data) {
          // Token有效，保存到localStorage并跳转到统计页
          localStorage.setItem('api_token', result.data.token_value)
          return next('/stats')
        } else {
          // Token无效，跳转到引导页让用户重新输入
          return next('/onboarding')
        }
      } else {
        // 数据库中没有Token，跳转到引导页
        return next('/onboarding')
      }
    } catch (error) {
      console.error('检查Token失败:', error)
      // 查询失败也跳转到引导页
      return next('/onboarding')
    }
  }

  // 访问 /onboarding 路径总是允许
  if (to.path === '/onboarding') {
    return next()
  }

  // 检查是否已配置智谱AI Token
  try {
    const result = await api.getToken()
    if (result.success && result.data && result.data.token_value) {
      // 已配置Token，允许访问
      return next()
    }
  } catch (error) {
    console.error('检查Token失败:', error)
  }

  // 没有配置Token，跳转到引导页
  next('/onboarding')
})

export default router
