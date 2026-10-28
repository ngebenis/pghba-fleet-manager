<template>
  <div class="page-content">
    <div class="toolbar">
      <select class="filter-select" v-model="filterRole" @change="load">
        <option value="">All Roles</option>
        <option>SuperAdmin</option>
        <option>DBA</option>
        <option>Operator</option>
        <option>Viewer</option>
        <option>AccessManager</option>
        <option>TempAccessManager</option>
      </select>
      <select class="filter-select" v-model="filterStatus" @change="load">
        <option value="">All Status</option>
        <option>Active</option>
        <option>Inactive</option>
      </select>
      <div class="toolbar-right">
        <button class="btn btn-primary" @click="openCreate">
          <i class="ti ti-user-plus"></i> Add User
        </button>
      </div>
    </div>

    <!-- RBAC summary -->
    <div class="rbac-grid">
      <div v-for="role in roles" :key="role.name" class="rbac-card">
        <div class="rbac-header">
          <span :class="['badge', role.badge]">{{ role.name }}</span>
          <span class="rbac-count">{{ countByRole(role.name) }} user{{ countByRole(role.name) !== 1 ? 's' : '' }}</span>
        </div>
        <ul class="rbac-perms">
          <li v-for="p in role.perms" :key="p"><i class="ti ti-check" style="color:var(--c-green);font-size:11px"></i> {{ p }}</li>
        </ul>
      </div>
    </div>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Username</th><th>Full Name</th><th>Email</th>
            <th>Role</th><th>Last Login</th><th>Status</th><th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="7" style="text-align:center;padding:24px;color:var(--c-muted)">
              <i class="ti ti-loader-2 spin"></i> Loading...
            </td>
          </tr>
          <tr v-for="user in users" :key="user.id">
            <td>
              <div style="display:flex;align-items:center;gap:8px">
                <div class="user-avatar">{{ initials(user) }}</div>
                <span style="font-weight:600">{{ user.username }}</span>
              </div>
            </td>
            <td>{{ user.full_name }}</td>
            <td style="color:var(--c-muted);font-size:12px">{{ user.email }}</td>
            <td><span :class="['badge', roleBadge(user.role)]">{{ user.role }}</span></td>
            <td style="font-size:11px;color:var(--c-muted)">
              {{ user.last_login ? fmt(user.last_login) : 'Never' }}
            </td>
            <td>
              <button
                :class="['toggle', user.status === 'Active' ? 'on' : 'off']"
                @click="toggleStatus(user)"
                :title="user.status === 'Active' ? 'Deactivate' : 'Activate'"
              ></button>
            </td>
            <td>
              <div style="display:flex;gap:4px">
                <button class="btn btn-ghost btn-sm" @click="openEdit(user)" title="Edit">
                  <i class="ti ti-edit"></i>
                </button>
                <button class="btn btn-ghost btn-sm" @click="openResetPw(user)" title="Reset password">
                  <i class="ti ti-key"></i>
                </button>
                <button class="btn btn-danger btn-sm" @click="deleteUser(user)" title="Delete"
                  :disabled="user.username === 'admin'">
                  <i class="ti ti-trash"></i>
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && users.length === 0">
            <td colspan="7" style="text-align:center;padding:32px;color:var(--c-muted)">No users found</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add/Edit Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
        <div class="modal">
          <div class="modal-header">
            <div class="modal-title">
              <i :class="['ti', editing ? 'ti-user-edit' : 'ti-user-plus']"></i>
              {{ editing ? 'Edit User' : 'Add User' }}
            </div>
            <button class="btn btn-ghost btn-sm" @click="showModal = false"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body">
            <div class="field-row">
              <div class="field">
                <label>Username *</label>
                <input v-model="form.username" placeholder="jdoe" :disabled="!!editing">
              </div>
              <div class="field">
                <label>Full Name *</label>
                <input v-model="form.full_name" placeholder="John Doe">
              </div>
            </div>
            <div class="field">
              <label>Email *</label>
              <input v-model="form.email" type="email" placeholder="jdoe@company.com">
            </div>
            <div class="field-row">
              <div class="field">
                <label>Role *</label>
                <select v-model="form.role">
                  <option>DBA</option>
                  <option>Operator</option>
                  <option>Viewer</option>
                  <option>SuperAdmin</option>
                  <option>AccessManager</option>
                  <option>TempAccessManager</option>
                </select>
              </div>
              <div class="field" v-if="!editing">
                <label>Password *</label>
                <input v-model="form.password" type="password" placeholder="Min 8 characters">
              </div>
            </div>

            <!-- Role description -->
            <div class="role-hint" v-if="form.role">
              <i class="ti ti-info-circle"></i>
              {{ roleHint(form.role) }}
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="showModal = false">Cancel</button>
            <button class="btn btn-primary" @click="saveUser" :disabled="saving">
              <i v-if="saving" class="ti ti-loader-2 spin"></i>
              <i v-else class="ti ti-check"></i>
              {{ editing ? 'Save Changes' : 'Create User' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Reset Password Modal -->
    <Teleport to="body">
      <div v-if="resetTarget" class="modal-backdrop" @click.self="resetTarget = null">
        <div class="modal" style="width:380px">
          <div class="modal-header">
            <div class="modal-title"><i class="ti ti-key"></i> Reset Password — {{ resetTarget.username }}</div>
            <button class="btn btn-ghost btn-sm" @click="resetTarget = null"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body">
            <div class="field">
              <label>New Password *</label>
              <input v-model="newPassword" type="password" placeholder="Min 8 characters">
            </div>
            <div class="field">
              <label>Confirm Password *</label>
              <input v-model="confirmPassword" type="password" placeholder="Repeat new password">
            </div>
            <div v-if="pwError" class="pw-error">{{ pwError }}</div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="resetTarget = null">Cancel</button>
            <button class="btn btn-primary" @click="doResetPw" :disabled="saving">
              <i class="ti ti-check"></i> Reset Password
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useToast } from 'vue-toastification'
import { usersApi } from '@/api'
import type { User, UserRole } from '@/types'
import dayjs from 'dayjs'

const toast = useToast()
const users = ref<User[]>([])
const loading = ref(false)
const saving  = ref(false)
const showModal   = ref(false)
const editing     = ref<User | null>(null)
const resetTarget = ref<User | null>(null)
const filterRole   = ref('')
const filterStatus = ref('')
const newPassword  = ref('')
const confirmPassword = ref('')
const pwError = ref('')

const form = ref({ username: '', full_name: '', email: '', role: 'DBA' as UserRole, password: '' })

const roles = [
  { name: 'SuperAdmin',    badge: 'badge-red',    perms: ['Full access to all features', 'Manage users & RBAC', 'View audit log'] },
  { name: 'DBA',           badge: 'badge-blue',   perms: ['Manage hosts & instances', 'Full CRUD on rules', 'Backup & restore'] },
  { name: 'AccessManager', badge: 'badge-purple', perms: ['Full CRUD on rules', 'Grant / revoke temp access', 'Dashboard only — no hosts, instances, backups, or users'] },
  { name: 'TempAccessManager', badge: 'badge-green', perms: ['Grant / revoke temp access only', 'No access to Access Rules', 'Dashboard only — no hosts, instances, backups, or users'] },
  { name: 'Operator',      badge: 'badge-amber',  perms: ['Enable / disable rules', 'Grant temp access', 'View rules'] },
  { name: 'Viewer',        badge: 'badge-gray',   perms: ['Read-only access to rules', 'View dashboard'] },
]

function roleBadge(role: string) {
  const m: Record<string, string> = { SuperAdmin: 'badge-red', DBA: 'badge-blue', AccessManager: 'badge-purple', TempAccessManager: 'badge-green', Operator: 'badge-amber', Viewer: 'badge-gray' }
  return m[role] || 'badge-gray'
}
function roleHint(role: string) {
  const m: Record<string, string> = {
    SuperAdmin: 'Full access including user management and all system settings.',
    DBA: 'Can manage hosts, instances, rules, and backups. Cannot manage users.',
    AccessManager: 'Full CRUD on Access Rules and Temporary Access (including write-back to the server). Can view the Dashboard. No access to Hosts, Instances, Backups, Users, or Settings.',
    TempAccessManager: 'Can only grant/revoke Temporary Access and view the Dashboard. Cannot see or edit Access Rules, and no access to Hosts, Instances, Backups, Users, or Settings.',
    Operator: 'Can enable/disable rules and grant temporary access. Cannot edit rules.',
    Viewer: 'Read-only access. Can view rules and dashboard only.',
  }
  return m[role] || ''
}
function initials(u: User) { return u.full_name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2) }
function countByRole(role: string) { return users.value.filter(u => u.role === role).length }
function fmt(d: string) { return dayjs(d).format('DD MMM YYYY HH:mm') }

