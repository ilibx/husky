<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import request from '@/api/request'

const apiBaseURL = import.meta.env.VITE_API_BASE_URL || '/api'

const router = useRouter()
const route = useRoute()
const user = useUserStore()
const app = useAppStore()
const unreadCount = ref(0)
const menuItems = ref<any[]>([])
const menuLoading = ref(true)
let notificationSocket: WebSocket | undefined

function getNotificationSocketURL() {
  const token = localStorage.getItem('token') || ''
  const base = apiBaseURL.startsWith('http') ? apiBaseURL : `${window.location.origin}${apiBaseURL}`
  const url = new URL(`${base}/notifications/stream`)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.searchParams.set('token', token)
  return url.toString()
}

function connectNotificationSocket() {
  notificationSocket = new WebSocket(getNotificationSocketURL())
  notificationSocket.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.type === 'unread_count') {
        unreadCount.value = data.count ?? 0
      }
    } catch {
      unreadCount.value = 0
    }
  }
  notificationSocket.onerror = () => {
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

const activeMenu = computed(() => {
  const menuID = route.query.menu_id
  if (menuID) return String(menuID)
  const menu = flattenMenus(menuItems.value).find(item => item.path === route.path)
  return menu ? String(menu.id) : route.path
})

function flattenMenus(items: any[]): any[] {
  return items.flatMap(item => [item, ...flattenMenus(item.children || [])])
}

function handleMenuSelect(index: string) {
  const menu = flattenMenus(menuItems.value).find(item => String(item.id) === index)
  if (!menu) return
  if (menu.external) {
    if (menu.iframe) {
      router.push({ path: '/external-frame', query: { url: menu.path, menu_id: String(menu.id) } })
    } else {
      window.open(menu.path, '_blank', 'noopener,noreferrer')
    }
    return
  }
  if (menu.path) {
    router.push(menu.path)
  }
}

function goToNotifications() {
  router.push('/notifications')
}

onMounted(() => {
  fetchMenus()
  connectNotificationSocket()
})

onUnmounted(() => {
  notificationSocket?.close()
})

function handleLogout() {
  user.logout()
  router.push('/login')
}
</script>

<template>
  <el-container class="layout-container">
    <el-aside :width="app.sidebarCollapsed ? '64px' : '220px'" class="aside">
      <div class="logo">
        <span class="logo-icon">
          <el-icon :size="22"><Setting /></el-icon>
        </span>
        <span v-show="!app.sidebarCollapsed" class="logo-text">智能工单系统</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="app.sidebarCollapsed"
        :collapse-transition="false"
        background-color="transparent"
        text-color="#94a3b8"
        active-text-color="#ffffff"
        @select="handleMenuSelect"
      >
        <template v-for="item in menuItems" :key="item.id || item.name">
          <el-sub-menu v-if="item.children && item.children.length" :index="String(item.id)">
            <template #title>
              <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
              <span>{{ item.name }}</span>
            </template>
            <el-menu-item v-for="child in item.children" :key="child.id" :index="String(child.id)">
              <el-icon v-if="child.icon"><component :is="child.icon" /></el-icon>
              <template #title>{{ child.name }}</template>
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item v-else :index="String(item.id)">
            <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
            <template #title>{{ item.name }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="app.toggleSidebar" size="18">
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
              <el-avatar :size="30" :icon="'UserFilled'" class="user-avatar" />
              <span class="user-name">{{ user.userInfo?.nickname || user.userInfo?.username || '管理员' }}</span>
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
        <div class="page-wrapper">
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </div>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout-container {
  height: 100vh;
}

.aside {
  transition: width 0.25s ease;
  overflow-x: hidden;
  overflow-y: auto;
  background: var(--sidebar-bg);
  border-right: 1px solid rgba(255, 255, 255, 0.06);
}

.aside::-webkit-scrollbar {
  width: 4px;
}

.aside::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.12);
  border-radius: 2px;
}

.logo {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #fff;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 1px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.logo-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff;
  flex-shrink: 0;
}

.logo-text {
  white-space: nowrap;
  overflow: hidden;
}

.el-menu {
  border-right: none !important;
  padding: 8px 0;
}

.el-menu :deep(.el-menu-item),
.el-menu :deep(.el-sub-menu__title) {
  height: 44px;
  line-height: 44px;
  margin: 2px 8px;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.el-menu :deep(.el-menu-item):hover,
.el-menu :deep(.el-sub-menu__title):hover {
  background-color: rgba(255, 255, 255, 0.08) !important;
}

.el-menu :deep(.el-menu-item.is-active) {
  background-color: rgba(99, 102, 241, 0.15) !important;
  color: #fff !important;
}

.el-menu :deep(.el-sub-menu .el-menu) {
  background-color: transparent !important;
  padding: 0;
}

.el-menu :deep(.el-sub-menu .el-menu .el-menu-item) {
  padding-left: 48px !important;
  margin: 1px 8px;
}

.el-menu :deep(.el-menu-item .el-icon),
.el-menu :deep(.el-sub-menu__title .el-icon) {
  margin-right: 8px;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: var(--header-height);
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  cursor: pointer;
  color: var(--text-secondary);
  transition: color 0.2s;
  display: flex;
  align-items: center;
}

.collapse-btn:hover {
  color: var(--primary);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 8px;
  transition: background-color 0.2s;
}

.user-info:hover {
  background-color: #f1f5f9;
}

.user-avatar {
  background: linear-gradient(135deg, #6366f1, #8b5cf6) !important;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notif-icon {
  cursor: pointer;
  color: var(--text-secondary);
  transition: color 0.2s;
  display: flex;
  align-items: center;
}

.notif-icon:hover {
  color: var(--primary);
}

.notif-badge :deep(.el-badge__content) {
  top: 6px;
  right: 10px;
  border: 2px solid #fff;
}

.main {
  background: var(--bg-page);
  padding: 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.page-wrapper {
  flex: 1;
  width: 100%;
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
</style>
