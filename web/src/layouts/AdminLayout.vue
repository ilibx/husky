<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'

const router = useRouter()
const route = useRoute()
const user = useUserStore()
const app = useAppStore()

interface MenuItem { path?: string; label: string; icon?: string; children?: MenuItem[] }
const menuItems: MenuItem[] = [
  { path: '/dashboard', label: '仪表盘', icon: 'Odometer' },
  { path: '/tickets', label: '工单管理', icon: 'Ticket' },
  { path: '/tasks', label: '我的待办', icon: 'List' },
  { path: '/notifications', label: '通知中心', icon: 'Bell' },
  { path: '/users', label: '用户管理', icon: 'User' },
  {
    label: '智能服务', icon: 'Cpu',
    children: [
      { path: '/knowledge', label: '知识库管理', icon: 'Reading' },
      { path: '/agents', label: 'Agent 管理', icon: 'Cpu' },
      { path: '/bot-config', label: 'Bot 配置', icon: 'ChatDotSquare' },
    ],
  },
  {
    label: '系统设置', icon: 'Setting',
    children: [
      { path: '/categories', label: '分类管理', icon: 'FolderOpened' },
      { path: '/departments', label: '部门管理', icon: 'OfficeBuilding' },
      { path: '/roles', label: '角色管理', icon: 'Key' },
      { path: '/tags', label: '标签管理', icon: 'PriceTag' },
      { path: '/channels', label: '渠道配置', icon: 'Connection' },
      { path: '/ticket-fields', label: '自定义字段', icon: 'Setting' },
      { path: '/ldap', label: 'LDAP 同步', icon: 'RefreshRight' },
      { path: '/sla-configs', label: 'SLA 配置', icon: 'Timer' },
      { path: '/webhook-configs', label: 'Webhook 配置', icon: 'Connection' },
    ],
  },
  { path: '/sops', label: 'SOP 管理', icon: 'List' },
]

const activeMenu = computed(() => route.path)

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
        :default-active="activeMenu"
        :collapse="app.sidebarCollapsed"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <template v-for="item in menuItems" :key="item.label">
          <el-sub-menu v-if="item.children" :index="item.label">
            <template #title>
              <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
              <span>{{ item.label }}</span>
            </template>
            <el-menu-item v-for="child in item.children" :key="child.path" :index="child.path!">
              <el-icon v-if="child.icon"><component :is="child.icon" /></el-icon>
              <template #title>{{ child.label }}</template>
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item v-else :index="item.path!">
            <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
            <template #title>{{ item.label }}</template>
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
  overflow: hidden;
  background: #304156;
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
.main {
  background: #f0f2f5;
}
</style>
