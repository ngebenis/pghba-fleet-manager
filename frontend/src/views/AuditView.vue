<template>
  <div class="page-content">
    <!-- Filters -->
    <div class="toolbar" style="flex-wrap:wrap;gap:8px">
      <select class="filter-select" v-model="filters.event" @change="resetLoad">
        <option value="">All Events</option>
        <optgroup label="Access Rules">
          <option>AddRule</option>
          <option>EditRule</option>
          <option>DeleteRule</option>
          <option>EnableRule</option>
          <option>DisableRule</option>
        </optgroup>
        <optgroup label="Temporary Access">
          <option>GrantTempAccess</option>
          <option>RevokeTempAccess</option>
          <option>ExpireTempAccess</option>
        </optgroup>
        <optgroup label="Hosts & Instances">
          <option>AddHost</option>
          <option>EditHost</option>
          <option>DeleteHost</option>
          <option>AddInstance</option>
        </optgroup>
        <optgroup label="Auth">
          <option>Login</option>
          <option>Logout</option>
          <option>LoginFailed</option>
        </optgroup>
        <optgroup label="System">
          <option>ReloadConf</option>
          <option>CreateBackup</option>
          <option>RestoreBackup</option>
          <option>AddUser</option>
          <option>EditUser</option>
          <option>DeleteUser</option>
        </optgroup>
      </select>

      <select class="filter-select" v-model="filters.username" @change="resetLoad">
        <option value="">All Users</option>
        <option v-for="u in allUsers" :key="u" :value="u">{{ u }}</option>
      </select>

      <select class="filter-select" v-model="filters.instance_id" @change="resetLoad" style="width:180px">
        <option value="">All Instances</option>
        <option v-for="i in instances" :key="i.id" :value="i.id">
          {{ i.instance_name }}:{{ i.port }}
        </option>
      </select>

      <div style="display:flex;align-items:center;gap:6px">
        <span style="font-size:11px;color:var(--c-muted)">From</span>
        <input class="filter-select" type="date" v-model="filters.from" @change="resetLoad" style="cursor:pointer;width:130px">
      </div>
      <div style="display:flex;align-items:center;gap:6px">
        <span style="font-size:11px;color:var(--c-muted)">To</span>
        <input class="filter-select" type="date" v-model="filters.to" @change="resetLoad" style="cursor:pointer;width:130px">
      </div>

      <div class="toolbar-right">
        <button class="btn btn-ghost" @click="clearFilters">
          <i class="ti ti-filter-off"></i> Clear
        </button>
        <button class="btn btn-ghost" @click="exportCSV" :disabled="exporting">
          <i :class="['ti', exporting ? 'ti-loader-2 spin' : 'ti-download']"></i>
          Export CSV
        </button>
      </div>
    </div>

    <!-- Event stats -->
    <div class="event-stats">
      <div v-for="s in eventStats" :key="s.label"
        :class="['event-stat', filters.event === s.event ? 'active' : '']"
        @click="toggleEventFilter(s.event)"
        style="cursor:pointer">
        <div class="event-stat-count" :style="{ color: s.color }">{{ s.count }}</div>
        <div class="event-stat-label">{{ s.label }}</div>
      </div>
    </div>

    <div class="table-wrap">
      <div class="table-header">
        <span class="table-title">
          Audit Log
          <span class="total-badge">{{ total.toLocaleString() }} events</span>
        </span>
        <div style="display:flex;align-items:center;gap:8px">
          <span style="font-size:12px;color:var(--c-muted)">Per page:</span>
          <select class="filter-select" v-model="pageSize" @change="resetLoad" style="width:70px">
            <option :value="25">25</option>
            <option :value="50">50</option>
            <option :value="100">100</option>
          </select>
        </div>
      </div>

      <table>
        <thead>
          <tr>
            <th style="width:145px">Time</th>
            <th style="width:100px">User</th>
            <th style="width:130px">Event</th>
            <th>Target</th>
            <th style="width:180px">Instance</th>
            <th style="width:110px">Client IP</th>
            <th style="width:60px">Detail</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="7" style="text-align:center;padding:32px;color:var(--c-muted)">
              <i class="ti ti-loader-2 spin" style="font-size:20px"></i>
            </td>
          </tr>
          <tr v-for="log in logs" :key="log.id" class="log-row">
            <td style="font-family:monospace;font-size:11px;color:var(--c-muted);white-space:nowrap">
              {{ fmt(log.created_at) }}
            </td>
            <td>
              <div style="display:flex;align-items:center;gap:6px">
                <div class="log-avatar">{{ (log.username||'?')[0].toUpperCase() }}</div>
                <span style="font-size:12px">{{ log.username }}</span>
              </div>
            </td>
            <td>
              <span :class="['event-badge', eventBadge(log.event)]">
                <i :class="['ti', eventIcon(log.event)]"></i>
                {{ log.event }}
              </span>
            </td>
            <td class="mono truncate" style="max-width:180px;font-size:11px" :title="log.target">
              {{ log.target || '—' }}
            </td>
            <td style="font-size:11px;color:var(--c-muted)">{{ log.instance_label || '—' }}</td>
            <td style="font-family:monospace;font-size:11px;color:var(--c-muted)">{{ log.client_ip || '—' }}</td>
            <td>
              <button v-if="log.detail && log.detail !== 'null' && log.detail !== '{}'"
                class="btn btn-ghost btn-sm" @click="showDetail(log)" title="View detail">
                <i class="ti ti-braces"></i>
              </button>
            </td>
          </tr>
          <tr v-if="!loading && logs.length === 0">
            <td colspan="7" style="text-align:center;padding:40px;color:var(--c-muted)">
              <i class="ti ti-search-off" style="font-size:24px;display:block;margin-bottom:8px"></i>
              No events found
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <div class="pagination" v-if="totalPages > 1">
        <button class="btn btn-ghost btn-sm" @click="goPage(0)" :disabled="page === 0">
          <i class="ti ti-chevrons-left"></i>
        </button>
        <button class="btn btn-ghost btn-sm" @click="goPage(page - 1)" :disabled="page === 0">
          <i class="ti ti-chevron-left"></i>
        </button>
        <template v-for="p in pageNumbers" :key="p">
          <span v-if="p === '...'" class="page-ellipsis">...</span>
          <button v-else
            :class="['btn','btn-sm', p === page + 1 ? 'btn-primary' : 'btn-ghost']"
            @click="goPage((p as number) - 1)">{{ p }}</button>
        </template>
        <button class="btn btn-ghost btn-sm" @click="goPage(page + 1)" :disabled="page >= totalPages - 1">
          <i class="ti ti-chevron-right"></i>
        </button>
        <button class="btn btn-ghost btn-sm" @click="goPage(totalPages - 1)" :disabled="page >= totalPages - 1">
          <i class="ti ti-chevrons-right"></i>
        </button>
        <span class="page-info">
          {{ page * pageSize + 1 }}–{{ Math.min((page + 1) * pageSize, total) }} of {{ total.toLocaleString() }}
        </span>
      </div>
    </div>

    <!-- Detail Modal -->
    <Teleport to="body">
      <div v-if="detailLog" class="modal-backdrop" @click.self="detailLog = null">
        <div class="modal" style="width:540px">
          <div class="modal-header">
            <div class="modal-title" style="display:flex;align-items:center;gap:8px">
              <span :class="['event-badge', eventBadge(detailLog.event)]">
                <i :class="['ti', eventIcon(detailLog.event)]"></i>
                {{ detailLog.event }}
              </span>
              <span style="font-size:12px;color:var(--c-muted)">{{ fmt(detailLog.created_at) }}</span>
            </div>
            <button class="btn btn-ghost btn-sm" @click="detailLog = null"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body">
            <div class="detail-grid">
              <div class="detail-row">
                <span class="dl">User</span>
                <span class="dv">
                  <div class="log-avatar" style="display:inline-flex;margin-right:6px">
                    {{ (detailLog.username||'?')[0].toUpperCase() }}
                  </div>
                  {{ detailLog.username }}
                </span>
              </div>
              <div class="detail-row">
                <span class="dl">Target</span>
                <span class="dv mono">{{ detailLog.target || '—' }}</span>
              </div>
              <div class="detail-row">
                <span class="dl">Instance</span>
                <span class="dv">{{ detailLog.instance_label || '—' }}</span>
              </div>
              <div class="detail-row">
                <span class="dl">Client IP</span>
                <span class="dv mono">{{ detailLog.client_ip || '—' }}</span>
              </div>
              <div class="detail-row">
                <span class="dl">Event ID</span>
                <span class="dv mono" style="font-size:10px;color:var(--c-muted)">{{ detailLog.id }}</span>
              </div>
            </div>
            <div style="margin-top:14px">
              <div style="font-size:11px;color:var(--c-muted);margin-bottom:6px;text-transform:uppercase;letter-spacing:.5px">Payload</div>
              <pre class="json-view">{{ prettyJSON(detailLog.detail) }}</pre>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>

    <!-- Diff Modal -->
    <Teleport to="body">
      <div v-if="diffData" class="modal-backdrop" @click.self="diffData = null">
        <div class="modal" style="width:780px;max-height:85vh;display:flex;flex-direction:column">
          <div class="modal-header">
            <div class="modal-title">
              <i class="ti ti-git-diff" style="color:var(--c-teal)"></i>
              pg_hba.conf Diff — {{ diffData.instance_label }}
            </div>
            <button class="btn btn-ghost btn-sm" @click="diffData = null"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body" style="flex:1;overflow:auto;padding:0">
            <div v-if="loadingDiff" style="text-align:center;padding:32px;color:var(--c-muted)">
              <i class="ti ti-loader-2 spin"></i> Loading diff...
            </div>
            <div v-else>
              <!-- Stats -->
              <div style="display:flex;gap:8px;padding:12px 16px;border-bottom:1px solid var(--c-border)">
                <span style="display:inline-flex;align-items:center;gap:4px;background:rgba(34,197,94,.1);color:var(--c-green);border:1px solid rgba(34,197,94,.2);padding:3px 10px;border-radius:5px;font-size:11px">
                  <i class="ti ti-plus"></i> {{ diffData.added }} lines added
                </span>
                <span style="display:inline-flex;align-items:center;gap:4px;background:rgba(239,68,68,.1);color:var(--c-red);border:1px solid rgba(239,68,68,.2);padding:3px 10px;border-radius:5px;font-size:11px">
                  <i class="ti ti-minus"></i> {{ diffData.removed }} lines removed
                </span>
                <span style="font-size:11px;color:var(--c-muted);margin-left:auto;align-self:center">
                  {{ diffData.created_at ? fmt(diffData.created_at) : '' }}
                </span>
              </div>

              <!-- Diff view tabs -->
              <div style="display:flex;border-bottom:1px solid var(--c-border)">
                <button :class="['diff-tab', diffTab === 'unified' ? 'active' : '']" @click="diffTab = 'unified'">
                  Unified Diff
                </button>
                <button :class="['diff-tab', diffTab === 'split' ? 'active' : '']" @click="diffTab = 'split'">
                  Split View
                </button>
                <button :class="['diff-tab', diffTab === 'after' ? 'active' : '']" @click="diffTab = 'after'">
                  After
                </button>
                <button :class="['diff-tab', diffTab === 'before' ? 'active' : '']" @click="diffTab = 'before'">
                  Before
                </button>
              </div>

              <!-- Unified diff -->
              <div v-if="diffTab === 'unified'" class="diff-container">
                <div v-for="(line, i) in diffData.diff_lines" :key="i"
                  :class="['diff-line', line.type === 'added' ? 'diff-added' : line.type === 'removed' ? 'diff-removed' : 'diff-context']">
                  <span class="diff-symbol">{{ line.type === 'added' ? '+' : line.type === 'removed' ? '-' : ' ' }}</span>
                  <span class="diff-content">{{ line.content }}</span>
                </div>
                <div v-if="!diffData.diff_lines?.length" style="padding:16px;color:var(--c-muted);text-align:center">
                  No changes detected
                </div>
              </div>

              <!-- Split view -->
              <div v-if="diffTab === 'split'" style="display:grid;grid-template-columns:1fr 1fr;gap:0;height:400px">
                <div style="border-right:1px solid var(--c-border);overflow:auto">
                  <div style="padding:6px 10px;background:var(--c-bg);font-size:10px;color:var(--c-muted);border-bottom:1px solid var(--c-border);position:sticky;top:0">
                    BEFORE
                  </div>
                  <pre class="diff-pre">{{ diffData.before_content }}</pre>
                </div>
                <div style="overflow:auto">
                  <div style="padding:6px 10px;background:var(--c-bg);font-size:10px;color:var(--c-green);border-bottom:1px solid var(--c-border);position:sticky;top:0">
                    AFTER
                  </div>
                  <pre class="diff-pre">{{ diffData.after_content }}</pre>
                </div>
              </div>

              <!-- After only -->
              <div v-if="diffTab === 'after'" class="diff-container">
                <pre class="diff-pre" style="padding:12px">{{ diffData.after_content }}</pre>
              </div>

              <!-- Before only -->
              <div v-if="diffTab === 'before'" class="diff-container">
                <pre class="diff-pre" style="padding:12px">{{ diffData.before_content }}</pre>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="diffData = null">Close</button>
            <a :href="'data:text/plain;charset=utf-8,' + encodeURIComponent(diffData.after_content || '')"
              :download="'pg_hba.conf.' + (diffData.instance_label || 'unknown') + '.txt'"
              class="btn btn-ghost">
              <i class="ti ti-download"></i> Download After
            </a>
          </div>
        </div>
      </div>
    </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useToast } from 'vue-toastification'
