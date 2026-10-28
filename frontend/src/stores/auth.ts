import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api'
import type { AuthUser } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(
    JSON.parse(localStorage.getItem('user') || 'null')
  )
  const accessToken  = ref(localStorage.getItem('access_token') || '')
  const refreshToken = ref(localStorage.getItem('refresh_token') || '')

  const isLoggedIn = computed(() => !!accessToken.value)
  const isAdmin    = computed(() => user.value?.role === 'SuperAdmin')
  const isDBA      = computed(() => ['SuperAdmin', 'DBA'].includes(user.value?.role ?? ''))
  const isOperator = computed(() => ['SuperAdmin', 'DBA', 'Operator', 'AccessManager'].includes(user.value?.role ?? ''))
  // Full CRUD on Access Rules & Temporary Access (including write-back to
  // the server), but no access to Hosts, Instances, Backups, Users, or
  // Settings — same edit-level rule permissions as DBA, scoped narrower.
  const canManageRules = computed(() => ['SuperAdmin', 'DBA', 'AccessManager'].includes(user.value?.role ?? ''))
  // AccessManager and TempAccessManager only see Dashboard + Access
  // Control in the sidebar — used to hide nav items/pages neither has
  // backend permission for anyway.
  const hidesAdminMenus = computed(() => ['AccessManager', 'TempAccessManager'].includes(user.value?.role ?? ''))
  // TempAccessManager can only grant/revoke Temporary Access — it doesn't
  // get the Access Rules page at all, unlike AccessManager.
  const canSeeAccessRules = computed(() => user.value?.role !== 'TempAccessManager')

  async function login(username: string, password: string) {
    const { data } = await authApi.login(username, password)
    accessToken.value  = data.access_token
    refreshToken.value = data.refresh_token
    user.value         = data.user
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    localStorage.setItem('user', JSON.stringify(data.user))
  }

  function applyPasswordChanged(newAccessToken: string) {
    accessToken.value = newAccessToken
    localStorage.setItem('access_token', newAccessToken)
    if (user.value) {
      user.value = { ...user.value, must_change_password: false }
      localStorage.setItem('user', JSON.stringify(user.value))
    }
  }

  async function logout() {
    try { await authApi.logout(refreshToken.value) } catch {}
    accessToken.value  = ''
    refreshToken.value = ''
    user.value = null
    localStorage.clear()
  }

  function hasRole(...roles: string[]) {
    return roles.includes(user.value?.role ?? '')
  }

  return { user, accessToken, isLoggedIn, isAdmin, isDBA, isOperator, canManageRules, hidesAdminMenus, canSeeAccessRules, login, logout, hasRole, applyPasswordChanged }
})
