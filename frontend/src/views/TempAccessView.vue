<template>
  <div class="page-content">
    <div class="toolbar">
      <span style="font-size:12px;color:var(--c-muted)">
        <i class="ti ti-clock" style="font-size:14px"></i>
        Background scheduler checks every 60s and auto-disables expired rules
      </span>
      <div class="toolbar-right">
        <button class="btn btn-primary" @click="showModal = true">
          <i class="ti ti-clock-plus"></i> Grant Temp Access
        </button>
      </div>
    </div>

    <div class="table-wrap">
      <div class="table-header">
        <span class="table-title">
          Temporary Access
          <span class="total-count">{{ filtered.length }} grants</span>
        </span>
        <div style="display:flex;align-items:center;gap:8px">
          <select class="filter-select" v-model="filterEnv" @change="currentPage = 1" style="width:130px">
            <option value="">All Environments</option>
            <option>Production</option>
            <option>Staging</option>
            <option>UAT</option>
            <option>Development</option>
          </select>
          <select class="filter-select" v-model="filterStatus" @change="currentPage = 1" style="width:120px">
            <option value="">All Status</option>
            <option value="active">Active</option>
            <option value="expired">Expired</option>
          </select>
          <span style="font-size:12px;color:var(--c-muted)">Per page:</span>
          <select class="filter-select" v-model="perPage" @change="currentPage = 1" style="width:70px">
            <option :value="10">10</option>
            <option :value="20">20</option>
            <option :value="30">30</option>
            <option :value="50">50</option>
            <option :value="100">100</option>
          </select>
        </div>
      </div>

      <table>
        <thead>
          <tr>
            <th>IP Address</th>
            <th>Database</th>
            <th>User</th>
            <th>Method</th>
            <th>Instance</th>
            <th>Host</th>
            <th>Granted By</th>
            <th>Expires At</th>
            <th>Remaining</th>
            <th>Status</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="11" style="text-align:center;padding:24px;color:var(--c-muted)">
              <i class="ti ti-loader-2 spin"></i> Loading...
            </td>
          </tr>
          <tr v-for="ta in paginated" :key="ta.id" :class="{ 'row-expired': isExpired(ta) }">
            <td class="mono">{{ ta.ip_address || '—' }}</td>
            <td>{{ ta.database || 'all' }}</td>
            <td>{{ ta.db_user || 'all' }}</td>
            <td class="method">{{ ta.auth_method || 'md5' }}</td>
            <td style="font-size:11px">
              {{ instanceName(ta.instance_id) }}
            </td>
            <td style="font-size:11px;color:var(--c-muted)">
              {{ hostName(ta.instance_id) }}
            </td>
            <td style="font-size:12px">
              {{ ta.granted_by_username || ta.granted_by?.substring(0,8) }}
            </td>
            <td style="font-size:11px;color:var(--c-muted);white-space:nowrap">
              {{ formatDate(ta.expires_at) }}
            </td>
            <td>
              <div v-if="!isExpired(ta)" class="remaining">
                <span class="remaining-label">{{ remaining(ta.expires_at) }}</span>
                <div class="prog-track">
                  <div class="prog-fill" :style="{ width: progressPct(ta) + '%' }"></div>
                </div>
              </div>
              <span v-else style="font-size:11px;color:var(--c-muted)">—</span>
            </td>
            <td>
              <span :class="['badge', isExpired(ta) || ta.revoked ? 'badge-gray' : 'badge-green']">
                {{ ta.revoked ? 'Revoked' : isExpired(ta) ? 'Expired' : 'Active' }}
              </span>
            </td>
            <td>
              <button
                v-if="!ta.revoked && !isExpired(ta)"
                class="btn btn-danger btn-sm"
                @click="revoke(ta.id)"
              >
                <i class="ti ti-ban"></i> Revoke
              </button>
            </td>
          </tr>
          <tr v-if="!loading && filtered.length === 0">
            <td colspan="11" style="text-align:center;padding:32px;color:var(--c-muted)">
              No temporary access grants
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <div class="pagination" v-if="totalPages > 1">
        <button class="btn btn-ghost btn-sm" @click="currentPage = 1" :disabled="currentPage === 1">
          <i class="ti ti-chevrons-left"></i>
        </button>
        <button class="btn btn-ghost btn-sm" @click="currentPage--" :disabled="currentPage === 1">
          <i class="ti ti-chevron-left"></i>
        </button>
        <template v-for="p in pageNumbers" :key="p">
          <span v-if="p === '...'" class="page-ellipsis">...</span>
          <button v-else :class="['btn','btn-sm', p === currentPage ? 'btn-primary' : 'btn-ghost']"
            @click="currentPage = p as number">{{ p }}</button>
        </template>
        <button class="btn btn-ghost btn-sm" @click="currentPage++" :disabled="currentPage === totalPages">
          <i class="ti ti-chevron-right"></i>
        </button>
        <button class="btn btn-ghost btn-sm" @click="currentPage = totalPages" :disabled="currentPage === totalPages">
          <i class="ti ti-chevrons-right"></i>
        </button>
        <span class="page-info">
          {{ (currentPage-1)*perPage+1 }}–{{ Math.min(currentPage*perPage, filtered.length) }} of {{ filtered.length }}
        </span>
      </div>
    </div>

    <!-- Grant Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
        <div class="modal">
          <div class="modal-header">
            <div class="modal-title"><i class="ti ti-clock-plus"></i> Grant Temporary Access</div>
            <button class="btn btn-ghost btn-sm" @click="showModal = false"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body">
            <div class="field">
              <label>Instance</label>
              <select v-model="form.instance_id">
                <option v-for="i in instances" :key="i.id" :value="i.id">
                  {{ i.instance_name }} @ port {{ i.port }} (PG {{ i.pg_version }}){{ i.host?.ip_address ? ` — ${i.host.ip_address}` : '' }}
                </option>
              </select>
              <div v-if="selectedInstance?.host" class="text-muted" style="font-size:12px;margin-top:4px">
                <i class="ti ti-server"></i> {{ selectedInstance.host.name }} &middot; {{ selectedInstance.host.ip_address }}
              </div>
            </div>
            <div class="field">
              <label>IP Address</label>
              <input v-model="form.ip_address" placeholder="103.25.100.50 or 10.0.0.0/24">
            </div>
            <div class="field-row">
              <div class="field"><label>Database</label><input v-model="form.database" placeholder="all"></div>
              <div class="field"><label>User</label><input v-model="form.db_user" placeholder="all"></div>
            </div>
            <div class="field">
              <label>Auth Method</label>
              <select v-model="form.auth_method">
                <option>md5</option>
                <option>scram-sha-256</option>
                <option>trust</option>
              </select>
            </div>
            <div class="field-row">
              <div class="field">
                <label>Duration</label>
                <input v-model.number="form.duration_value" type="number" min="1" placeholder="2">
              </div>
              <div class="field">
                <label>Unit</label>
                <select v-model="form.duration_unit">
                  <option>Minutes</option>
                  <option>Hours</option>
                  <option>Days</option>
                  <option>Weeks</option>
                  <option>Months</option>
                </select>
              </div>
            </div>
            <div class="field">
              <label>Reason / Note</label>
              <input v-model="form.reason" placeholder="e.g. Dev access for hotfix">
            </div>
            <div class="preview-box">
              <div style="font-size:10px;color:var(--c-muted);margin-bottom:6px">WILL GENERATE:</div>
              <div class="hba-line">
                host &nbsp;{{ form.database || 'all' }} &nbsp;{{ form.db_user || 'all' }}
                &nbsp;{{ form.ip_address || '0.0.0.0/32' }} &nbsp;{{ form.auth_method }}
              </div>
              <div style="font-size:10px;color:var(--c-muted);margin-top:4px">
                Expires: {{ previewExpiry }}
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="showModal = false">Cancel</button>
            <button class="btn btn-success" @click="grant" :disabled="saving">
              <i :class="['ti', saving ? 'ti-loader-2 spin' : 'ti-clock']"></i>
              Grant Access
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useToast } from 'vue-toastification'
import { tempAccessApi, instancesApi } from '@/api'
import { globalSearch } from '@/stores/search'
import type { PGInstance } from '@/types'
import dayjs from 'dayjs'