import { auditApi, instancesApi } from '@/api'
import api from '@/api'
import type { AuditLog, PGInstance } from '@/types'
import dayjs from 'dayjs'

const toast = useToast()
const logs      = ref<AuditLog[]>([])
const total     = ref(0)
const loading   = ref(false)
const exporting = ref(false)
const page      = ref(0)
const pageSize  = ref(50)
const detailLog = ref<AuditLog | null>(null)
const instances = ref<PGInstance[]>([])
const allUsers  = ref<string[]>([])

const filters = ref({
  event: '', username: '', instance_id: '', from: '', to: ''
})

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

const pageNumbers = computed(() => {
  const total = totalPages.value
  const cur   = page.value + 1
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages: (number | string)[] = [1]
  if (cur > 3) pages.push('...')
  for (let i = Math.max(2, cur - 1); i <= Math.min(total - 1, cur + 1); i++) pages.push(i)
  if (cur < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

const eventStats = computed(() => [
  { label: 'Rule Changes',   event: '',                color: 'var(--c-green)',  count: logs.value.filter(l => ['AddRule','EditRule','DeleteRule'].includes(l.event)).length },
  { label: 'Enable/Disable', event: 'EnableRule',      color: 'var(--c-amber)',  count: logs.value.filter(l => ['EnableRule','DisableRule'].includes(l.event)).length },
  { label: 'Temp Access',    event: 'GrantTempAccess', color: 'var(--c-accent)', count: logs.value.filter(l => (l.event||'').includes('TempAccess')).length },
  { label: 'Logins',         event: 'Login',           color: 'var(--c-teal)',   count: logs.value.filter(l => l.event === 'Login').length },
  { label: 'Failed Logins',  event: 'LoginFailed',     color: 'var(--c-red)',    count: logs.value.filter(l => l.event === 'LoginFailed').length },
  { label: 'Backups',        event: 'CreateBackup',    color: '#8b5cf6',         count: logs.value.filter(l => (l.event||'').includes('Backup')).length },
])

function toggleEventFilter(event: string) {
  filters.value.event = filters.value.event === event ? '' : event
  resetLoad()
}

function eventBadge(event: string) {
  if (['AddRule','AddHost','AddUser','Login','EnableRule'].includes(event)) return 'ev-green'
  if (['DeleteRule','DeleteHost','DeleteUser','LoginFailed'].includes(event)) return 'ev-red'
  if (['EditRule','EditHost','EditUser','DisableRule'].includes(event)) return 'ev-amber'
  if (['CreateBackup','RestoreBackup'].includes(event)) return 'ev-purple'
  if ((event||'').includes('TempAccess')) return 'ev-blue'
  if (['ReloadConf'].includes(event)) return 'ev-teal'
  return 'ev-gray'
}

function eventIcon(event: string) {
  const map: Record<string, string> = {
    Login: 'ti-login', Logout: 'ti-logout', LoginFailed: 'ti-lock-x',
    AddRule: 'ti-plus', EditRule: 'ti-edit', DeleteRule: 'ti-trash', EnableRule: 'ti-toggle-right', DisableRule: 'ti-toggle-left',
    AddHost: 'ti-server', EditHost: 'ti-server-cog', DeleteHost: 'ti-server-off',
    GrantTempAccess: 'ti-clock-plus', RevokeTempAccess: 'ti-clock-x', ExpireTempAccess: 'ti-clock-minus',
    CreateBackup: 'ti-device-floppy', RestoreBackup: 'ti-restore',
    ReloadConf: 'ti-refresh',
    AddUser: 'ti-user-plus', EditUser: 'ti-user-edit', DeleteUser: 'ti-user-minus',
  }
  return map[event] || 'ti-activity'
}

function parsedDetail(detail: string) {
  try { return JSON.parse(detail) } catch { return null }
}

async function showDiff(log: any) {
  const detail = parsedDetail(log.detail)
  if (!detail?.diff_id) return
  loadingDiff.value = true
  diffData.value = { instance_label: log.instance_label, created_at: log.created_at, diff_lines: [], added: 0, removed: 0 }
  diffTab.value = 'unified'
  try {
    const { data } = await api.get(`/audit/diff/${detail.diff_id}`)
    diffData.value = { ...data, instance_label: log.instance_label }
  } catch {
    toast.error('Failed to load diff')
    diffData.value = null
  } finally {
    loadingDiff.value = false
  }
}

function fmt(d: string) { return dayjs(d).format('DD MMM YYYY HH:mm:ss') }

function prettyJSON(s: string) {
  try { return JSON.stringify(JSON.parse(s), null, 2) }
  catch { return s || '{}' }
}

function showDetail(log: AuditLog) { detailLog.value = log }

async function load() {
  loading.value = true
  try {
    const params: any = { limit: pageSize.value, page: page.value }
    if (filters.value.event)       params.event       = filters.value.event
    if (filters.value.username)    params.username    = filters.value.username
    if (filters.value.instance_id) params.instance_id = filters.value.instance_id
    if (filters.value.from)        params.from        = filters.value.from
    if (filters.value.to)          params.to          = filters.value.to + 'T23:59:59'
    const { data } = await auditApi.list(params)
    logs.value  = data.data || []
    total.value = data.total || 0
    // Collect unique users
    const users = new Set([...allUsers.value, ...logs.value.map(l => l.username).filter(Boolean)])
    allUsers.value = [...users].sort()
  } catch { toast.error('Failed to load audit log') }
  finally { loading.value = false }
}

function resetLoad() { page.value = 0; load() }
function goPage(p: number) { page.value = p; load() }
function clearFilters() {
  filters.value = { event: '', username: '', instance_id: '', from: '', to: '' }
  resetLoad()
}

async function exportCSV() {
  exporting.value = true
  try {
    // Export all matching records
    const params: any = { limit: 10000, page: 0 }
    if (filters.value.event)       params.event       = filters.value.event
    if (filters.value.username)    params.username    = filters.value.username
    if (filters.value.instance_id) params.instance_id = filters.value.instance_id
    if (filters.value.from)        params.from        = filters.value.from
    if (filters.value.to)          params.to          = filters.value.to + 'T23:59:59'

    const { data } = await auditApi.list(params)
    const allLogs: AuditLog[] = data.data || []

    const header = 'ID,Time,User,Event,Target,Instance,ClientIP,Detail\n'
    const rows = allLogs.map(l => [
      l.id,
      fmt(l.created_at),
      l.username,
      l.event,
      '"' + (l.target || '').replace(/"/g, '""') + '"',
      l.instance_label || '',
      l.client_ip || '',
      '"' + (l.detail || '').replace(/"/g, '""') + '"',
    ].join(',')).join('\n')

    const blob = new Blob(['\uFEFF' + header + rows], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `audit-log-${dayjs().format('YYYYMMDD-HHmm')}.csv`
    a.click()
    URL.revokeObjectURL(url)
    toast.success(`Exported ${allLogs.length} events`)
  } catch { toast.error('Export failed') }
  finally { exporting.value = false }
}

onMounted(async () => {
  const { data } = await instancesApi.list()
  instances.value = data
  await load()
})
</script>

<style scoped>
.event-stats { display:flex;gap:8px;margin-bottom:16px;flex-wrap:wrap; }
.event-stat { background:var(--c-surface);border:1px solid var(--c-border);border-radius:8px;padding:10px 16px;text-align:center;min-width:90px;transition:all .15s; }
.event-stat:hover { border-color:var(--c-accent);background:var(--c-surface2); }
.event-stat.active { border-color:var(--c-accent);background:rgba(79,142,247,.08); }
.event-stat-count { font-size:22px;font-weight:700;line-height:1; }
.event-stat-label { font-size:10px;color:var(--c-muted);margin-top:3px;text-transform:uppercase;letter-spacing:.3px; }

.total-badge { background:var(--c-bg);border:1px solid var(--c-border);border-radius:4px;padding:2px 8px;font-size:11px;color:var(--c-muted);margin-left:8px;font-weight:400; }

.log-row:hover { background:var(--c-surface2); }
.mono { font-family:monospace; }
.truncate { overflow:hidden;text-overflow:ellipsis;white-space:nowrap; }

.event-badge { display:inline-flex;align-items:center;gap:4px;padding:3px 8px;border-radius:5px;font-size:11px;font-weight:500;white-space:nowrap; }
.ev-green  { background:rgba(34,197,94,.1);color:var(--c-green);border:1px solid rgba(34,197,94,.15); }
.ev-red    { background:rgba(239,68,68,.1);color:var(--c-red);border:1px solid rgba(239,68,68,.15); }
.ev-amber  { background:rgba(245,158,11,.1);color:var(--c-amber);border:1px solid rgba(245,158,11,.15); }
.ev-blue   { background:rgba(79,142,247,.1);color:var(--c-accent);border:1px solid rgba(79,142,247,.15); }
.ev-purple { background:rgba(139,92,246,.1);color:#8b5cf6;border:1px solid rgba(139,92,246,.15); }
.ev-teal   { background:rgba(20,184,166,.1);color:var(--c-teal);border:1px solid rgba(20,184,166,.15); }
.ev-gray   { background:var(--c-surface2);color:var(--c-muted);border:1px solid var(--c-border); }

.log-avatar { width:20px;height:20px;border-radius:4px;background:var(--c-surface2);border:1px solid var(--c-border);display:inline-flex;align-items:center;justify-content:center;font-size:10px;font-weight:600;flex-shrink:0; }

.detail-grid { display:flex;flex-direction:column;gap:0; }
.detail-row { display:flex;gap:16px;padding:8px 0;border-bottom:1px solid var(--c-border); }
.detail-row:last-of-type { border-bottom:none; }
.dl { font-size:11px;color:var(--c-muted);width:70px;flex-shrink:0;padding-top:1px; }
.dv { font-size:12px;display:flex;align-items:center; }

.json-view { background:var(--c-bg);border:1px solid var(--c-border);border-radius:6px;padding:12px;font-family:monospace;font-size:12px;color:var(--c-accent);overflow-x:auto;white-space:pre;max-height:280px;overflow-y:auto; }

.pagination { display:flex;align-items:center;gap:4px;padding:12px 16px;border-top:1px solid var(--c-border);flex-wrap:wrap; }
.page-info { font-size:12px;color:var(--c-muted);margin-left:auto; }
.page-ellipsis { font-size:12px;color:var(--c-muted);padding:0 4px; }

.spin { animation:spin 1s linear infinite; }
@keyframes spin { to { transform:rotate(360deg); } }

.diff-tab { padding:8px 16px;font-size:12px;background:none;border:none;color:var(--c-muted);cursor:pointer;border-bottom:2px solid transparent; }
.diff-tab.active { color:var(--c-text);border-bottom-color:var(--c-accent); }
.diff-container { max-height:400px;overflow:auto; }
.diff-line { display:flex;align-items:flex-start;gap:8px;padding:2px 12px;font-family:monospace;font-size:11px;min-height:20px; }
.diff-added   { background:rgba(34,197,94,.08);color:var(--c-green); }
.diff-removed { background:rgba(239,68,68,.08);color:var(--c-red); }
.diff-context { color:var(--c-muted); }
.diff-symbol  { width:12px;flex-shrink:0;font-weight:700; }
.diff-content { flex:1;word-break:break-all; }
.diff-pre     { font-family:monospace;font-size:11px;color:var(--c-text);white-space:pre;margin:0; }
</style>
