<template>
  <div class="page-content">
    <!-- Environment Filter -->
    <div class="env-tabs">
      <button v-for="env in envOptions" :key="env.value"
        :class="['env-tab', selectedEnv === env.value ? 'active' : '']"
        @click="selectEnv(env.value)">
        <i :class="env.icon"></i>
        {{ env.label }}
      </button>
    </div>

    <!-- Summary Cards -->
    <div class="stat-grid">
      <div class="stat-card">
        <div class="stat-icon blue"><i class="ti ti-server"></i></div>
        <div class="stat-body">
          <div class="stat-num">{{ stats.total_hosts ?? 0 }}</div>
          <div class="stat-label">Database Hosts</div>
          <div class="stat-sub">{{ selectedEnv || 'All environments' }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon green"><i class="ti ti-database"></i></div>
        <div class="stat-body">
          <div class="stat-num">{{ stats.total_instances ?? 0 }}</div>
          <div class="stat-label">PG Instances</div>
          <div class="stat-sub">{{ stats.online_instances ?? 0 }} online</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon purple"><i class="ti ti-shield-lock"></i></div>
        <div class="stat-body">
          <div class="stat-num">{{ stats.total_rules ?? 0 }}</div>
          <div class="stat-label">Access Rules</div>
          <div class="stat-sub">in pg_hba.conf</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon amber"><i class="ti ti-clock"></i></div>
        <div class="stat-body">
          <div class="stat-num">{{ stats.temp_rules_active ?? 0 }}</div>
          <div class="stat-label">Active Temp Access</div>
          <div class="stat-sub">auto-expires</div>
        </div>
      </div>
    </div>

    <div class="dash-row">
      <!-- Environment Breakdown -->
      <div class="dash-card" style="flex:1.2">
        <div class="dash-card-title">
          <i class="ti ti-layers-intersect"></i> Environment Overview
        </div>
        <div v-if="!stats.env_stats?.length" class="empty-state">No data</div>
        <table v-else class="env-table">
          <thead>
            <tr>
              <th>Environment</th>
              <th>Hosts</th>
              <th>Instances</th>
              <th>Online</th>
              <th>Rules</th>
              <th>Health</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="e in stats.env_stats" :key="e.environment"
              :class="['env-row', selectedEnv === e.environment ? 'active' : '']"
              @click="selectEnv(selectedEnv === e.environment ? '' : e.environment)"
              style="cursor:pointer">
              <td>
                <span :class="['env-badge', envBadgeClass(e.environment)]">
                  {{ e.environment || 'Unknown' }}
                </span>
              </td>
              <td>{{ e.hosts }}</td>
              <td>{{ e.instances }}</td>
              <td>
                <span class="dot dot-green"></span>{{ e.online }}
                <span v-if="e.instances - e.online > 0" style="color:var(--c-red);margin-left:6px">
                  <span class="dot dot-red"></span>{{ e.instances - e.online }}
                </span>
              </td>
              <td>{{ e.rules }}</td>
              <td>
                <div class="health-bar">
                  <div class="health-fill"
                    :style="{ width: e.instances ? (e.online/e.instances*100)+'%' : '100%',
                              background: e.instances && e.online < e.instances ? 'var(--c-amber)' : 'var(--c-green)' }">
                  </div>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- PG Version Distribution -->
      <div class="dash-card" style="flex:0.8">
        <div class="dash-card-title">
          <i class="ti ti-chart-bar"></i> PostgreSQL Versions
        </div>
        <div v-if="!stats.version_dist?.length" class="empty-state">No instances</div>
        <div v-else class="ver-list">
          <div v-for="v in stats.version_dist" :key="v.pg_version" class="ver-item">
            <div class="ver-label">
              <span class="badge badge-gray">PG {{ v.pg_version }}</span>
              <span class="ver-count">{{ v.count }}</span>
            </div>
            <div class="ver-bar">
              <div class="ver-fill" :style="{
                width: maxVerCount ? (v.count/maxVerCount*100)+'%' : '0%',
                background: pgVerColor(v.pg_version)
              }"></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Trend Charts Row -->
    <div class="dash-row">
      <!-- Rules trend -->
      <div class="dash-card" style="flex:1">
        <div class="dash-card-title">
          <i class="ti ti-chart-line"></i> Rule Changes (Last 7 Days)
        </div>
        <div class="chart-wrap">
          <div v-for="(day, i) in trendData" :key="i" class="chart-bar-col">
            <div class="chart-bar-stack">
              <div class="chart-bar-val" :style="{
                height: day.count ? (day.count / maxTrend * 80) + 'px' : '2px',
                background: day.type === 'add' ? 'var(--c-green)' : day.type === 'delete' ? 'var(--c-red)' : 'var(--c-accent)'
              }" :title="day.label + ': ' + day.count"></div>
            </div>
            <div class="chart-bar-label">{{ day.day }}</div>
          </div>
        </div>
        <div class="chart-legend">
          <span class="legend-item green"><i class="ti ti-plus"></i> Added</span>
          <span class="legend-item red"><i class="ti ti-minus"></i> Deleted</span>
          <span class="legend-item blue"><i class="ti ti-edit"></i> Edited</span>
        </div>
      </div>

      <!-- Host availability -->
      <div class="dash-card" style="flex:1">
        <div class="dash-card-title">
          <i class="ti ti-server-2"></i> Host Availability
        </div>
        <div v-if="!stats.env_stats?.length" class="empty-state">No data</div>
        <div v-else class="host-avail-list">
          <div v-for="env in stats.env_stats" :key="env.environment" class="host-avail-item">
            <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:4px">
              <span :class="['env-badge', envBadgeClass(env.environment)]">{{ env.environment }}</span>
              <span style="font-size:11px;color:var(--c-muted)">
                {{ env.online }}/{{ env.instances }} online
              </span>
            </div>
            <div class="avail-bar">
              <div class="avail-fill"
                :style="{
                  width: env.instances ? (env.online/env.instances*100)+'%' : '100%',
                  background: env.instances && env.online < env.instances ? 'var(--c-amber)' : 'var(--c-green)'
                }">
              </div>
            </div>
            <div style="font-size:10px;color:var(--c-muted);margin-top:2px">
              {{ env.rules }} rules · {{ env.hosts }} hosts
            </div>
          </div>
        </div>

        <!-- Temp access active -->
        <div style="margin-top:14px;padding-top:12px;border-top:1px solid var(--c-border)">
          <div class="dash-card-title" style="margin-bottom:8px;font-size:11px">
            ACTIVE TEMP ACCESS
          </div>
          <div style="display:flex;align-items:center;gap:12px">
            <div style="font-size:32px;font-weight:700;color:var(--c-amber)">
              {{ stats.temp_rules_active ?? 0 }}
            </div>
            <div style="font-size:11px;color:var(--c-muted)">
              <div>temporary grants</div>
              <div>auto-expires by scheduler</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    
    <!-- Health Check -->
    <div class="dash-card" style="margin-bottom:16px">
      <div class="dash-card-title" style="display:flex;align-items:center;justify-content:space-between">
        <span><i class="ti ti-heartbeat" style="color:var(--c-green)"></i> Host Health Status</span>
        <div style="display:flex;align-items:center;gap:8px">
          <span v-if="lastHealthCheck" style="font-size:10px;color:var(--c-muted)">
            Last: {{ lastHealthCheck }}
          </span>
          <button class="btn btn-ghost btn-sm" @click="runHealthCheck" :disabled="checkingHealth">
            <i :class="['ti', checkingHealth ? 'ti-loader-2 spin' : 'ti-refresh']"></i>
            {{ checkingHealth ? 'Checking...' : 'Check Now' }}
          </button>
        </div>
      </div>

      <div v-if="!healthResults.length" style="text-align:center;padding:20px;color:var(--c-muted);font-size:12px">
        <i class="ti ti-heartbeat" style="font-size:24px;display:block;margin-bottom:6px"></i>
        Click "Check Now" to run health check on all hosts
      </div>

      <div v-else class="health-grid">
        <div v-for="h in healthResults" :key="h.host_id"
          :class="['health-card', h.status === 'online' ? 'health-ok' : h.status === 'offline' ? 'health-err' : 'health-warn']">
          <div class="health-header">
            <span :class="['dot', h.status === 'online' ? 'dot-green' : h.status === 'offline' ? 'dot-red' : 'dot-amber']"></span>
            <span class="health-name">{{ h.host_name }}</span>
            <span :class="['badge', h.status === 'online' ? 'badge-green' : h.status === 'offline' ? 'badge-red' : 'badge-amber']"
              style="margin-left:auto;font-size:9px">
              {{ h.status.toUpperCase() }}
            </span>
          </div>
          <div class="health-meta">
            <span class="mono" style="font-size:10px">{{ h.ip_address }}</span>
            <span style="font-size:10px;color:var(--c-muted)">{{ h.environment }}</span>
          </div>
          <div class="health-stats">
            <span style="color:var(--c-green);font-size:11px">
              <i class="ti ti-circle-check"></i> {{ h.online }}/{{ h.instances }}
            </span>
            <span v-if="h.offline > 0" style="color:var(--c-red);font-size:11px">
              <i class="ti ti-circle-x"></i> {{ h.offline }} offline
            </span>
            <span v-if="h.response_ms" style="font-size:10px;color:var(--c-muted);margin-left:auto">
              {{ h.response_ms }}ms
            </span>
          </div>
          <div v-if="h.error" style="font-size:10px;color:var(--c-red);margin-top:4px;word-break:break-all">
            {{ h.error }}
          </div>
          <div v-if="h.changed" style="font-size:10px;color:var(--c-amber);margin-top:2px">
            <i class="ti ti-alert-triangle"></i> Status changed! Notification sent.
          </div>
        </div>
      </div>
    </div>

    <!-- Recent Activity -->
    <div class="dash-card">
      <div class="dash-card-title">
        <i class="ti ti-activity"></i> Recent Activity
      </div>
      <div v-if="!stats.recent_activity?.length" class="empty-state">No recent activity</div>
      <table v-else style="width:100%">
        <thead>
          <tr>
            <th>Event</th>
            <th>User</th>
            <th>Target</th>
            <th>Instance</th>
            <th>Time</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in stats.recent_activity" :key="log.id">
            <td>
              <span :class="['event-badge', eventClass(log.event_type)]">
                {{ log.event_type }}
              </span>
            </td>
            <td style="font-size:12px">{{ log.username }}</td>
            <td class="mono" style="font-size:11px;max-width:200px;overflow:hidden;text-overflow:ellipsis">
              {{ log.target || '—' }}
            </td>
            <td style="font-size:11px;color:var(--c-muted)">{{ log.instance_label || '—' }}</td>
            <td style="font-size:11px;color:var(--c-muted);white-space:nowrap">
              {{ fmt(log.created_at) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api'
import dayjs from 'dayjs'

const router = useRouter()
const stats = ref<any>({})
const selectedEnv = ref('')

const envOptions = [
  { value: '', label: 'All', icon: 'ti ti-layout-grid' },
  { value: 'Production', label: 'Production', icon: 'ti ti-server-2' },
  { value: 'Staging', label: 'Staging', icon: 'ti ti-test-pipe' },
  { value: 'UAT', label: 'UAT', icon: 'ti ti-checklist' },
  { value: 'Development', label: 'Development', icon: 'ti ti-code' },
]

const maxVerCount = computed(() => {
  if (!stats.value.version_dist?.length) return 1
  return Math.max(...stats.value.version_dist.map((v: any) => v.count))
})

function envBadgeClass(env: string) {
  if (env === 'Production') return 'env-prod'
  if (env === 'Staging') return 'env-staging'
  if (env === 'UAT') return 'env-uat'
  if (env === 'Development') return 'env-dev'
  return 'env-unknown'
}

function pgVerColor(ver: string) {
  const v = parseInt(ver)
  if (v >= 16) return 'var(--c-accent)'
  if (v >= 14) return 'var(--c-teal)'
  if (v >= 12) return 'var(--c-green)'
  return 'var(--c-muted)'
}

function eventClass(event: string) {
  if (event?.includes('Add') || event?.includes('Grant')) return 'ev-add'
  if (event?.includes('Delete') || event?.includes('Revoke') || event?.includes('Expire')) return 'ev-del'
  if (event?.includes('Edit') || event?.includes('Update')) return 'ev-edit'
  if (event?.includes('Reload') || event?.includes('Write')) return 'ev-reload'
  if (event?.includes('Disable')) return 'ev-warn'
  return 'ev-default'
}

function fmt(d: string) { return dayjs(d).format('DD MMM HH:mm') }

const trendData = computed(() => {
  // Generate last 7 days from audit log activity
  const days = []
  const now = new Date()
  for (let i = 6; i >= 0; i--) {
    const d = new Date(now)
    d.setDate(d.getDate() - i)
    const dayStr = d.toLocaleDateString('en', { weekday: 'short' })
    const dateStr = d.toISOString().slice(0, 10)

    // Count from recent_activity if available
    const activity = stats.value.recent_activity || []
    const dayLogs = activity.filter((l: any) => l.created_at?.startsWith(dateStr))
    const added = dayLogs.filter((l: any) => l.event === 'AddRule').length
    const deleted = dayLogs.filter((l: any) => l.event === 'DeleteRule').length
    const edited = dayLogs.filter((l: any) => l.event === 'EditRule').length
    const total = added + deleted + edited

    days.push({ day: dayStr, count: total, date: dateStr,
      label: dateStr, type: added > deleted ? 'add' : deleted > 0 ? 'delete' : 'edit' })
  }
  return days
})

const maxTrend = computed(() => {
  const vals = trendData.value.map(d => d.count)
  return Math.max(...vals, 1)
})

const healthResults    = ref<any[]>([])
const checkingHealth   = ref(false)
const lastHealthCheck  = ref('')

async function runHealthCheck() {
  checkingHealth.value = true
  try {
    const { data } = await api.get('/health/check')
    healthResults.value = data.hosts || []
    lastHealthCheck.value = new Date().toLocaleTimeString('id-ID')
    const offline = data.offline || 0
    if (offline > 0) {
      toast.error(`${offline} host(s) offline! Check Health Status panel.`)
    } else {
      toast.success(`All ${data.online} hosts online (${data.total} checked)`)
    }
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Health check failed')
  } finally {
    checkingHealth.value = false
  }
}

async function load() {
  const params = selectedEnv.value ? { environment: selectedEnv.value } : {}
  const { data } = await api.get('/dashboard/stats', { params })
  stats.value = data
}

function selectEnv(env: string) {
  selectedEnv.value = env
  load()
}

onMounted(load)
</script>

<style scoped>
.env-tabs { display:flex;gap:6px;margin-bottom:20px;flex-wrap:wrap; }
.env-tab { display:flex;align-items:center;gap:6px;padding:7px 14px;border-radius:8px;border:1px solid var(--c-border);background:var(--c-surface);color:var(--c-muted);cursor:pointer;font-size:12px;font-family:var(--font);transition:all .15s; }
.env-tab:hover { background:var(--c-surface2);color:var(--c-text); }
.env-tab.active { background:var(--c-accent);border-color:var(--c-accent);color:#fff; }
.env-tab i { font-size:14px; }

.stat-grid { display:grid;grid-template-columns:repeat(4,1fr);gap:12px;margin-bottom:16px; }
.stat-card { background:var(--c-surface);border:1px solid var(--c-border);border-radius:10px;padding:16px;display:flex;gap:14px;align-items:center; }
.stat-icon { width:44px;height:44px;border-radius:10px;display:flex;align-items:center;justify-content:center;font-size:20px;flex-shrink:0; }
.stat-icon.blue   { background:rgba(79,142,247,.15);color:var(--c-accent); }
.stat-icon.green  { background:rgba(34,197,94,.15);color:var(--c-green); }
.stat-icon.purple { background:rgba(139,92,246,.15);color:#8b5cf6; }
.stat-icon.amber  { background:rgba(245,158,11,.15);color:var(--c-amber); }
.stat-num { font-size:28px;font-weight:700;color:var(--c-text);line-height:1; }
.stat-label { font-size:12px;color:var(--c-muted);margin-top:2px; }
.stat-sub { font-size:10px;color:var(--c-muted);margin-top:2px; }

.dash-row { display:flex;gap:12px;margin-bottom:16px; }
.dash-card { background:var(--c-surface);border:1px solid var(--c-border);border-radius:10px;padding:16px;margin-bottom:16px; }
.dash-card-title { font-size:13px;font-weight:600;color:var(--c-text);margin-bottom:14px;display:flex;align-items:center;gap:8px; }
.empty-state { text-align:center;padding:24px;color:var(--c-muted);font-size:12px; }

.env-table { width:100%;border-collapse:collapse;font-size:12px; }
.env-table th { text-align:left;padding:6px 10px;font-size:10px;color:var(--c-muted);text-transform:uppercase;letter-spacing:.5px;border-bottom:1px solid var(--c-border); }
.env-table td { padding:8px 10px;border-bottom:1px solid var(--c-border); }
.env-row:hover { background:var(--c-surface2); }
.env-row.active { background:rgba(79,142,247,.08); }
.env-badge { padding:3px 8px;border-radius:4px;font-size:11px;font-weight:500; }
.env-prod    { background:rgba(239,68,68,.15);color:var(--c-red); }
.env-staging { background:rgba(245,158,11,.15);color:var(--c-amber); }
.env-uat     { background:rgba(139,92,246,.15);color:#8b5cf6; }
.env-dev     { background:rgba(34,197,94,.15);color:var(--c-green); }
.env-unknown { background:var(--c-surface2);color:var(--c-muted); }

.health-bar { height:5px;background:var(--c-border);border-radius:3px;width:80px;overflow:hidden; }
.health-fill { height:100%;border-radius:3px;transition:width .3s; }

.ver-list { display:flex;flex-direction:column;gap:10px; }
.ver-item { display:flex;flex-direction:column;gap:4px; }
.ver-label { display:flex;justify-content:space-between;align-items:center; }
.ver-count { font-size:13px;font-weight:600;color:var(--c-text); }
.ver-bar { height:6px;background:var(--c-border);border-radius:3px;overflow:hidden; }
.ver-fill { height:100%;border-radius:3px;transition:width .4s; }

.event-badge { padding:2px 8px;border-radius:4px;font-size:10px;font-weight:500; }
.ev-add    { background:rgba(34,197,94,.1);color:var(--c-green); }
.ev-del    { background:rgba(239,68,68,.1);color:var(--c-red); }
.ev-edit   { background:rgba(79,142,247,.1);color:var(--c-accent); }
.ev-reload { background:rgba(20,184,166,.1);color:var(--c-teal); }
.ev-warn   { background:rgba(245,158,11,.1);color:var(--c-amber); }
.ev-default { background:var(--c-surface2);color:var(--c-muted); }
.mono { font-family:monospace; }

/* Charts */
.chart-wrap { display:flex;gap:4px;align-items:flex-end;height:100px;padding:8px 0; }
.chart-bar-col { flex:1;display:flex;flex-direction:column;align-items:center;gap:4px; }
.chart-bar-stack { display:flex;align-items:flex-end;height:80px; }
.chart-bar-val { width:100%;min-height:2px;border-radius:3px 3px 0 0;transition:height .3s;min-width:12px; }
.chart-bar-label { font-size:9px;color:var(--c-muted);text-align:center; }
.chart-legend { display:flex;gap:12px;margin-top:8px;flex-wrap:wrap; }
.legend-item { font-size:10px;display:flex;align-items:center;gap:4px;color:var(--c-muted); }
.legend-item.green { color:var(--c-green); }
.legend-item.red   { color:var(--c-red); }
.legend-item.blue  { color:var(--c-accent); }
.host-avail-list { display:flex;flex-direction:column;gap:12px; }
.host-avail-item { }
.avail-bar  { height:6px;background:var(--c-border);border-radius:3px;overflow:hidden; }
.avail-fill { height:100%;border-radius:3px;transition:width .4s; }

.health-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(220px,1fr));gap:10px;margin-top:10px; }
.health-card { background:var(--c-bg);border:1px solid var(--c-border);border-radius:8px;padding:12px; }
.health-ok   { border-color:rgba(34,197,94,.25); }
.health-err  { border-color:rgba(239,68,68,.35);background:rgba(239,68,68,.04); }
.health-warn { border-color:rgba(245,158,11,.25); }
.health-header { display:flex;align-items:center;gap:6px;margin-bottom:4px; }
.health-name { font-size:12px;font-weight:600;color:var(--c-text); }
.health-meta { display:flex;justify-content:space-between;margin-bottom:6px; }
.health-stats { display:flex;align-items:center;gap:8px; }
@media (max-width:900px) {
  .stat-grid { grid-template-columns:repeat(2,1fr); }
  .dash-row  { flex-direction:column; }
}
</style>
