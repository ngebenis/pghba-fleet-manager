import axios from 'axios'
import type {
  LoginResponse, DBHost, CreateHostPayload, PGInstance, HBARule,
  CreateRulePayload, TempAccess, GrantTempPayload, HBABackup,
  BackupSchedule, BackupSchedulePayload,
  User, AuditListResponse, DashboardStats
} from '@/types'

const api = axios.create({
  baseURL: '/api',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' }
})

// Attach JWT on every request
api.interceptors.request.use(config => {
  const token = localStorage.getItem('access_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// Auto-refresh on 401
let isRefreshing = false
api.interceptors.response.use(
  res => res,
  async err => {
    const original = err.config
    if (err.response?.status === 401 && !original._retry) {
      original._retry = true
      if (!isRefreshing) {
        isRefreshing = true
        try {
          const rt = localStorage.getItem('refresh_token')
          const { data } = await axios.post('/api/auth/refresh', { refresh_token: rt })
          localStorage.setItem('access_token', data.access_token)
          api.defaults.headers.common['Authorization'] = `Bearer ${data.access_token}`
          isRefreshing = false
          return api(original)
        } catch {
          isRefreshing = false
          localStorage.clear()
          window.location.href = '/login'
        }
      }
    }
    return Promise.reject(err)
  }
)

// ── Auth ──────────────────────────────────────────────────────────────────────
export const authApi = {
  login: (username: string, password: string) =>
    api.post<LoginResponse>('/auth/login', { username, password }),
  logout: (refreshToken: string) =>
    api.post('/auth/logout', { refresh_token: refreshToken }),
  me: () => api.get<User>('/auth/me'),
  changePassword: (currentPassword: string, newPassword: string) =>
    api.post<{ access_token: string; expires_in: number }>('/auth/change-password', {
      current_password: currentPassword, new_password: newPassword,
    })
}

// ── Dashboard ─────────────────────────────────────────────────────────────────
export const dashboardApi = {
  stats: () => api.get<DashboardStats>('/dashboard/stats')
}

// ── Hosts ─────────────────────────────────────────────────────────────────────
export const hostsApi = {
  list: (params?: { environment?: string; active?: boolean }) =>
    api.get<DBHost[]>('/hosts', { params }),
  get: (id: string) => api.get<DBHost>(`/hosts/${id}`),
  create: (payload: CreateHostPayload) => api.post<DBHost>('/hosts', payload),
  update: (id: string, payload: Partial<CreateHostPayload>) =>
    api.put<DBHost>(`/hosts/${id}`, payload),
  delete: (id: string) => api.delete(`/hosts/${id}`),
  validate: (payload: { host_id?: string; ip_address: string; ssh_port?: number; ssh_user?: string; ssh_key?: string }) =>
    api.post<{ ok: boolean; message?: string; error?: string }>('/hosts/validate', payload),
  resetHostKey: (id: string) => api.post(`/hosts/${id}/reset-host-key`)
}

// ── Instances ─────────────────────────────────────────────────────────────────
export const instancesApi = {
  list: (params?: { host_id?: string; status?: string }) =>
    api.get<PGInstance[]>('/instances', { params }),
  discover: (hostId: string) => api.post('/instances/discover', { host_id: hostId })
}

// ── Rules ─────────────────────────────────────────────────────────────────────
export const rulesApi = {
  list: (params?: { instance_id?: string; rule_type?: string; enabled?: boolean }) =>
    api.get<HBARule[]>('/rules', { params }),
  create: (payload: CreateRulePayload) => api.post<HBARule>('/rules', payload),
  update: (id: string, payload: Partial<CreateRulePayload>) =>
    api.put<HBARule>(`/rules/${id}`, payload),
  delete: (id: string) => api.delete(`/rules/${id}`),
  enable: (id: string)  => api.post(`/rules/${id}/enable`),
  disable: (id: string) => api.post(`/rules/${id}/disable`)
}

// ── Temporary Access ──────────────────────────────────────────────────────────
export const tempAccessApi = {
  list: (params?: { instance_id?: string }) =>
    api.get<TempAccess[]>('/temporary-access', { params }),
  grant: (payload: GrantTempPayload) =>
    api.post('/temporary-access', payload),
  revoke: (id: string) => api.post(`/temporary-access/${id}/revoke`)
}

// ── Backups ───────────────────────────────────────────────────────────────────
export const backupsApi = {
  list: (params?: { instance_id?: string }) =>
    api.get<HBABackup[]>('/backups', { params }),
  create: (instanceId: string, note?: string) =>
    api.post('/backups', { instance_id: instanceId, note }),
  restore: (backupId: string) =>
    api.post('/backups/restore', { backup_id: backupId }),
  download: (id: string) =>
    api.get(`/backups/${id}/download`, { responseType: 'blob' })
}

// ── Backup Schedules ────────────────────────────────────────────────────────────
export const backupSchedulesApi = {
  list: () => api.get<BackupSchedule[]>('/backup-schedules'),
  create: (payload: BackupSchedulePayload) =>
    api.post<BackupSchedule>('/backup-schedules', payload),
  update: (id: string, payload: BackupSchedulePayload) =>
    api.put(`/backup-schedules/${id}`, payload),
  delete: (id: string) => api.delete(`/backup-schedules/${id}`)
}

// ── Users ─────────────────────────────────────────────────────────────────────
export const usersApi = {
  list: (params?: { role?: string }) => api.get<User[]>('/users', { params }),
  create: (payload: Partial<User> & { password: string }) =>
    api.post<User>('/users', payload),
  update: (id: string, payload: Partial<User>) =>
    api.put<User>(`/users/${id}`, payload),
  resetPassword: (id: string, newPassword: string) =>
    api.post(`/users/${id}/reset-password`, { new_password: newPassword }),
  delete: (id: string) => api.delete(`/users/${id}`)
}

// ── Audit ─────────────────────────────────────────────────────────────────────
export const auditApi = {
  list: (params?: {
    event?: string; username?: string; instance_id?: string
    from?: string; to?: string; limit?: number; page?: number
  }) => api.get<AuditListResponse>('/audit', { params })
}

export default api
