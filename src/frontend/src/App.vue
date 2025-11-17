<template>
  <div class="app-container">
    <header class="app-header" v-show="!isOnboardingPage">
      <div class="header-content">
        <div class="logo">
          <div class="logo-icon">
            <el-icon><Monitor /></el-icon>
          </div>
          <h1>智谱AI GLM Coding Plan 账单统计</h1>
        </div>
        <nav class="top-nav">
          <el-menu
            mode="horizontal"
            :default-active="activeMenu"
            class="top-menu"
            router
            background-color="transparent"
          >
            <el-menu-item index="/stats">
              <el-icon><DataAnalysis /></el-icon>
              <span>统计信息</span>
            </el-menu-item>
            <el-menu-item index="/bills">
              <el-icon><Document /></el-icon>
              <span>账单列表</span>
            </el-menu-item>
            <el-menu-item index="/sync">
              <el-icon><Refresh /></el-icon>
              <span>数据同步</span>
            </el-menu-item>
            <el-menu-item index="/settings">
              <el-icon><Setting /></el-icon>
              <span>设置</span>
            </el-menu-item>
          </el-menu>
        </nav>
      </div>
    </header>

    <main class="app-main" :class="{ 'onboarding-mode': isOnboardingPage }">
      <div class="main-content">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { useRoute } from 'vue-router'
import { Monitor, Setting, Document, DataAnalysis, Refresh } from '@element-plus/icons-vue'

const route = useRoute()
const activeMenu = ref(route.path)

watch(() => route.path, (newPath) => {
  activeMenu.value = newPath
})

const isOnboardingPage = computed(() => {
  return route.path === '/onboarding'
})

const openGithubRepo = () => {
  window.open('https://github.com/pxvp2008/AreYouOk', '_blank')
}
</script>

<style>
  html,body{
    width: 100%;
    height: 100%;
    margin: 0;
  }
</style>

<style scoped>
.app-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f5f3 0%, #e8e6e1 100%);
  position: relative;
  padding: 0;
  margin: 0;
  overflow-x: hidden;
}

.app-header {
  background: rgba(255, 255, 255, 0.75);
  border-bottom: 1px solid rgba(77, 103, 130, 0.15);
  box-shadow: 0 1px 4px rgba(77, 103, 130, 0.06);
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
  will-change: transform;
  transform: translateZ(0);
}

.header-content {
  max-width: 100%;
  padding: 0 20px;
  display: flex;
  align-items: center;
  height: 70px;
  box-sizing: border-box;
  margin: 0 auto;
  position: relative;
}

@media (min-width: 768px) {
  .header-content {
    padding: 0 40px;
  }
}

.logo {
  display: flex;
  align-items: center;
  margin-right: 60px;
}

.logo-icon {
  width: 45px;
  height: 45px;
  background: linear-gradient(135deg, #d4c5b9 0%, #b8a99a 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}

.logo-icon .el-icon {
  font-size: 24px;
  color: #FFFFFF;
}

.logo h1 {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #4D6782;
  letter-spacing: 0.5px;
}

.top-nav {
  flex: 1;
}

.top-menu {
  border-bottom: none;
}

.top-menu .el-menu-item {
  height: 70px;
  line-height: 70px;
  font-size: 15px;
  font-weight: 500;
  color: #6B7280;
  border-bottom: 3px solid transparent;
  position: relative;
  transition: all 0.3s ease;
  background: transparent;
}

.top-menu .el-menu-item::before {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  width: 0;
  height: 2px;
  background: #a8b5c0;
  transition: width 0.2s ease;
  transform: translateX(-50%);
}

.top-menu .el-menu-item:hover {
  background: rgba(77, 103, 130, 0.05);
  color: #4D6782;
}

.top-menu .el-menu-item:hover::before {
  width: 80%;
}

.top-menu .el-menu-item.is-active {
  background: rgba(77, 103, 130, 0.08);
  color: #4D6782;
}

.top-menu .el-menu-item.is-active::before {
  width: 100%;
  background: #a8b5c0;
}

.github-float {
  position: absolute;
  right: 40px;
  top: 50%;
  transform: translateY(-50%);
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 2px 8px rgba(77, 103, 130, 0.15);
  transition: all 0.3s ease;
  z-index: 10;
}

.github-float:hover {
  background: rgba(77, 103, 130, 0.1);
  box-shadow: 0 4px 12px rgba(77, 103, 130, 0.25);
  transform: translateY(-50%) scale(1.05);
}

.github-icon {
  width: 24px;
  height: 24px;
  fill: #4D6782;
  transition: fill 0.05s ease;
}

.github-float:hover .github-icon {
  fill: #4D6782;
}

@media (max-width: 768px) {
  .github-float {
    right: 20px;
    width: 35px;
    height: 35px;
  }

  .github-icon {
    width: 20px;
    height: 20px;
  }
}

.top-menu .el-menu-item .el-icon {
  margin-right: 8px;
  font-size: 18px;
}

.app-main {
  position: relative;
  z-index: 1;
  padding: 75px 0 0 0;
  min-height: 100vh;
  width: 100%;
  background: #f5f5f3;
}

.app-main.onboarding-mode {
  padding: 0;
}

.main-content {
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

@media (min-width: 768px) {
  .main-content {
    padding: 10px 15px;
  }
}

@media (min-width: 1200px) {
  .main-content {
    padding: 10px 15px;
  }
}
</style>