async function load() {
  loading.value = true
  try {
    const params: any = {}
    if (filterRole.value)   params.role   = filterRole.value
    const { data } = await usersApi.list(params)
    users.value = filterStatus.value
      ? data.filter(u => u.status === filterStatus.value)
      : data
  } catch { toast.error('Failed to load users') }
  finally { loading.value = false }
}

function openCreate() { editing.value = null; form.value = { username: '', full_name: '', email: '', role: 'DBA', password: '' }; showModal.value = true }
function openEdit(user: User) {
  editing.value = user
  form.value = { username: user.username, full_name: user.full_name, email: user.email, role: user.role, password: '' }
  showModal.value = true
}
function openResetPw(user: User) { resetTarget.value = user; newPassword.value = ''; confirmPassword.value = ''; pwError.value = '' }

async function saveUser() {
  if (!form.value.username || !form.value.full_name || !form.value.email) {
    toast.error('All required fields must be filled'); return
  }
  if (!editing.value && form.value.password.length < 8) {
    toast.error('Password must be at least 8 characters'); return
  }
  saving.value = true
  try {
    if (editing.value) {
      await usersApi.update(editing.value.id, { full_name: form.value.full_name, email: form.value.email, role: form.value.role })
      toast.success('User updated')
    } else {
      await usersApi.create(form.value as any)
      toast.success('User created')
    }
    showModal.value = false
    await load()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Failed to save user')
  } finally { saving.value = false }
}

