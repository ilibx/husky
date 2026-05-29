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
        path: 'users',
        name: 'Users',
        component: () => import('@/views/users/UserList.vue'),
        meta: { title: '用户管理', icon: 'User' },
      },
      {
        path: 'sops',
        name: 'SOPs',
        component: () => import('@/views/sops/SOPList.vue'),
        meta: { title: 'SOP 管理', icon: 'List' },
      },
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
        meta: { title: '我的待办', icon: 'List' },
      },
      {
        path: 'knowledge',
        name: 'Knowledge',
        component: () => import('@/views/knowledge/KnowledgeList.vue'),
        meta: { title: '知识库管理', icon: 'Reading' },
      },
      {
        path: 'agents',
        name: 'Agents',
        component: () => import('@/views/agents/AgentList.vue'),
        meta: { title: 'Agent 管理', icon: 'Cpu' },
      },
      {
        path: 'categories',
        name: 'Categories',
        component: () => import('@/views/categories/CategoryList.vue'),
        meta: { title: '分类管理', icon: 'FolderOpened' },
      },
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
        path: 'tags',
        name: 'Tags',
        component: () => import('@/views/tags/TagList.vue'),
        meta: { title: '标签管理', icon: 'PriceTag' },
      },
      {
        path: 'channels',
        name: 'Channels',
        component: () => import('@/views/channels/ChannelList.vue'),
        meta: { title: '渠道配置', icon: 'Connection' },
      },
      {
        path: 'bot-config',
        name: 'BotConfig',
        component: () => import('@/views/botconfig/BotConfig.vue'),
        meta: { title: 'Bot 配置', icon: 'ChatDotSquare' },
      },
      {
        path: 'ticket-fields',
        name: 'TicketFields',
        component: () => import('@/views/ticketfields/TicketFields.vue'),
        meta: { title: '自定义字段', icon: 'Setting' },
      },
      {
        path: 'ldap',
        name: 'LDAP',
        component: () => import('@/views/ldap/LDAPSync.vue'),
        meta: { title: 'LDAP 同步', icon: 'RefreshRight' },
      },
      {
        path: 'sla-configs',
        name: 'SLAConfigs',
        component: () => import('@/views/slaconfig/SLAConfigList.vue'),
        meta: { title: 'SLA 配置', icon: 'Timer' },
      },
      {
        path: 'webhook-configs',
        name: 'WebhookConfigs',
        component: () => import('@/views/webhooks/WebhookConfigList.vue'),
        meta: { title: 'Webhook 配置', icon: 'Connection' },
      },
      {
        path: 'notifications',
        name: 'Notifications',
        component: () => import('@/views/notifications/NotificationList.vue'),
        meta: { title: '通知中心', icon: 'Bell' },
      },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  document.title = `${to.meta.title as string} - 工单管理系统`
  if (to.meta.noAuth) return next()
  if (!getToken()) return next('/login')
  next()
})

export default router
