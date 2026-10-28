<template>
  <div class="page-content">
    <div class="toolbar">
      <select class="filter-select" v-model="filterEnv" @change="load">
        <option value="">All Environments</option>
        <option>Production</option><option>Staging</option><option>UAT</option><option>Development</option>
      </select>
      <select class="filter-select" v-model="filterActive" @change="load">
        <option value="">All Status</option>
        <option value="true">Active</option>
        <option value="false">Inactive</option>
      </select>
      <div class="toolbar-right">
        <button class="btn btn-ghost" @click="discoverAll">
          <i class="ti ti-radar"></i> Discover Instances
        </button>
        <button class="btn btn-primary" @click="openCreate">
          <i class="ti ti-plus"></i> Add Host
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" style="text-align:center;padding:40px;color:var(--c-muted)">
      <i class="ti ti-loader-2 spin" style="font-size:24px"></i>
    </div>

    <!-- Host Grid -->
    <div v-else class="host-grid">
      <div v-for="host in hosts" :key="host.id" class="host-card" @click="openDetail(host)">
        <div class="host-card-header">
          <div style="display:flex;align-items:center;gap:8px">
            <span :class="['dot', host.is_active !== false ? 'dot-green' : 'dot-red']"></span>
            <span class="host-name">{{ host.name }}</span>
          </div>
          <span :class="['badge', envBadge(host.environment)]">{{ host.environment }}</span>
        </div>
        <div class="host-ip">{{ host.ip_address }}</div>

        <div class="host-meta">
          <div class="host-meta-item">
            <div class="host-meta-label">SSH Port</div>
            <div class="host-meta-value">{{ host.ssh_port || 22 }}</div>
          </div>
          <div class="host-meta-item">
            <div class="host-meta-label">SSH User</div>
            <div class="host-meta-value">{{ host.ssh_user }}</div>
          </div>
          <div class="host-meta-item">
            <div class="host-meta-label">PG Versions</div>
            <div class="host-meta-value">
              <span v-if="hostVersions(host).length">
                <span v-for="v in hostVersions(host)" :key="v"
                  style="display:inline-block;background:var(--c-bg);border:1px solid var(--c-border);border-radius:3px;padding:1px 5px;font-size:10px;font-family:monospace;margin-right:3px">
                  PG {{ v }}
                </span>
              </span>
              <span v-else style="color:var(--c-muted);font-size:11px">Not discovered</span>
            </div>
          </div>
          <div class="host-meta-item">
            <div class="host-meta-label">Instances</div>
            <div class="host-meta-value">{{ host.instance_count ?? host.instances?.length ?? 0 }}</div>
          </div>
        </div>

        <div class="host-actions" @click.stop>
          <button class="btn btn-ghost btn-sm" @click="openEdit(host)" title="Edit host">
            <i class="ti ti-edit"></i>
          </button>
          <button class="btn btn-ghost btn-sm" @click="openDeployKey(host)" title="Deploy SSH Key"
            style="color:var(--c-accent)">
            <i class="ti ti-key"></i>
          </button>
          <button class="btn btn-danger btn-sm" @click="deleteHost(host)" title="Delete host">
            <i class="ti ti-trash"></i>
          </button>
          <button class="btn btn-ghost btn-sm" @click="testConn(host)" title="Test connection">
            <i class="ti ti-plug"></i> Test
          </button>
          <button class="btn btn-ghost btn-sm" @click="resetHostKey(host)" title="Reset pinned SSH host key (only after a legitimate key rotation)">
            <i class="ti ti-shield-x"></i>
          </button>
        </div>
      </div>
    </div>

    <!-- Detail Modal -->
    <Teleport to="body">
      <div v-if="detailHost" class="modal-backdrop" @click.self="detailHost = null">
        <div class="modal" style="width:600px;max-height:80vh;display:flex;flex-direction:column">
          <div class="modal-header">
            <div class="modal-title">
              <span :class="['dot', detailHost.is_active !== false ? 'dot-green' : 'dot-red']"></span>
              {{ detailHost.name }}
            </div>
            <button class="btn btn-ghost btn-sm" @click="detailHost = null"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body" style="overflow-y:auto;flex:1">
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;margin-bottom:16px">
              <div class="detail-item"><div class="detail-label">IP Address</div><div class="detail-value mono">{{ detailHost.ip_address }}</div></div>
              <div class="detail-item"><div class="detail-label">SSH Port</div><div class="detail-value">{{ detailHost.ssh_port || 22 }}</div></div>
              <div class="detail-item"><div class="detail-label">SSH User</div><div class="detail-value">{{ detailHost.ssh_user }}</div></div>
              <div class="detail-item"><div class="detail-label">Environment</div>
                <div class="detail-value"><span :class="['badge', envBadge(detailHost.environment)]">{{ detailHost.environment }}</span></div>
              </div>
              <div class="detail-item"><div class="detail-label">Status</div>
                <div class="detail-value"><span :class="['badge', detailHost.is_active !== false ? 'badge-green' : 'badge-red']">{{ detailHost.is_active !== false ? 'Active' : 'Inactive' }}</span></div>
              </div>
              <div class="detail-item">
                <div class="detail-label">PG Versions</div>
                <div class="detail-value">
                  <span v-for="v in hostVersions(detailHost)" :key="v"
                    style="display:inline-block;background:var(--c-bg);border:1px solid var(--c-border);border-radius:3px;padding:2px 6px;font-size:11px;font-family:monospace;margin-right:4px">
                    PG {{ v }}
                  </span>
                  <span v-if="!hostVersions(detailHost).length" style="color:var(--c-muted);font-size:11px">Not discovered</span>
                </div>
              </div>
            </div>
            <div style="margin-bottom:12px">
              <div class="detail-label" style="margin-bottom:8px">INSTANCES</div>
              <table style="width:100%;font-size:12px;border-collapse:collapse">
                <thead>
                  <tr style="background:var(--c-bg)">
                    <th style="text-align:left;padding:6px 8px;color:var(--c-muted)">Name</th>
                    <th style="padding:6px 8px;color:var(--c-muted)">Port</th>
                    <th style="padding:6px 8px;color:var(--c-muted)">PG Ver</th>
                    <th style="padding:6px 8px;color:var(--c-muted)">Status</th>
                    <th style="padding:6px 8px;color:var(--c-muted)">Last Scan</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="inst in detailHost.instances" :key="inst.id" style="border-top:1px solid var(--c-border)">
                    <td style="padding:5px 8px;font-weight:500">{{ inst.instance_name }}</td>
                    <td style="padding:5px 8px;text-align:center;font-family:monospace">{{ inst.port }}</td>
                    <td style="padding:5px 8px;text-align:center"><span class="badge badge-gray">PG {{ inst.pg_version }}</span></td>
                    <td style="padding:5px 8px;text-align:center">
                      <span :class="['dot', inst.status === 'Online' ? 'dot-green' : 'dot-red']"></span>{{ inst.status }}
                    </td>
                    <td style="padding:5px 8px;color:var(--c-muted);font-size:11px">{{ inst.last_scan ? fmt(inst.last_scan) : '—' }}</td>
                  </tr>
                  <tr v-if="!detailHost.instances?.length">
                    <td colspan="5" style="color:var(--c-muted);text-align:center;padding:16px">No instances discovered yet</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="detailHost = null">Close</button>
            <button class="btn btn-primary" @click="discoverHost(detailHost); detailHost = null">
              <i class="ti ti-radar"></i> Discover Instances
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Add/Edit Host Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
        <div class="modal" style="width:520px">
          <div class="modal-header">
            <div class="modal-title">{{ editing ? 'Edit Host' : 'Add Database Host' }}</div>
            <button class="btn btn-ghost btn-sm" @click="showModal = false"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body">
            <div class="field-row">
              <div class="field">
                <label>Host Name *</label>
                <input v-model="form.name" placeholder="pg-prod-01">
              </div>
              <div class="field">
                <label>IP Address *</label>
                <input v-model="form.ip_address" placeholder="10.10.1.10">
              </div>
            </div>
            <div class="field-row">
              <div class="field">
                <label>SSH Port</label>
                <input v-model.number="form.ssh_port" type="number" placeholder="22">
              </div>
              <div class="field">
                <label>SSH User</label>
                <input v-model="form.ssh_user" placeholder="dco">
              </div>
            </div>
            <div class="field">
              <label>SSH Private Key (will be AES-256 encrypted)</label>
              <textarea v-model="form.ssh_key"
                placeholder="-----BEGIN RSA PRIVATE KEY-----&#10;...&#10;-----END RSA PRIVATE KEY-----"
                style="height:80px;font-family:monospace;font-size:11px"></textarea>
              <div style="font-size:11px;color:var(--c-muted);margin-top:4px">
                Leave blank to inherit from existing host. Or use Deploy SSH Key after adding.
              </div>
            </div>
            <div v-if="connTestError" style="background:rgba(239,68,68,.08);border:1px solid rgba(239,68,68,.2);border-radius:8px;padding:10px 12px;font-size:12px;color:var(--c-red);margin-top:4px">
              <i class="ti ti-alert-circle"></i> SSH test gagal: {{ connTestError }}
            </div>
            <div class="field-row">
              <div class="field">
                <label>Environment</label>
                <select v-model="form.environment">
                  <option>Production</option><option>Staging</option><option>UAT</option><option>Development</option>
                </select>
              </div>
              <div class="field">
                <label>Notes</label>
                <input v-model="form.notes" placeholder="Optional notes about this host">
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="showModal = false">Cancel</button>
            <button class="btn btn-primary" @click="saveHost" :disabled="saving">
              <i :class="['ti', saving ? 'ti-loader-2 spin' : 'ti-check']"></i>
              {{ saving ? 'Testing SSH...' : (editing ? 'Save Changes' : 'Add Host') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Deploy SSH Key Modal -->
    <Teleport to="body">
      <div v-if="deployKeyTarget" class="modal-backdrop" @click.self="deployKeyTarget = null; deployResult = null">
        <div class="modal" style="width:500px">
          <div class="modal-header">
            <div class="modal-title">
              <i class="ti ti-key" style="color:var(--c-accent)"></i>
              Deploy SSH Key — {{ deployKeyTarget.name }}
            </div>
            <button class="btn btn-ghost btn-sm" @click="deployKeyTarget = null; deployResult = null">
              <i class="ti ti-x"></i>
            </button>
          </div>
          <div class="modal-body">
            <div v-if="!deployResult">
              <div style="background:rgba(79,142,247,.08);border:1px solid rgba(79,142,247,.2);border-radius:8px;padding:12px 14px;font-size:12px;margin-bottom:14px;color:var(--c-muted)">
                <i class="ti ti-info-circle" style="color:var(--c-accent)"></i>
                App akan SSH ke <strong style="color:var(--c-text)">{{ deployKeyTarget.ip_address }}</strong> sebagai root,
                lalu copy public key ke user postgres.
              </div>
              <div class="field-row">
                <div class="field">
                  <label>Login As</label>
                  <input v-model="deployKeyForm.ssh_user" placeholder="root">
                </div>
                <div class="field">
                  <label>Ops User (SSH, bukan postgres)</label>
                  <input v-model="deployKeyForm.ops_user" placeholder="dco">
                </div>
              </div>
              <div class="field">
                <label>Password untuk {{ deployKeyForm.ssh_user }}@{{ deployKeyTarget.ip_address }}</label>
                <input v-model="deployKeyForm.root_password" type="password"
                  placeholder="Enter SSH password..." autocomplete="off">
              </div>
              <div style="font-size:11px;color:var(--c-muted);margin-top:4px">
                <i class="ti ti-lock" style="font-size:10px"></i>
                Password hanya digunakan sekali dan tidak disimpan.
              </div>
            </div>
            <div v-if="deployResult && !deployResult.error">
              <div :style="{ background: deployResult.ready ? 'rgba(34,197,94,.08)' : 'rgba(245,158,11,.08)', border: '1px solid ' + (deployResult.ready ? 'rgba(34,197,94,.2)' : 'rgba(245,158,11,.2)'), borderRadius: '8px', padding: '12px', marginBottom: '12px' }">
                <div style="font-weight:600;margin-bottom:6px;font-size:13px">
                  <i :class="['ti', deployResult.ready ? 'ti-circle-check' : 'ti-alert-triangle']"
                    :style="{ color: deployResult.ready ? 'var(--c-green)' : 'var(--c-amber)' }"></i>
                  {{ deployResult.ready ? 'SSH Key Deployed!' : 'Key Deployed — Test Failed' }}
                </div>
                <div style="font-size:11px;color:var(--c-muted)">{{ deployResult.test_result }}</div>
              </div>
              <div style="font-size:11px;font-family:monospace;background:var(--c-bg);border:1px solid var(--c-border);border-radius:6px;padding:10px;max-height:160px;overflow-y:auto">
                <div v-if="deployResult.sudoers" style="color:var(--c-amber);margin-bottom:6px">sudoers: {{ deployResult.sudoers }}</div>
                <div v-for="cmd in deployResult.commands" :key="cmd" style="color:var(--c-green);margin-bottom:2px">{{ cmd }}</div>
              </div>
            </div>
            <div v-if="deployResult?.error" style="background:rgba(239,68,68,.08);border:1px solid rgba(239,68,68,.2);border-radius:8px;padding:12px;font-size:12px;color:var(--c-red)">
              <i class="ti ti-alert-circle"></i> {{ deployResult.error }}
              <div v-if="deployResult.detail" style="color:var(--c-muted);margin-top:4px;font-size:11px">{{ deployResult.detail }}</div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="deployKeyTarget = null; deployResult = null">
              {{ deployResult ? 'Close' : 'Cancel' }}
            </button>
            <button v-if="!deployResult || deployResult.error" class="btn btn-primary"
              @click="doDeployKey" :disabled="deployingKey || !deployKeyForm.root_password">
              <i :class="['ti', deployingKey ? 'ti-loader-2 spin' : 'ti-key']"></i>
              {{ deployingKey ? 'Deploying...' : 'Deploy SSH Key' }}
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
import { hostsApi } from '@/api'
import api from '@/api'
import { globalSearch } from '@/stores/search'
import type { DBHost, CreateHostPayload, EnvType } from '@/types'
import dayjs from 'dayjs'

const toast = useToast()
const hosts = ref<DBHost[]>([])
const loading = ref(false)
const saving = ref(false)
const showModal = ref(false)
const editing = ref<DBHost | null>(null)
const filteredHosts = computed(() => {
  const q = globalSearch.value.toLowerCase().trim()
  if (!q) return hosts.value
  return hosts.value.filter(h =>
    h.name?.toLowerCase().includes(q) ||
    h.ip_address?.toLowerCase().includes(q) ||
    h.environment?.toLowerCase().includes(q) ||
    h.notes?.toLowerCase().includes(q)
  )
})
const deployKeyTarget = ref<DBHost | null>(null)
const deployKeyForm = ref({ root_password: '', ssh_user: 'root', ops_user: 'dco' })
const deployingKey = ref(false)
const deployResult = ref<any>(null)
const detailHost = ref<DBHost | null>(null)
const filterEnv = ref('')
const filterActive = ref('')

const defaultForm = (): CreateHostPayload => ({
  name: '', ip_address: '', ssh_port: 22, ssh_user: 'dco',
  ssh_key: '', environment: 'Production' as EnvType, notes: '',
})
const form = ref(defaultForm())
const connTestError = ref('')

function envBadge(env: string) {
  return env === 'Production' ? 'badge-red' : env === 'UAT' ? 'badge-amber' : 'badge-blue'
}
function fmt(d: string) { return dayjs(d).format('DD MMM HH:mm') }

async function load() {
  loading.value = true
  try {
    const params: any = {}
    if (filterEnv.value)    params.environment = filterEnv.value
    if (filterActive.value) params.active = filterActive.value === 'true'
    const { data } = await hostsApi.list(params)
    hosts.value = data
  } catch { toast.error('Failed to load hosts') }
  finally { loading.value = false }
}

function openCreate() { editing.value = null; form.value = defaultForm(); connTestError.value = ''; showModal.value = true }
function openEdit(host: DBHost) {
  editing.value = host
  form.value = {
    name: host.name, ip_address: host.ip_address, ssh_port: host.ssh_port,
    ssh_user: host.ssh_user, ssh_key: '',
    environment: host.environment, notes: host.notes,
  }
  connTestError.value = ''
  showModal.value = true
}
function openDetail(host: DBHost) { detailHost.value = host }

async function saveHost() {
  if (!form.value.name || !form.value.ip_address) {
    toast.error('Name, IP Address, and PG Version are required'); return
  }
  connTestError.value = ''
  saving.value = true
  try {
    // Pre-flight: refuse to save a host whose SSH connection doesn't
    // actually work — catches a wrong IP, closed port, firewall, or an
    // unauthorized key before it's discovered later during a write/reload.
    const { data: test } = await hostsApi.validate({
      host_id: editing.value?.id,
      ip_address: form.value.ip_address,
      ssh_port: form.value.ssh_port,
      ssh_user: form.value.ssh_user,
      ssh_key: form.value.ssh_key,
    })
    if (!test.ok) {
      connTestError.value = test.error || 'SSH connection test failed'
      toast.error(connTestError.value)
      return
    }

    if (editing.value) {
      await hostsApi.update(editing.value.id, form.value)
      toast.success('Host updated')
    } else {
      await hostsApi.create(form.value)
      toast.success('Host added. Discovering instances...')
    }
    showModal.value = false
    await load()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Failed to save host')
  } finally { saving.value = false }
}

function hostVersions(host: any): string[] {
  if (!host) return []
  const instances = (host as any).instances || []
  const versions = [...new Set(instances.map((i: any) => i.pg_version).filter(Boolean))] as string[]
  return versions.sort()
}

function openDeployKey(host: DBHost) {
  deployKeyTarget.value = host
  deployKeyForm.value = { root_password: '', ssh_user: 'root', ops_user: (host.ssh_user && host.ssh_user !== 'postgres') ? host.ssh_user : 'dco' }
  deployResult.value = null
}

async function doDeployKey() {
  if (!deployKeyTarget.value) return
  deployingKey.value = true
  deployResult.value = null
  try {
    const { data } = await api.post(`/hosts/${deployKeyTarget.value.id}/deploy-key`, deployKeyForm.value)
    deployResult.value = data
    if (data.ready) {
      toast.success(`SSH key deployed to ${deployKeyTarget.value.name} — ready to use!`)
    } else {
      toast.warning(`Key deployed but SSH test failed — check manually`)
    }
  } catch (e: any) {
    deployResult.value = { error: e.response?.data?.error || 'Deploy failed', detail: e.response?.data?.detail }
    toast.error(e.response?.data?.error || 'Deploy SSH key failed')
  } finally {
    deployingKey.value = false
    deployKeyForm.value.root_password = '' // Clear password from memory
  }
}

async function deleteHost(host: DBHost) {
  if (!confirm(`Delete host "${host.name}"?\n\nAll associated instances and rules will be removed.`)) return
  try {
    await hostsApi.delete(host.id)
    toast.success('Host deleted')
    await load()
  } catch { toast.error('Failed to delete host') }
}

async function resetHostKey(host: DBHost) {
  if (!confirm(
    `Reset pinned SSH host key for "${host.name}"?\n\n` +
    `Only do this if you KNOW this host legitimately rebuilt/rotated its SSH key ` +
    `(reinstalled OS, replaced server, etc). The next connection will trust and ` +
    `pin whatever key it sees — if you're not sure why connections to this host ` +
    `started failing, verify with the host's admin first instead of resetting blindly.`
  )) return
  try {
    await hostsApi.resetHostKey(host.id)
    toast.success(`Host key fingerprint cleared for ${host.name}`)
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Failed to reset host key')
  }
}

async function testConn(host: DBHost) {
  toast.info(`Testing SSH connection to ${host.ip_address}:${host.ssh_port}...`)
  try {
    const { data } = await hostsApi.validate({
      host_id: host.id, ip_address: host.ip_address,
      ssh_port: host.ssh_port, ssh_user: host.ssh_user,
    })
    if (data.ok) toast.success(`${host.name} — ${data.message}`)
    else toast.error(`${host.name} — ${data.error}`)
  } catch (e: any) {
    toast.error(e.response?.data?.error || `${host.name} — SSH test failed`)
  }
}

async function discoverAll() {
  toast.info('Discovering instances on all active hosts...')
  for (const host of hosts.value) {
    try { await api.post('/instances/discover', { host_id: host.id }) } catch {}
  }
  await load()
  toast.success('Discovery complete for all hosts')
}
async function discoverHost(host: DBHost) {
  toast.info(`Discovering instances on ${host.name}...`)
  try {
    const { data } = await api.post('/instances/discover', { host_id: host.id })
    toast.success(`${host.name} — found ${data.upserted} instance(s)`)
    await load()
    const { data: updated } = await hostsApi.get(host.id)
    detailHost.value = updated
  } catch (e: any) {
    const detail = e.response?.data?.detail || e.response?.data?.error || 'Discovery failed'
    toast.error(`${host.name}: ${detail}`)
    detailHost.value = null
  }
}

onMounted(load)
</script>


<style scoped>
.host-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:16px; }
.host-card { background:var(--c-surface);border:1px solid var(--c-border);border-radius:10px;padding:16px;cursor:pointer;transition:border-color .15s; }
.host-card:hover { border-color:var(--c-accent); }
.host-card-header { display:flex;align-items:center;justify-content:space-between;margin-bottom:4px; }
.host-name { font-size:14px;font-weight:600;color:var(--c-text); }
.host-ip { font-size:11px;font-family:monospace;color:var(--c-muted);margin-bottom:12px; }
.host-meta { display:grid;grid-template-columns:1fr 1fr;gap:8px;margin-bottom:12px; }
.host-meta-item { }
.host-meta-label { font-size:10px;color:var(--c-muted);text-transform:uppercase;letter-spacing:.3px;margin-bottom:2px; }
.host-meta-value { font-size:12px;font-weight:500;color:var(--c-text); }
.host-actions { display:flex;gap:4px;border-top:1px solid var(--c-border);padding-top:10px; }
.detail-item { background:var(--c-bg);border:1px solid var(--c-border);border-radius:6px;padding:8px 10px; }
.detail-label { font-size:10px;color:var(--c-muted);text-transform:uppercase;letter-spacing:.3px;margin-bottom:3px; }
.detail-value { font-size:12px;font-weight:500; }
.mono { font-family:monospace; }
.spin { animation:spin 1s linear infinite; }
@keyframes spin { to { transform:rotate(360deg); } }
</style>
