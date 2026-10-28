<template>
  <div :class="['app-layout', theme.current]">
    <div v-if="sidebarOpen" class="sidebar-overlay" @click="sidebarOpen = false"></div>
    <nav :class="['sidebar', { 'sidebar-open': sidebarOpen }]">
      <div class="brand">
        <div class="brand-icon">PG</div>
        <div class="brand-text">
          <div class="brand-name">PGHBA Manager</div>
          <div class="brand-sub">v1.0.0</div>
        </div>
      </div>
      <div class="nav-section-label">{{ $t("nav.overview").toUpperCase() }}</div>
      <router-link to="/dashboard" class="nav-item"><i class="ti ti-layout-dashboard"></i> {{ $t("nav.dashboard") }}</router-link>
      <router-link v-if="!auth.hidesAdminMenus" to="/audit" class="nav-item"><i class="ti ti-history"></i> {{ $t("nav.auditLog") }}</router-link>
      <template v-if="!auth.hidesAdminMenus">
        <div class="nav-section-label">{{ $t("nav.infrastructure").toUpperCase() }}</div>
        <router-link to="/hosts" class="nav-item"><i class="ti ti-server"></i> {{ $t("nav.databaseHosts") }}</router-link>
        <router-link to="/instances" class="nav-item"><i class="ti ti-database"></i> {{ $t("nav.pgInstances") }}</router-link>
      </template>
      <div class="nav-section-label">{{ $t("nav.accessControl").toUpperCase() }}</div>
      <router-link v-if="auth.canSeeAccessRules" to="/rules" class="nav-item"><i class="ti ti-shield"></i> {{ $t("nav.accessRules") }}</router-link>
      <router-link to="/temp-access" class="nav-item"><i class="ti ti-clock"></i> {{ $t("nav.tempAccess") }}</router-link>
      <template v-if="!auth.hidesAdminMenus">
        <div class="nav-section-label">{{ $t("nav.system").toUpperCase() }}</div>
        <router-link to="/backup" class="nav-item"><i class="ti ti-device-floppy"></i> {{ $t("nav.backup") }}</router-link>
        <router-link to="/users" class="nav-item"><i class="ti ti-users"></i> {{ $t("nav.userMgmt") }}</router-link>
        <router-link to="/settings" class="nav-item"><i class="ti ti-settings"></i> {{ $t("nav.settings") }}</router-link>
      </template>
      <div class="sidebar-footer">
        <div class="user-info">
          <div class="user-avatar">{{ auth.user?.username?.[0]?.toUpperCase() }}</div>
          <div>
            <div class="user-name">{{ auth.user?.username }}</div>
            <div class="user-role">{{ auth.user?.role }}</div>
          </div>
        </div>
        <button class="btn-logout" @click="logout"><i class="ti ti-logout"></i></button>
      </div>
    </nav>
    <div class="main-content">
      <div class="topbar">
        <button class="sidebar-toggle" @click="sidebarOpen = !sidebarOpen">
          <i :class="['ti', sidebarOpen ? 'ti-x' : 'ti-menu-2']"></i>
        </button>
        <span class="page-title">{{ pageTitle }}</span>
        <div class="topbar-search">
          <i class="ti ti-search"></i>
          <input v-model="searchQuery" :placeholder="$t('common.search')" />
        </div>
        <div class="topbar-right">
          <button :class="['lang-btn', locale === 'en' ? 'active' : '']" @click="setLocale('en')">&#127468;&#127463; EN</button>
          <button :class="['lang-btn', locale === 'id' ? 'active' : '']" @click="setLocale('id')">&#127470;&#127465; ID</button>
          <NotificationBell />
          <button class="theme-btn" @click="theme.toggle()">
            <i :class="['ti', theme.current === 'dark' ? 'ti-sun' : 'ti-moon']"></i>
            {{ theme.current === 'dark' ? 'Light' : 'Dark' }}
          </button>
        </div>
      </div>
      <div class="page-scroll">
        <router-view />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import { globalSearch } from '@/stores/search'
import { useI18n } from '@/i18n'
import NotificationBell from '@/components/NotificationBell.vue'

const auth    = useAuthStore()
const theme   = useThemeStore()
const route   = useRoute()
const router  = useRouter()
const { $t, locale, setLocale } = useI18n()

const sidebarOpen = ref(false)
const searchQuery = ref('')

const currentLang = computed(() => locale.value)

watch(() => route.path, () => { sidebarOpen.value = false })
watch(searchQuery, (v) => { globalSearch.value = v })





const pageTitle = computed(() => {
  const map: Record<string, string> = {
    '/': 'Dashboard', '/audit': 'Audit Log',
    '/hosts': 'Database Hosts', '/instances': 'PG Instances',
    '/rules': 'Access Rules (pg_hba.conf)', '/temp-access': 'Temporary Access',
    '/backup': 'Backup & Restore', '/users': 'User Management', '/settings': 'Settings',
  }
  return map[route.path] || 'PGHBA Manager'
})

