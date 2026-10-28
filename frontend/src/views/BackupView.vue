<template>
  <div class="page-content">
    <div class="top-row">
      <!-- Create Backup Card -->
      <div class="card" style="flex:1">
        <div class="card-title-row">
          <span class="card-title-text">
            <i class="ti ti-device-floppy" style="color:var(--c-green)"></i> Create Backup
          </span>
        </div>
        <div class="field">
          <label>Instance</label>
          <select v-model="createForm.instance_id">
            <option value="" disabled>Choose instance...</option>
            <option v-for="i in instances" :key="i.id" :value="i.id">
              {{ i.instance_name }}:{{ i.port }}
              <template v-if="(i as any).host"> @ {{ (i as any).host.name }}</template>
            </option>
          </select>
        </div>
        <div class="field">
          <label>Note (optional)</label>
          <input v-model="createForm.note" placeholder="e.g. before adding new IP whitelist">
        </div>
        <button class="btn btn-primary" @click="createBackup" :disabled="creating"
          style="width:100%;justify-content:center;margin-top:4px">
          <i :class="['ti', creating ? 'ti-loader-2 spin' : 'ti-device-floppy']"></i>
          {{ creating ? 'Creating...' : 'Create Backup Now' }}
        </button>
      </div>

      <!-- Auto-backup Policy -->
      <div class="card" style="flex:1">
        <div class="card-title-row">
          <span class="card-title-text">
            <i class="ti ti-shield-check" style="color:var(--c-accent)"></i> Auto-Backup Policy
          </span>
        </div>
        <div class="policy-list">
          <div class="policy-item">
            <span class="policy-label"><i class="ti ti-edit"></i> Before edit rule</span>
            <span class="badge badge-green">Active</span>
          </div>
          <div class="policy-item">
            <span class="policy-label"><i class="ti ti-trash"></i> Before delete rule</span>
            <span class="badge badge-green">Active</span>
          </div>
          <div class="policy-item">
            <span class="policy-label"><i class="ti ti-history"></i> Before restore</span>
            <span class="badge badge-green">Active</span>
          </div>
          <div class="policy-item">
            <span class="policy-label"><i class="ti ti-clock"></i> Auto on temp access</span>
            <span class="badge badge-green">Active</span>
          </div>
          <div class="policy-item">
            <span class="policy-label"><i class="ti ti-database"></i> Keep last</span>
            <span style="font-size:13px;font-weight:600;color:var(--c-text)">50 backups / instance</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Backup List -->
    <div class="table-wrap">
      <div class="table-header">
        <span class="table-title">
          Backup History
          <span class="total-badge">{{ total }} backups</span>
        </span>
        <div style="display:flex;align-items:center;gap:8px">
          <select class="filter-select" v-model="filterInstance" @change="load" style="width:200px">
            <option value="">All Instances</option>
            <option v-for="i in instances" :key="i.id" :value="i.id">
              {{ i.instance_name }}:{{ i.port }}
            </option>
          </select>
          <button class="btn btn-ghost btn-sm" @click="load">
            <i class="ti ti-refresh"></i>
          </button>
        </div>
      </div>

      <table>
        <thead>
          <tr>
            <th>Filename</th>
            <th>Instance</th>
            <th>Trigger</th>
            <th>Note</th>
            <th>Size</th>
            <th>Created At</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="7" style="text-align:center;padding:32px;color:var(--c-muted)">
              <i class="ti ti-loader-2 spin" style="font-size:20px"></i>
            </td>
          </tr>
          <tr v-for="bk in backups" :key="bk.id" class="bk-row">
            <td style="font-family:monospace;font-size:11px">{{ bk.filename }}</td>
            <td style="font-size:11px">
              <span v-if="bk.instance">
                {{ bk.instance.instance_name }}:{{ bk.instance.port }}
                <span style="color:var(--c-muted)"> (PG {{ bk.instance.pg_version }})</span>
              </span>
              <span v-else style="color:var(--c-muted)">—</span>
            </td>
            <td>
              <span :class="['badge', triggerBadge(bk.trigger_event)]">
                {{ triggerLabel(bk.trigger_event) }}
              </span>
            </td>
            <td style="font-size:11px;color:var(--c-muted);max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
              :title="bk.note">
              {{ bk.note || '—' }}
            </td>
            <td style="font-size:11px;color:var(--c-muted);white-space:nowrap">
              {{ formatSize(bk.file_size) }}
            </td>
            <td style="font-size:11px;color:var(--c-muted);white-space:nowrap">
              {{ fmt(bk.created_at) }}
            </td>
            <td>
              <div style="display:flex;gap:4px">
                <button class="btn btn-ghost btn-sm" @click="previewBackup(bk)" title="Preview content">
                  <i class="ti ti-eye"></i>
                </button>
                <button class="btn btn-ghost btn-sm" @click="downloadBackup(bk)" title="Download">
                  <i class="ti ti-download"></i>
                </button>
                <button class="btn btn-ghost btn-sm" @click="confirmRestore(bk)" title="Restore"
                  style="color:var(--c-amber)">
                  <i class="ti ti-restore"></i>
                </button>
                <button class="btn btn-danger btn-sm" @click="deleteBackup(bk)" title="Delete">
                  <i class="ti ti-trash"></i>
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && backups.length === 0">
            <td colspan="7" style="text-align:center;padding:40px;color:var(--c-muted)">
              <i class="ti ti-device-floppy" style="font-size:28px;display:block;margin-bottom:8px"></i>
              No backups yet — create one above or edit/delete a rule to trigger auto-backup
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Preview Modal -->
    <Teleport to="body">
      <div v-if="previewBk" class="modal-backdrop" @click.self="previewBk = null">
        <div class="modal" style="width:700px;max-height:80vh;display:flex;flex-direction:column">
          <div class="modal-header">
            <div class="modal-title">
              <i class="ti ti-eye"></i> {{ previewBk.filename }}
            </div>
            <button class="btn btn-ghost btn-sm" @click="previewBk = null"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body" style="flex:1;overflow:auto">
            <div style="display:flex;gap:12px;margin-bottom:12px;flex-wrap:wrap">
              <div class="info-chip">
                <i class="ti ti-database"></i>
                {{ previewBk.instance?.instance_name }}:{{ previewBk.instance?.port }}
              </div>
              <div class="info-chip">
                <i class="ti ti-list"></i>
                {{ previewRuleCount }} rules
              </div>
              <div class="info-chip">
                <i class="ti ti-clock"></i>
                {{ fmt(previewBk.created_at) }}
              </div>
              <div v-if="previewBk.note" class="info-chip">
                <i class="ti ti-notes"></i>
                {{ previewBk.note }}
              </div>
            </div>
            <pre class="content-preview">{{ previewContent }}</pre>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="previewBk = null">Close</button>
            <button class="btn btn-ghost" @click="downloadBackup(previewBk!)">
              <i class="ti ti-download"></i> Download
            </button>
            <button class="btn btn-primary" @click="confirmRestore(previewBk!); previewBk = null"
              style="background:var(--c-amber);border-color:var(--c-amber)">
              <i class="ti ti-restore"></i> Restore This
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Restore Confirm Modal -->
    <Teleport to="body">
      <div v-if="restoreTarget" class="modal-backdrop" @click.self="restoreTarget = null">
        <div class="modal" style="width:480px">
          <div class="modal-header">
            <div class="modal-title" style="color:var(--c-amber)">
              <i class="ti ti-restore"></i> Confirm Restore
            </div>
            <button class="btn btn-ghost btn-sm" @click="restoreTarget = null"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body">
            <div class="warning-box">
              <i class="ti ti-alert-triangle"></i>
              <div>
                <div style="font-weight:600;margin-bottom:6px">Tindakan ini akan:</div>
                <ol style="margin-left:16px;line-height:1.9;font-size:12px;color:var(--c-muted)">
                  <li>Auto-backup kondisi saat ini terlebih dahulu</li>
                  <li>Menghapus semua rules saat ini untuk instance ini</li>
                  <li>Mengimpor rules dari backup <strong style="color:var(--c-text)">{{ restoreTarget.filename }}</strong></li>
                  <li>Menulis pg_hba.conf ke server + reload PostgreSQL</li>
                </ol>
              </div>
            </div>
            <div style="margin-top:12px;background:var(--c-bg);border:1px solid var(--c-border);border-radius:6px;padding:10px 12px;font-size:12px">
              <div style="color:var(--c-muted);margin-bottom:4px">Backup info:</div>
              <div style="font-family:monospace">{{ restoreTarget.filename }}</div>
              <div style="color:var(--c-muted);font-size:11px;margin-top:2px">
                Created: {{ fmt(restoreTarget.created_at) }}
                <span v-if="restoreTarget.note"> — {{ restoreTarget.note }}</span>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="restoreTarget = null">Cancel</button>
            <button class="btn btn-primary" @click="doRestore" :disabled="restoring"
              style="background:var(--c-amber);border-color:var(--c-amber)">
              <i :class="['ti', restoring ? 'ti-loader-2 spin' : 'ti-restore']"></i>
              {{ restoring ? 'Restoring...' : 'Restore Now' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useToast } from 'vue-toastification'
import api from '@/api'
import { instancesApi } from '@/api'
import type { PGInstance } from '@/types'
import dayjs from 'dayjs'

const toast = useToast()
const backups      = ref<any[]>([])
const instances    = ref<PGInstance[]>([])
const total        = ref(0)
const loading      = ref(false)
const creating     = ref(false)
const restoring    = ref(false)
const filterInstance = ref('')
const previewBk    = ref<any>(null)
const restoreTarget = ref<any>(null)
const previewContent = ref('')
const previewRuleCount = ref(0)

const createForm = ref({ instance_id: '', note: '' })

function fmt(d: string) { return dayjs(d).format('DD MMM YYYY HH:mm') }
function formatSize(bytes: number) {
  if (!bytes) return '—'
  if (bytes < 1024) return bytes + ' B'
  return (bytes / 1024).toFixed(1) + ' KB'
}

function triggerLabel(t: string) {
  const map: Record<string, string> = {
    manual: 'Manual', before_edit: 'Pre-Edit', before_delete: 'Pre-Delete',
    before_restore: 'Pre-Restore', auto: 'Auto', import: 'Import'
  }
  return map[t] || t
}

function triggerBadge(t: string) {
  if (t === 'manual') return 'badge-green'
  if (t === 'before_restore') return 'badge-purple'
  if (t === 'before_delete') return 'badge-red'
  return 'badge-gray'
}

async function load() {
  loading.value = true
  try {
    const params: any = {}
    if (filterInstance.value) params.instance_id = filterInstance.value
    const { data } = await api.get('/backups', { params })
    backups.value = data.data || []
    total.value   = data.total || 0
  } catch { toast.error('Failed to load backups') }
  finally { loading.value = false }
}

async function createBackup() {
  if (!createForm.value.instance_id) {
    toast.error('Please select an instance')
    return
  }
  creating.value = true
  try {
    const { data } = await api.post('/backups', createForm.value)
    toast.success(`Backup created: ${data.filename}`)
    createForm.value.note = ''
    await load()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Failed to create backup')
  } finally { creating.value = false }
}

async function previewBackup(bk: any) {
  previewContent.value = ''
  previewRuleCount.value = 0
  previewBk.value = bk
  try {
    const { data } = await api.get(`/backups/${bk.id}`)
    previewContent.value   = data.content || ''
    previewRuleCount.value = data.rule_count || 0
    previewBk.value = { ...bk, instance: data.instance }
  } catch {
    previewContent.value = bk.content || '(content not available)'
  }
}

function downloadBackup(bk: any) {
  const token = localStorage.getItem('access_token') || sessionStorage.getItem('access_token')
  const a = document.createElement('a')
  a.href = `/api/backups/${bk.id}/download`
  a.download = bk.filename
  // Add auth header via fetch
  fetch(`/api/backups/${bk.id}/download`, {
    headers: { Authorization: `Bearer ${token}` }
  }).then(r => r.blob()).then(blob => {
    const url = URL.createObjectURL(blob)
    a.href = url
    a.click()
    URL.revokeObjectURL(url)
  })
}

function confirmRestore(bk: any) {
  restoreTarget.value = bk
}

async function doRestore() {
  if (!restoreTarget.value) return
  restoring.value = true
  try {
    const { data } = await api.post(`/backups/${restoreTarget.value.id}/restore`)
    toast.success(`Restored from ${restoreTarget.value.filename} — ${data.imported} rules imported`)
    if (data.warning) toast.warning(data.warning)
    restoreTarget.value = null
    await load()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Restore failed')
  } finally { restoring.value = false }
}

async function deleteBackup(bk: any) {
  if (!confirm(`Delete backup "${bk.filename}"?`)) return
  try {
    await api.delete(`/backups/${bk.id}`)
    toast.success('Backup deleted')
    await load()
  } catch {
    toast.error('Failed to delete backup')
  }
}

onMounted(async () => {
  const { data } = await instancesApi.list()
  instances.value = data
  await load()
})
</script>

<style scoped>
.top-row { display:flex;gap:12px;margin-bottom:16px; }
@media (max-width:800px) { .top-row { flex-direction:column; } }

.card { background:var(--c-surface);border:1px solid var(--c-border);border-radius:10px;padding:16px; }
.card-title-row { margin-bottom:14px; }
.card-title-text { font-size:13px;font-weight:600;color:var(--c-text);display:flex;align-items:center;gap:8px; }

.policy-list { display:flex;flex-direction:column;gap:8px; }
.policy-item { display:flex;align-items:center;justify-content:space-between;padding:6px 0;border-bottom:1px solid var(--c-border); }
.policy-item:last-child { border-bottom:none; }
.policy-label { font-size:12px;color:var(--c-muted);display:flex;align-items:center;gap:6px; }

.total-badge { background:var(--c-bg);border:1px solid var(--c-border);border-radius:4px;padding:2px 8px;font-size:11px;color:var(--c-muted);margin-left:8px;font-weight:400; }
.bk-row:hover { background:var(--c-surface2); }

.badge-gray   { background:var(--c-surface2);color:var(--c-muted);border:1px solid var(--c-border);padding:2px 7px;border-radius:4px;font-size:10px; }
.badge-green  { background:rgba(34,197,94,.1);color:var(--c-green);border:1px solid rgba(34,197,94,.2);padding:2px 7px;border-radius:4px;font-size:10px; }
.badge-red    { background:rgba(239,68,68,.1);color:var(--c-red);border:1px solid rgba(239,68,68,.2);padding:2px 7px;border-radius:4px;font-size:10px; }
.badge-purple { background:rgba(139,92,246,.1);color:#8b5cf6;border:1px solid rgba(139,92,246,.2);padding:2px 7px;border-radius:4px;font-size:10px; }

.info-chip { display:inline-flex;align-items:center;gap:6px;background:var(--c-surface2);border:1px solid var(--c-border);border-radius:6px;padding:4px 10px;font-size:11px;color:var(--c-muted); }
.content-preview { background:var(--c-bg);border:1px solid var(--c-border);border-radius:6px;padding:12px;font-family:monospace;font-size:11px;color:var(--c-text);white-space:pre;overflow-x:auto;max-height:400px;overflow-y:auto; }

.warning-box { background:rgba(245,158,11,.08);border:1px solid rgba(245,158,11,.2);border-radius:8px;padding:14px;display:flex;gap:12px;color:var(--c-amber);font-size:12px; }

.spin { animation:spin 1s linear infinite; }
@keyframes spin { to { transform:rotate(360deg); } }
</style>