const toast = useToast()

// Use any for now since structure changed
const tempAccess = ref<any[]>([])
const instances  = ref<PGInstance[]>([])
const loading    = ref(false)
const showModal  = ref(false)
const saving     = ref(false)
const filterStatus = ref('')
const filterEnv    = ref('')
let timer: ReturnType<typeof setInterval>

// Pagination
const perPage     = ref(20)
const currentPage = ref(1)

const filtered = computed(() => {
  const q = globalSearch.value.toLowerCase().trim()
  return tempAccess.value.filter(ta => {
    if (filterEnv.value) {
      const inst = instances.value.find(i => i.id === ta.instance_id) as any
      if (inst?.host?.environment !== filterEnv.value) return false
    }
    if (filterStatus.value) {
      const exp = isExpired(ta) || ta.revoked
      if (filterStatus.value === 'expired' && !exp) return false
      if (filterStatus.value === 'active' && exp) return false
    }
    if (q) {
      const inst = instanceName(ta.instance_id).toLowerCase()
      const host = hostName(ta.instance_id).toLowerCase()
      const haystack = [
        ta.ip_address, ta.database, ta.db_user,
        ta.auth_method, ta.granted_by_username, ta.reason, inst, host
      ].join(' ').toLowerCase()
      if (!haystack.includes(q)) return false
    }
    return true
  })
})

const totalPages = computed(() => Math.max(1, Math.ceil(filtered.value.length / perPage.value)))

const paginated = computed(() => {
  const start = (currentPage.value - 1) * perPage.value
  return filtered.value.slice(start, start + perPage.value)
})

