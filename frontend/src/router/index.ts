import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { public: true }
    },
    {
      path: '/change-password',
      name: 'force-password-change',
      component: () => import('@/views/ForcePasswordChangeView.vue')
    },
    {
      path: '/',
      component: () => import('@/components/layout/AppLayout.vue'),
      children: [
        { path: '',          redirect: '/dashboard' },
        { path: 'dashboard', name: 'dashboard', component: () => import('@/views/DashboardView.vue') },
        { path: 'hosts',     name: 'hosts',     component: () => import('@/views/HostsView.vue'), meta: { roles: ['SuperAdmin', 'DBA'] } },
        { path: 'instances', name: 'instances', component: () => import('@/views/InstancesView.vue'), meta: { roles: ['SuperAdmin', 'DBA'] } },
        // Explicit role list so TempAccessManager — which should only see
        // Temporary Access, not Access Rules — can be excluded without
        // changing access for anyone who could already reach this page.
        { path: 'rules',     name: 'rules',     component: () => import('@/views/RulesView.vue'), meta: { roles: ['SuperAdmin', 'DBA', 'Operator', 'Viewer', 'AccessManager'] } },
        { path: 'temp-access', name: 'temp-access', component: () => import('@/views/TempAccessView.vue'), meta: { roles: ['SuperAdmin', 'DBA', 'Operator', 'AccessManager', 'TempAccessManager'] } },
        { path: 'backup',    name: 'backup',    component: () => import('@/views/BackupView.vue'), meta: { roles: ['SuperAdmin', 'DBA'] } },
        { path: 'users',     name: 'users',     component: () => import('@/views/UsersView.vue'), meta: { roles: ['SuperAdmin'] } },
        { path: 'audit',     name: 'audit',     component: () => import('@/views/AuditView.vue'), meta: { roles: ['SuperAdmin', 'DBA'] } },
        // Explicit role list (rather than "any authenticated user") so
        // AccessManager — which should only see Dashboard + Access
        // Control — can be excluded without changing access for anyone
        // who could already reach Settings.
        { path: 'settings',  name: 'settings',  component: () => import('@/views/SettingsView.vue'), meta: { roles: ['SuperAdmin', 'DBA', 'Operator', 'Viewer'] } },
      ]
    },
    { path: '/:pathMatch(.*)*', redirect: '/dashboard' }
  ]
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.public) return next()
  if (!auth.isLoggedIn) return next('/login')
  // Account still uses the published default password (see migration 006)
  // — block every other page until it's changed, same as the backend does.
  if (auth.user?.must_change_password && to.name !== 'force-password-change') {
    return next('/change-password')
  }
  const roles = to.meta.roles as string[] | undefined
  if (roles && !auth.hasRole(...roles)) return next('/dashboard')
  next()
})

export default router
