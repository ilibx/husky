<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import request from '@/api/request'

const router = useRouter()
const route = useRoute()
const user = useUserStore()
const app = useAppStore()
const unreadCount = ref(0)
const menuItems = ref<any[]>([])
const menuLoading = ref(true)
let pollTimer: ReturnType<typeof setInterval> | undefined

async function fetchUnreadCount() {
  try {
    const res: any = await request.get('/notifications/unread-count')
    unreadCount.value = res.count ?? res.data?.count ?? 0
  } catch {
    unreadCount.value = 0
  }
}

async function fetchMenus() {
  try {
    const res: any = await request.get('/menus')
    menuItems.value = res.data?.data || []
  } catch {
    menuItems.value = []
  } finally {
    menuLoading.value = false
  }
}

function goToNotifications() {
  router.push('/notifications')
}

onMounted(() => {
  fetchMenus()
  fetchUnreadCount()
  pollTimer = setInterval(fetchUnreadCount, 30000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

function handleLogout() {
  user.logout()
  router.push('/login')
}
</script>

<template>
  <el-container style="height: 100vh">
    <el-aside :width="app.sidebarCollapsed ? '64px' : '220px'" class="aside">
      <div class="logo">{{ app.sidebarCollapsed ? 'HA' : '工单管理' }}</div>
      <el-menu
        :default-active="route.path"
        :collapse="app.sidebarCollapsed"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <template v-for="item in menuItems" :key="item.id || item.name">
          <el-sub-menu v-if="item.children && item.children.length" :index="item.name">
            <template #title>
              <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
              <span>{{ item.name }}</span>
            </template>
            <el-menu-item v-for="child in item.children" :key="child.path" :index="child.path!">
              <el-icon v-if="child.icon"><component :is="child.icon" /></el-icon>
              <template #title>{{ child.name }}</template>
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item v-else :index="item.path!">
            <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
            <template #title>{{ item.name }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="app.toggleSidebar" size="20">
            <Fold v-if="!app.sidebarCollapsed" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="route.meta.title">{{ route.meta.title as string }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-badge :value="unreadCount" :hidden="unreadCount === 0" class="notif-badge">
            <el-icon size="20" class="notif-icon" @click="goToNotifications">
              <Bell />
            </el-icon>
          </el-badge>
          <el-dropdown trigger="click">
            <span class="user-info">
              <el-avatar :size="28" :icon="'UserFilled'" />
              {{ user.userInfo?.nickname || user.userInfo?.username || '管理员' }}
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="handleLogout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.aside {
  transition: width 0.3s;
  overflow-x: hidden;
  overflow-y: auto;
  background: #304156;
}
.aside::-webkit-scrollbar {
  width: 4px;
}
.aside::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
}
.logo {
  height: 56px;
  line-height: 56px;
  text-align: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  letter-spacing: 2px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background: #fff;
  border-bottom: 1px solid #e6e6e6;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}
.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}
.collapse-btn {
  cursor: pointer;
}
.header-right {
  display: flex;
  align-items: center;
}
.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.notif-icon {
  cursor: pointer;
  margin-right: 16px;
  color: #666;
}
.notif-icon:hover {
  color: #409eff;
}
.notif-badge :deep(.el-badge__content) {
  top: 8px;
  right: 14px;
}
.main {
  background: #f0f2f5;
}
</style>
