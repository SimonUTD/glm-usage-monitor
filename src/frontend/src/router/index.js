import { createRouter, createWebHistory } from 'vue-router'
import { initializeAuth, checkAuth } from '@/composables/useApiState'

// 路由懒加载
const Settings = () => import('../views/Settings.vue')
const Bills = () => import('../views/Bills.vue')
const Stats = () => import('../views/Stats.vue')
const Sync = () => import('../views/Sync.vue')
const Onboarding = () => import('../views/Onboarding.vue')

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
  // 初始化认证状态
  await initializeAuth()

  // 访问根路径时，检查认证状态
  if (to.path === '/') {
    if (checkAuth()) {
      return next('/stats')
    } else {
      return next('/onboarding')
    }
  }

  // 访问 /onboarding 路径总是允许
  if (to.path === '/onboarding') {
    return next()
  }

  // 检查是否已认证
  if (checkAuth()) {
    return next()
  } else {
    // 没有认证，跳转到引导页
    return next('/onboarding')
  }
})

export default router