async function toggleStatus(user: User) {
  const newStatus = user.status === 'Active' ? 'Inactive' : 'Active'
  try {
    await usersApi.update(user.id, { status: newStatus as any })
    user.status = newStatus as any
    toast.success(`${user.username} ${newStatus.toLowerCase()}`)
  } catch { toast.error('Failed to update status') }
}

async function deleteUser(user: User) {
  if (user.username === 'admin') { toast.error('Cannot delete the default admin user'); return }
  if (!confirm(`Delete user "${user.username}"?\n\nThis action cannot be undone.`)) return
  try {
    await usersApi.delete(user.id)
    toast.success('User deleted')
    await load()
  } catch { toast.error('Failed to delete user') }
}

async function doResetPw() {
  pwError.value = ''
  if (newPassword.value.length < 8) { pwError.value = 'Password must be at least 8 characters'; return }
  if (newPassword.value !== confirmPassword.value) { pwError.value = 'Passwords do not match'; return }
  try {
    await usersApi.resetPassword(resetTarget.value!.id, newPassword.value)
    toast.success(`Password reset for ${resetTarget.value!.username}`)
    resetTarget.value = null
  } catch { toast.error('Failed to reset password') }
}

onMounted(load)
</script>

<style scoped>
.rbac-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px; margin-bottom: 16px; }
.rbac-card { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: 8px; padding: 12px; }
.rbac-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
.rbac-count { font-size: 11px; color: var(--c-muted); }
.rbac-perms { list-style: none; display: flex; flex-direction: column; gap: 5px; }
.rbac-perms li { font-size: 11px; color: var(--c-muted); display: flex; align-items: flex-start; gap: 5px; }
.user-avatar { width: 26px; height: 26px; border-radius: 6px; background: var(--c-accent); display: flex; align-items: center; justify-content: center; font-size: 10px; font-weight: 600; color: #fff; flex-shrink: 0; }
.role-hint { background: var(--c-bg); border: 1px solid var(--c-border); border-radius: 6px; padding: 8px 12px; font-size: 11px; color: var(--c-muted); display: flex; gap: 8px; align-items: flex-start; margin-top: 4px; }
.pw-error { background: rgba(239,68,68,.1); border: 1px solid rgba(239,68,68,.25); color: var(--c-red); padding: 8px 12px; border-radius: 6px; font-size: 12px; margin-top: 8px; }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
