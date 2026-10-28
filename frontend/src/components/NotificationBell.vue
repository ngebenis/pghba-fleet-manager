<template>
  <div class="notif-wrap" ref="wrapRef">
    <!-- Bell Button -->
    <button class="notif-btn" @click="toggle" :class="{ active: open }">
      <i class="ti ti-bell"></i>
      <span v-if="warningCount > 0" class="notif-badge" :class="badgeClass">
        {{ warningCount > 9 ? '9+' : warningCount }}
      </span>
    </button>

    <!-- Dropdown Panel -->
    <Teleport to="body">
      <div v-if="open" class="notif-panel" :style="panelStyle" @click.stop>
        <div class="notif-header">
          <span class="notif-title">
            <i class="ti ti-bell"></i> Notifications
            <span v-if="total > 0" class="notif-count-badge">{{ total }}</span>
          </span>
          <div style="display:flex;gap:4px">
            <button class="btn btn-ghost btn-sm" @click="load" title="Refresh">
              <i class="ti ti-refresh" :class="{ spin: loading }"></i>
            </button>
            <button class="btn btn-ghost btn-sm" @click="open = false">
              <i class="ti ti-x"></i>
            </button>
          </div>
        </div>

        <div v-if="loading && !notifications.length" class="notif-empty">
          <i class="ti ti-loader-2 spin" style="font-size:20px"></i>
        </div>

        <div v-else-if="!notifications.length" class="notif-empty">
          <i class="ti ti-bell-off" style="font-size:28px;color:var(--c-muted)"></i>
          <div>No active temporary access</div>
        </div>

        <div v-else class="notif-list">
          <div v-for="n in notifications" :key="n.id"
            :class="['notif-item', `notif-${n.level}`]">
            <div class="notif-item-icon">
              <i :class="['ti', levelIcon(n.level)]"></i>
            </div>
            <div class="notif-item-body">
              <div class="notif-item-title">{{ n.title }}</div>
              <div class="notif-item-sub">
                <span class="mono">{{ n.ip_address }}</span>
                <span class="sep">→</span>
                <span>{{ n.instance_label }}</span>
              </div>
              <div class="notif-item-meta">
                <span v-if="n.host_name" class="meta-chip">
                  <i class="ti ti-server" style="font-size:9px"></i>
                  {{ n.host_name }}
                </span>
                <span class="meta-chip">
                  <i class="ti ti-user" style="font-size:9px"></i>
                  {{ n.granted_by }}
                </span>
                <span v-if="n.reason" class="meta-chip" style="max-width:120px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" :title="n.reason">
                  {{ n.reason }}
                </span>
              </div>
              <div class="notif-item-time">
                <i class="ti ti-clock" style="font-size:10px"></i>
                <span v-if="n.minutes_left > 0">
                  Expires in
                  <strong :style="{ color: timeColor(n.minutes_left) }">
                    {{ formatRemaining(n.minutes_left) }}
                  </strong>
                  ({{ formatDate(n.expires_at) }})
                </span>
                <span v-else style="color:var(--c-red)">
                  Expired {{ formatDate(n.expires_at) }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div class="notif-footer">
          <router-link to="/temp-access" @click="open = false"
            style="font-size:11px;color:var(--c-accent);text-decoration:none">
            View all temporary access →
          </router-link>
          <span style="font-size:10px;color:var(--c-muted)">
            Auto-refresh {{ countdown }}s
          </span>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import api from '@/api'
import dayjs from 'dayjs'

const open         = ref(false)
const loading      = ref(false)
const notifications = ref<any[]>([])
const total        = ref(0)
const warningCount = ref(0)
const wrapRef      = ref<HTMLElement>()
const panelStyle   = ref({})
const countdown    = ref(30)

let refreshTimer: ReturnType<typeof setInterval>
let countdownTimer: ReturnType<typeof setInterval>

const badgeClass = computed(() => {
  if (notifications.value.some(n => n.level === 'danger')) return 'badge-danger'
  if (notifications.value.some(n => n.level === 'warning')) return 'badge-warning'
  return 'badge-info'
})

function levelIcon(level: string) {
  if (level === 'danger')  return 'ti-alert-triangle'
  if (level === 'warning') return 'ti-clock-exclamation'
  return 'ti-clock-check'
}

function timeColor(minutes: number) {
  if (minutes <= 5)  return 'var(--c-red)'
  if (minutes <= 15) return 'var(--c-amber)'
  return 'var(--c-green)'
}

function formatRemaining(minutes: number) {
  if (minutes <= 0)  return 'Expired'
  if (minutes < 60)  return `${minutes}m`
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  return m > 0 ? `${h}h ${m}m` : `${h}h`
}

function formatDate(d: string) {
  return dayjs(d).format('DD MMM HH:mm')
}

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/notifications')
    notifications.value = data.notifications || []
    total.value         = data.total || 0
    warningCount.value  = data.warnings || 0
    countdown.value     = 30
  } catch (e) {
    console.error('Failed to load notifications', e)
  } finally {
    loading.value = false
  }
}

