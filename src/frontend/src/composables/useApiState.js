/**
 * API状态管理组合式函数
 * 统一管理Token和用户状态，提供全局状态管理
 */
import { ref, computed, watch } from 'vue'
import api from '@/api'

// 全局状态
const token = ref('')
const isAuthenticated = ref(false)
const currentUser = ref(null)
const membershipTier = ref('GLM Coding Pro')
const autoSyncConfig = ref({
  enabled: false,
  frequencySeconds: 0
})

// 计算属性
const isLoggedIn = computed(() => isAuthenticated.value && !!token.value)

/**
 * 初始化认证状态
 */
export const initializeAuth = async () => {
  try {
    // 从localStorage恢复token
    const savedToken = localStorage.getItem('api_token')
    if (savedToken) {
      token.value = savedToken
      isAuthenticated.value = true
      
      // 验证token有效性
      const validation = await api.validateSavedToken()
      if (!validation.success || !validation.data) {
        // token无效，清除状态
        clearAuth()
        return false
      }
    }
    
    // 从数据库获取最新token
    const result = await api.getToken()
    if (result.success && result.data && result.data.token) {
      token.value = result.data.token
      isAuthenticated.value = true
      localStorage.setItem('api_token', result.data.token)
    }
    
    return isAuthenticated.value
  } catch (error) {
    console.error('初始化认证状态失败:', error)
    clearAuth()
    return false
  }
}

/**
 * 设置认证信息
 */
export const setAuth = (authToken, user = null) => {
  token.value = authToken
  isAuthenticated.value = true
  currentUser.value = user
  localStorage.setItem('api_token', authToken)
}

/**
 * 清除认证信息
 */
export const clearAuth = () => {
  token.value = ''
  isAuthenticated.value = false
  currentUser.value = null
  localStorage.removeItem('api_token')
}

/**
 * 获取当前Token
 */
export const getToken = () => token.value

/**
 * 检查是否已登录
 */
export const checkAuth = () => isAuthenticated.value

/**
 * 加载会员等级信息
 */
export const loadMembershipTier = async () => {
  try {
    const result = await api.getCurrentMembershipTier()
    if (result.success && result.data?.membershipTier) {
      membershipTier.value = result.data.membershipTier
    }
  } catch (error) {
    console.error('获取会员等级失败:', error)
  }
}

/**
 * 加载自动同步配置
 */
export const loadAutoSyncConfig = async () => {
  try {
    const result = await api.getAutoSyncConfig()
    if (result.success && result.data) {
      autoSyncConfig.value = {
        enabled: result.data.enabled || false,
        frequencySeconds: result.data.frequency_seconds || 0
      }
    }
  } catch (error) {
    console.error('加载自动同步配置失败:', error)
    autoSyncConfig.value = {
      enabled: false,
      frequencySeconds: 0
    }
  }
}

/**
 * 更新自动同步配置
 */
export const updateAutoSyncConfig = (config) => {
  autoSyncConfig.value = { ...autoSyncConfig.value, ...config }
}

/**
 * 监听认证状态变化
 */
export const watchAuth = (callback) => {
  watch([isAuthenticated, token], ([auth, authToken]) => {
    callback(auth, authToken)
  })
}

/**
 * 监听自动同步配置变化
 */
export const watchAutoSyncConfig = (callback) => {
  watch(autoSyncConfig, (config) => {
    callback(config)
  }, { deep: true })
}

// 导出状态
export {
  token,
  isAuthenticated,
  currentUser,
  membershipTier,
  autoSyncConfig,
  isLoggedIn
}