async function logout() {
  await auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.app-layout { display:grid;grid-template-columns:200px 1fr;height:100vh;background:var(--c-bg);color:var(--c-text);font-family:var(--font); }
.sidebar { display:flex;flex-direction:column;background:var(--c-surface);border-right:1px solid var(--c-border);height:100vh;overflow-y:auto;position:relative;z-index:10; }
.brand { display:flex;align-items:center;gap:10px;padding:16px 14px;border-bottom:1px solid var(--c-border); }
.brand-icon { width:32px;height:32px;background:var(--c-accent);border-radius:6px;display:flex;align-items:center;justify-content:center;font-weight:700;font-size:12px;color:#fff; }
.brand-name { font-size:13px;font-weight:700;color:var(--c-text); }
.brand-sub { font-size:10px;color:var(--c-muted); }
.nav-section-label { font-size:9px;font-weight:600;color:var(--c-muted);text-transform:uppercase;letter-spacing:.8px;padding:12px 14px 4px; }
.nav-item { display:flex;align-items:center;gap:8px;padding:8px 14px;font-size:12px;color:var(--c-muted);text-decoration:none;border-radius:6px;margin:1px 6px;transition:all .15s; }
.nav-item:hover,.nav-item.router-link-active { background:var(--c-surface2);color:var(--c-text); }
.nav-item.router-link-exact-active { background:rgba(79,142,247,.12);color:var(--c-accent); }
.sidebar-footer { margin-top:auto;padding:12px;border-top:1px solid var(--c-border);display:flex;align-items:center;gap:8px; }
.user-info { display:flex;align-items:center;gap:8px;flex:1;min-width:0; }
.user-avatar { width:28px;height:28px;background:var(--c-accent);border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:11px;font-weight:700;color:#fff;flex-shrink:0; }
.user-name { font-size:12px;font-weight:500;color:var(--c-text); }
.user-role { font-size:10px;color:var(--c-muted); }
.btn-logout { background:none;border:none;color:var(--c-muted);cursor:pointer;padding:6px;border-radius:4px; }
.btn-logout:hover { color:var(--c-red); }
.main-content { display:flex;flex-direction:column;overflow:hidden;min-height:0; }
.topbar { display:flex;align-items:center;gap:10px;padding:10px 16px;border-bottom:1px solid var(--c-border);background:var(--c-surface);flex-shrink:0; }
.page-title { font-size:14px;font-weight:600;white-space:nowrap;margin-right:8px; }
.topbar-search { flex:1;max-width:320px;display:flex;align-items:center;gap:8px;background:var(--c-bg);border:1px solid var(--c-border);border-radius:6px;padding:6px 10px; }
.topbar-search input { background:none;border:none;outline:none;color:var(--c-text);font-size:12px;width:100%; }
.topbar-right { display:flex;align-items:center;gap:6px;margin-left:auto; }
.lang-btn { background:var(--c-surface2);border:1px solid var(--c-border);color:var(--c-muted);padding:4px 8px;border-radius:4px;cursor:pointer;font-size:11px; }
.lang-btn.active { background:#ffffff;color:#1a1a2e;border-color:#ffffff;font-weight:700; }
.theme-btn { display:flex;align-items:center;gap:4px;background:var(--c-surface2);border:1px solid var(--c-border);color:var(--c-muted);padding:5px 10px;border-radius:5px;cursor:pointer;font-size:11px; }
.page-scroll { flex:1;overflow-y:auto;overflow-x:hidden; }
.sidebar-toggle { display:none;background:none;border:none;color:var(--c-text);font-size:20px;cursor:pointer;padding:4px 8px;flex-shrink:0; }
.sidebar-overlay { display:none;position:fixed;inset:0;background:rgba(0,0,0,.5);z-index:99; }
@media (max-width:1024px) {
  .app-layout { grid-template-columns:1fr; }
  .sidebar { position:fixed;left:-210px;top:0;bottom:0;z-index:100;width:200px;transition:left .25s ease;box-shadow:4px 0 20px rgba(0,0,0,.3); }
  .sidebar.sidebar-open { left:0; }
  .sidebar-toggle { display:flex;align-items:center; }
  .sidebar-overlay { display:block; }
  .topbar-search { display:none; }
  .page-title { font-size:13px; }
  .table-wrap { overflow-x:auto; }
  .modal { width:100% !important;max-width:100% !important;border-radius:16px 16px 0 0 !important;position:fixed !important;bottom:0 !important;top:auto !important;max-height:90vh !important; }
  .modal-backdrop { align-items:flex-end !important; }
}
</style>