function toggle() {
  open.value = !open.value
  if (open.value) {
    load()
    // Position panel
    setTimeout(() => {
      const btn = wrapRef.value?.getBoundingClientRect()
      if (btn) {
        panelStyle.value = {
          position: 'fixed',
          top: (btn.bottom + 8) + 'px',
          right: (window.innerWidth - btn.right) + 'px',
          zIndex: 9999,
        }
      }
    }, 10)
  }
}

function onClickOutside(e: MouseEvent) {
  if (open.value && wrapRef.value && !wrapRef.value.contains(e.target as Node)) {
    open.value = false
  }
}

onMounted(() => {
  load()
  // Auto-refresh every 30 seconds
  refreshTimer = setInterval(load, 30000)
  // Countdown display
  countdownTimer = setInterval(() => {
    countdown.value = Math.max(0, countdown.value - 1)
  }, 1000)
  document.addEventListener('click', onClickOutside)
})

onUnmounted(() => {
  clearInterval(refreshTimer)
  clearInterval(countdownTimer)
  document.removeEventListener('click', onClickOutside)
})
</script>

<style scoped>
.notif-wrap { position:relative; }

.notif-btn {
  position:relative;
  background:none;
  border:1px solid var(--c-border);
  border-radius:8px;
  padding:6px 10px;
  color:var(--c-muted);
  cursor:pointer;
  font-size:16px;
  display:flex;
  align-items:center;
  transition:all .15s;
}
.notif-btn:hover, .notif-btn.active {
  background:var(--c-surface2);
  color:var(--c-text);
  border-color:var(--c-accent);
}

.notif-badge {
  position:absolute;
  top:-5px;
  right:-5px;
  min-width:16px;
  height:16px;
  border-radius:8px;
  font-size:9px;
  font-weight:700;
  display:flex;
  align-items:center;
  justify-content:center;
  padding:0 4px;
  border:2px solid var(--c-bg);
}
.badge-danger  { background:var(--c-red);color:#fff; }
.badge-warning { background:var(--c-amber);color:#000; }
.badge-info    { background:var(--c-accent);color:#fff; }

.notif-panel {
  width:360px;
  background:var(--c-surface);
  border:1px solid var(--c-border);
  border-radius:12px;
  box-shadow:0 8px 32px rgba(0,0,0,.3);
  overflow:hidden;
  max-height:520px;
  display:flex;
  flex-direction:column;
}

.notif-header {
  display:flex;
  align-items:center;
  justify-content:space-between;
  padding:12px 14px;
  border-bottom:1px solid var(--c-border);
  background:var(--c-surface);
}
.notif-title {
  font-size:13px;
  font-weight:600;
  color:var(--c-text);
  display:flex;
  align-items:center;
  gap:6px;
}
.notif-count-badge {
  background:var(--c-accent);
  color:#fff;
  font-size:10px;
  padding:1px 6px;
  border-radius:10px;
}

.notif-empty {
  padding:40px 20px;
  text-align:center;
  color:var(--c-muted);
  font-size:12px;
  display:flex;
  flex-direction:column;
  align-items:center;
  gap:8px;
}

.notif-list {
  overflow-y:auto;
  flex:1;
}

.notif-item {
  display:flex;
  gap:10px;
  padding:12px 14px;
  border-bottom:1px solid var(--c-border);
  transition:background .1s;
}
.notif-item:hover { background:var(--c-surface2); }
.notif-item:last-child { border-bottom:none; }

.notif-danger  { border-left:3px solid var(--c-red); }
.notif-warning { border-left:3px solid var(--c-amber); }
.notif-info    { border-left:3px solid var(--c-accent); }

.notif-item-icon {
  width:28px;
  height:28px;
  border-radius:8px;
  display:flex;
  align-items:center;
  justify-content:center;
  flex-shrink:0;
  font-size:14px;
}
.notif-danger  .notif-item-icon { background:rgba(239,68,68,.15);color:var(--c-red); }
.notif-warning .notif-item-icon { background:rgba(245,158,11,.15);color:var(--c-amber); }
.notif-info    .notif-item-icon { background:rgba(79,142,247,.15);color:var(--c-accent); }

.notif-item-body { flex:1;min-width:0; }
.notif-item-title { font-size:12px;font-weight:600;color:var(--c-text);margin-bottom:3px; }
.notif-item-sub {
  font-size:11px;
  color:var(--c-muted);
  display:flex;
  align-items:center;
  gap:4px;
  margin-bottom:4px;
}
.notif-item-sub .mono { font-family:monospace;color:var(--c-text); }
.sep { color:var(--c-border); }

.notif-item-meta {
  display:flex;
  gap:4px;
  flex-wrap:wrap;
  margin-bottom:4px;
}
.meta-chip {
  display:inline-flex;
  align-items:center;
  gap:3px;
  background:var(--c-bg);
  border:1px solid var(--c-border);
  border-radius:4px;
  padding:1px 6px;
  font-size:10px;
  color:var(--c-muted);
}

.notif-item-time {
  font-size:10px;
  color:var(--c-muted);
  display:flex;
  align-items:center;
  gap:4px;
}

.notif-footer {
  padding:10px 14px;
  border-top:1px solid var(--c-border);
  display:flex;
  align-items:center;
  justify-content:space-between;
  background:var(--c-surface);
}

.spin { animation:spin 1s linear infinite; }
@keyframes spin { to { transform:rotate(360deg); } }
</style>
