import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/auth/Login.vue'),
      meta: { 
        requiresAuth: false,
        title: '登录 - MySQLer'
      }
    },
    {
      path: '/',
      component: () => import('@/views/layout/MainLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'Dashboard',
          component: () => import('@/views/dashboard/Dashboard.vue'),
          meta: { 
            title: '仪表板 - MySQLer'
          }
        },
        {
          path: '/connections',
          name: 'Connections',
          component: () => import('@/views/connections/ConnectionList.vue'),
          meta: { 
            title: '连接管理 - MySQLer'
          }
        },
        {
          path: '/execution',
          name: 'Execution',
          component: () => import('@/views/execution/ExecutionPage.vue'),
          meta: { 
            title: 'DDL执行 - MySQLer'
          }
        },
        {
          path: '/history',
          name: 'History',
          component: () => import('@/views/history/HistoryList.vue'),
          meta: { 
            title: '执行历史 - MySQLer'
          }
        },
        {
          path: '/admin',
          name: 'Admin',
          component: () => import('@/views/admin/AdminPanel.vue'),
          meta: { 
            title: '系统管理 - MySQLer',
            requiresRole: 'admin'
          }
        }
      ]
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/error/NotFound.vue'),
      meta: { 
        title: '页面未找到 - MySQLer'
      }
    }
  ]
})

// 路由守卫
router.beforeEach((to) => {
  const authStore = useAuthStore()
  
  // 设置页面标题
  if (to.meta.title) {
    document.title = to.meta.title as string
  }
  
  // 检查认证
  if (to.meta.requiresAuth !== false && !authStore.isAuthenticated) {
    return { name: 'Login', query: { redirect: to.fullPath } }
  }
  
  // 检查角色权限
  if (to.meta.requiresRole && authStore.user?.role !== to.meta.requiresRole) {
    // 权限不足，跳转到首页
    return { name: 'Dashboard' }
  }
  
  // 已登录用户访问登录页，跳转到首页
  if (to.name === 'Login' && authStore.isAuthenticated) {
    return { name: 'Dashboard' }
  }
})

export default router