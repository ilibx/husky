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
