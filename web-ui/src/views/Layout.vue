<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '240px'" class="sidebar">
      <div class="logo">
        <el-icon v-if="isCollapse" :size="32" color="#409EFF">
          <Shield />
        </el-icon>
        <template v-else>
          <el-icon :size="32" color="#409EFF">
            <Shield />
          </el-icon>
          <span class="logo-text">Cap Agent</span>
        </template>
      </div>
      
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :unique-opened="true"
        router
        class="sidebar-menu"
      >
        <el-menu-item
          v-for="route in menuRoutes"
          :key="route.path"
          :index="route.path"
        >
          <el-icon>
            <component :is="route.meta.icon" />
          </el-icon>
          <template #title>{{ route.meta.title }}</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <!-- 顶部导航 -->
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-icon" @click="toggleCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>Cap Agent</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentRoute?.meta?.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        
        <div class="header-right">
          <!-- 系统状态 -->
          <div class="system-status">
            <el-badge :value="alertCount" :max="99" class="badge-item">
              <el-icon :size="20">
                <Bell />
              </el-icon>
            </el-badge>
            <el-tag :type="systemStatus.type" size="small" class="status-tag">
              {{ systemStatus.text }}
            </el-tag>
          </div>
          
          <!-- 主题切换 -->
          <el-switch
            v-model="isDark"
            inline-prompt
            :active-icon="Moon"
            :inactive-icon="Sunny"
            @change="toggleTheme"
            class="theme-switch"
          />
          
          <!-- 用户菜单 -->
          <el-dropdown>
            <div class="user-avatar">
              <el-avatar :size="32" :icon="UserFilled" />
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item>个人中心</el-dropdown-item>
                <el-dropdown-item divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区域 -->
      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Shield, Fold, Expand, Bell, Moon, Sunny, UserFilled } from '@element-plus/icons-vue'
import { useAlertStore } from '@/stores/alert'

const route = useRoute()
const router = useRouter()
const alertStore = useAlertStore()

const isCollapse = ref(false)
const isDark = ref(false)
const alertCount = ref(12)

const systemStatus = computed(() => {
  return {
    text: 'Zeek运行中',
    type: 'success'
  }
})

const menuRoutes = computed(() => {
  return router.options.routes[0].children || []
})

const activeMenu = computed(() => route.path)
const currentRoute = computed(() => route)

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const toggleTheme = (value) => {
  if (value) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}

onMounted(() => {
  alertStore.connectWebSocket()
})
</script>

<style lang="scss" scoped>
.layout-container {
  height: 100vh;
  
  .sidebar {
    background: linear-gradient(180deg, #001529 0%, #002140 100%);
    transition: width 0.3s;
    box-shadow: 2px 0 8px rgba(0, 0, 0, 0.15);
    
    .logo {
      height: 60px;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 12px;
      padding: 0 20px;
      border-bottom: 1px solid rgba(255, 255, 255, 0.1);
      
      .logo-text {
        font-size: 20px;
        font-weight: 600;
        color: #fff;
        letter-spacing: 1px;
      }
    }
    
    .sidebar-menu {
      border: none;
      background: transparent;
      
      :deep(.el-menu-item) {
        color: rgba(255, 255, 255, 0.8);
        
        &:hover {
          background: rgba(255, 255, 255, 0.1);
          color: #fff;
        }
        
        &.is-active {
          background: #409EFF;
          color: #fff;
        }
      }
    }
  }
  
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: #fff;
    border-bottom: 1px solid #f0f0f0;
    padding: 0 24px;
    box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
    
    .header-left {
      display: flex;
      align-items: center;
      gap: 20px;
      
      .collapse-icon {
        font-size: 20px;
        cursor: pointer;
        transition: color 0.3s;
        
        &:hover {
          color: #409EFF;
        }
      }
    }
    
    .header-right {
      display: flex;
      align-items: center;
      gap: 24px;
      
      .system-status {
        display: flex;
        align-items: center;
        gap: 12px;
        
        .badge-item {
          cursor: pointer;
        }
        
        .status-tag {
          font-weight: 500;
        }
      }
      
      .theme-switch {
        --el-switch-on-color: #409EFF;
      }
      
      .user-avatar {
        cursor: pointer;
      }
    }
  }
  
  .main-content {
    background: #f5f7fa;
    padding: 24px;
    overflow-y: auto;
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s, transform 0.3s;
}

.fade-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.fade-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

// 暗黑模式适配
.dark {
  .header {
    background: #141414;
    border-bottom-color: #303030;
    color: rgba(255, 255, 255, 0.85);
  }
  
  .main-content {
    background: #000;
  }
}
</style>
