import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { getToken } from '@/utils/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/LoginView.vue'),
    meta: { title: '登录', noAuth: true },
  },
  {
    path: '/',
    component: () => import('@/layouts/AdminLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/DashboardView.vue'),
        meta: { title: '仪表盘', icon: 'Odometer' },
      },
      {
        path: 'external-frame',
        name: 'ExternalFrame',
        component: () => import('@/views/external/ExternalFrame.vue'),
        meta: { title: '外部页面' },
      },
      // ---- 工单管理 ----
      {
        path: 'tickets',
        name: 'Tickets',
        component: () => import('@/views/tickets/TicketList.vue'),
        meta: { title: '工单管理', icon: 'Ticket' },
      },
      {
        path: 'tasks',
        name: 'Tasks',
        component: () => import('@/views/tasks/TaskList.vue'),
        meta: { title: '我的待办' },
      },
      // ---- 智能服务 ----
      {
        path: 'agents',
        name: 'Agents',
        component: () => import('@/views/agents/AgentList.vue'),
        meta: { title: 'Agent 管理', icon: 'Cpu' },
      },
      {
        path: 'sops',
        name: 'SOPs',
        component: () => import('@/views/sops/SOPList.vue'),
        meta: { title: 'SOP 管理' },
      },
      {
        path: 'skills',
        name: 'Skills',
        component: () => import('@/views/skills/SkillManagement.vue'),
        meta: { title: 'Skill 管理', icon: 'MagicStick' },
      },
      {
        path: 'mcps',
        name: 'MCPs',
        component: () => import('@/views/mcps/MCPManagement.vue'),
        meta: { title: 'MCP 服务', icon: 'Cpu' },
      },
      // ---- 资料管理 ----
      {
        path: 'knowledge',
        redirect: '/knowledge/search',
        children: [
          {
            path: 'search',
            name: 'KnowledgeSearch',
            component: () => import('@/views/knowledge/KnowledgeSearch.vue'),
            meta: { title: '资料检索' },
          },
          {
            path: 'docs',
            name: 'KnowledgeDocs',
            component: () => import('@/views/knowledge/KnowledgeDocs.vue'),
            meta: { title: '文档库管理' },
          },
          {
            path: 'config',
            name: 'KnowledgeConfig',
            component: () => import('@/views/knowledge/KnowledgeConfig.vue'),
            meta: { title: '知识库管理' },
          },
        ],
      },
      // ---- 系统管理 ----
      {
        path: 'departments',
        name: 'Departments',
        component: () => import('@/views/departments/DepartmentList.vue'),
        meta: { title: '部门管理', icon: 'OfficeBuilding' },
      },
      {
        path: 'roles',
        name: 'Roles',
        component: () => import('@/views/roles/RoleList.vue'),
        meta: { title: '角色管理', icon: 'Key' },
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/users/UserList.vue'),
        meta: { title: '用户管理', icon: 'User' },
      },
      {
        path: 'menus',
        name: 'Menus',
        component: () => import('@/views/menus/MenuManagement.vue'),
        meta: { title: '菜单管理', icon: 'Menu' },
      },
      {
        path: 'categories',
        name: 'Categories',
        component: () => import('@/views/categories/CategoryList.vue'),
        meta: { title: '分类管理', icon: 'FolderOpened' },
      },
      // ---- 通知管理 ----
      {
        path: 'notifications',
        name: 'Notifications',
        component: () => import('@/views/notifications/NotificationList.vue'),
        meta: { title: '通知列表', icon: 'Bell' },
      },
      {
        path: 'push-config',
        name: 'PushConfig',
        component: () => import('@/views/pushconfig/PushManagement.vue'),
        meta: { title: '推送管理' },
      },
      {
        path: 'notification-levels',
        name: 'NotificationLevels',
        component: () => import('@/views/notificationlevels/LevelManagement.vue'),
        meta: { title: '等级管理' },
      },
      {
        path: 'sla-configs',
        name: 'SLAConfigs',
        component: () => import('@/views/slaconfig/SLAConfigList.vue'),
        meta: { title: 'SLA 配置' },
      },
      // ---- 渠道管理 ----
      {
        path: 'channels',
        redirect: '/channels/list',
        children: [
          {
            path: 'list',
            name: 'Channels',
            component: () => import('@/views/channels/ChannelList.vue'),
            meta: { title: '渠道列表', icon: 'Connection' },
          },
        ],
      },
      // ---- 模型管理 ----
      {
        path: 'models',
        redirect: '/models/platforms',
        children: [
          {
            path: 'platforms',
            name: 'ModelPlatforms',
            component: () => import('@/views/models/ModelPlatform.vue'),
            meta: { title: '平台管理', icon: 'Monitor' },
          },
          {
            path: 'configs',
            name: 'ModelConfigs',
            component: () => import('@/views/models/ModelConfig.vue'),
            meta: { title: '模型配置', icon: 'Monitor' },
          },
        ],
      },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to, _from, next) => {
  document.title = `${to.meta.title as string} - 智能工单系统`
  if (to.meta.noAuth) return next()
  if (!getToken()) return next('/login')
  // Load user info if not loaded yet
  const { useUserStore } = await import('@/stores/user')
  const user = useUserStore()
  if (!user.userInfo) {
    try {
      await user.fetchUserInfo()
    } catch {
      user.logout()
      return next('/login')
    }
  }
  next()
})

export default router