const pageNumbers = computed(() => {
  const total = totalPages.value
  const cur   = currentPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages: (number | string)[] = [1]
  if (cur > 3) pages.push('...')
  for (let i = Math.max(2, cur - 1); i <= Math.min(total - 1, cur + 1); i++) pages.push(i)
  if (cur < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

const form = ref({
  instance_id: '',
  ip_address: '',
  database: 'all',
  db_user: 'all',
  auth_method: 'md5',
  duration_value: 2,
  duration_unit: 'Hours',
  reason: '',
})

// Instance selected in the Grant Temporary Access modal
const selectedInstance = computed(() => instances.value.find(i => i.id === form.value.instance_id))

const previewExpiry = computed(() => {
  const now = dayjs()
  const val = form.value.duration_value || 1
  const unit = form.value.duration_unit
  let exp = now
  if (unit === 'Minutes') exp = now.add(val, 'minute')
  else if (unit === 'Hours') exp = now.add(val, 'hour')
  else if (unit === 'Days') exp = now.add(val, 'day')
  else if (unit === 'Weeks') exp = now.add(val * 7, 'day')
  else if (unit === 'Months') exp = now.add(val, 'month')
  return exp.format('DD MMM YYYY HH:mm')
})

function instanceName(instanceId: string): string {
  const inst = instances.value.find(i => i.id === instanceId)
  if (!inst) return instanceId?.substring(0, 8) || '—'
  return `${inst.instance_name}:${inst.port} (PG ${inst.pg_version})`
}

function hostName(instanceId: string): string {
  const inst = instances.value.find(i => i.id === instanceId) as any
  if (!inst?.host) return '—'
  return `${inst.host.name} (${inst.host.ip_address})`
}

function isExpired(ta: any) {
  return ta.revoked || new Date(ta.expires_at) <= new Date()
}

function remaining(expiresAt: string): string {
  const diff = new Date(expiresAt).getTime() - Date.now()
  if (diff <= 0) return 'Expired'
  const h = Math.floor(diff / 3600000)
  const m = Math.floor((diff % 3600000) / 60000)
  return h > 0 ? `${h}h ${m}m` : `${m}m`
}

function progressPct(ta: any): number {
  const start = new Date(ta.created_at).getTime()
  const end   = new Date(ta.expires_at).getTime()
  const now   = Date.now()
  return Math.max(0, Math.min(100, ((now - start) / (end - start)) * 100))
}

function formatDate(d: string) {
  return dayjs(d).format('DD MMM HH:mm')
}

async function load() {
  loading.value = true
  try {
    const { data } = await tempAccessApi.list()
    tempAccess.value = Array.isArray(data) ? data : []
  } catch (e) {
    tempAccess.value = []
  } finally {
    loading.value = false
  }
}

async function grant() {
  if (!form.value.instance_id || !form.value.ip_address) {
    toast.error('Instance and IP are required')
    return
  }
  saving.value = true
  try {
    await tempAccessApi.grant(form.value as any)
    toast.success(`Temp access granted to ${form.value.ip_address} for ${form.value.duration_value} ${form.value.duration_unit}`)
    showModal.value = false
    form.value.ip_address = ''
    form.value.reason = ''
    await load()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Failed to grant access')
  } finally {
    saving.value = false
  }
}

async function revoke(id: string) {
  if (!confirm('Revoke this temporary access?')) return
  try {
    await tempAccessApi.revoke(id)
    toast.success('Temp access revoked')
    await load()
  } catch {
    toast.error('Failed to revoke')
  }
}

onMounted(async () => {
  const { data } = await instancesApi.list()
  instances.value = data
  if (data.length) form.value.instance_id = data[0].id
  await load()
  timer = setInterval(load, 30000)
})
onUnmounted(() => clearInterval(timer))
</script>

<style scoped>
.remaining { display:flex;flex-direction:column;gap:3px; }
.remaining-label { font-size:11px;font-family:monospace;color:var(--c-amber); }
.prog-track { height:3px;background:var(--c-border);border-radius:2px;overflow:hidden;width:80px; }
.prog-fill  { height:100%;background:var(--c-amber);border-radius:2px;transition:width 1s; }
.mono       { font-family:monospace;font-size:12px; }
.method     { font-family:monospace;font-size:11px;color:var(--c-teal); }
.preview-box { background:var(--c-bg);border:1px solid var(--c-border);border-radius:6px;padding:10px 12px;margin-top:4px; }
.hba-line { font-family:monospace;font-size:12px;color:var(--c-green); }
.total-count { font-size:12px;color:var(--c-muted);font-weight:400;margin-left:8px; }
.row-expired td { opacity:.5; }
.pagination { display:flex;align-items:center;gap:4px;padding:12px 16px;border-top:1px solid var(--c-border);flex-wrap:wrap; }
.page-info { font-size:12px;color:var(--c-muted);margin-left:auto; }
.page-ellipsis { font-size:12px;color:var(--c-muted);padding:0 4px; }
.btn-success { background:rgba(34,197,94,.1);border-color:rgba(34,197,94,.3);color:var(--c-green); }
.btn-success:hover { background:rgba(34,197,94,.2); }
.spin { animation:spin 1s linear infinite; }
@keyframes spin { to { transform:rotate(360deg); } }
</style>
