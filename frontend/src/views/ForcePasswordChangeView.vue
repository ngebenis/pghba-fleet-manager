<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-logo">
        <div class="logo-icon"><i class="ti ti-shield-lock"></i></div>
        <div class="login-title">Password Change Required</div>
        <div class="login-sub">This account still uses the published default password — set a new one to continue.</div>
      </div>

      <form class="login-form" @submit.prevent="handleSubmit">
        <div class="field">
          <label>Current Password</label>
          <input v-model="form.current" type="password" placeholder="••••••••" autocomplete="current-password" required>
        </div>
        <div class="field">
          <label>New Password</label>
          <input v-model="form.next" type="password" placeholder="At least 8 characters" autocomplete="new-password" minlength="8" required>
        </div>
        <div class="field">
          <label>Confirm New Password</label>
          <input v-model="form.confirm" type="password" placeholder="Repeat new password" autocomplete="new-password" minlength="8" required>
        </div>

        <div v-if="error" class="error-msg">
          <i class="ti ti-alert-circle"></i> {{ error }}
        </div>

        <button class="btn btn-primary" type="submit" :disabled="loading" style="width:100%;justify-content:center;padding:10px">
          <i v-if="loading" class="ti ti-loader-2 spin"></i>
          <i v-else class="ti ti-check"></i>
          {{ loading ? 'Saving...' : 'Set New Password' }}
        </button>
      </form>

      <div class="login-footer">
        <button class="btn btn-ghost btn-sm" @click="handleLogout" :disabled="loading">
          <i class="ti ti-logout"></i> Log out instead
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api'

const auth   = useAuthStore()
const router = useRouter()

const form    = ref({ current: '', next: '', confirm: '' })
const loading = ref(false)
const error   = ref('')

async function handleSubmit() {
  error.value = ''
  if (form.value.next.length < 8) {
    error.value = 'New password must be at least 8 characters'; return
  }
  if (form.value.next !== form.value.confirm) {
    error.value = 'New password and confirmation do not match'; return
  }
  if (form.value.next === form.value.current) {
    error.value = 'New password must be different from the current password'; return
  }
  loading.value = true
  try {
    const { data } = await authApi.changePassword(form.value.current, form.value.next)
    auth.applyPasswordChanged(data.access_token)
    router.push('/dashboard')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to change password'
  } finally {
    loading.value = false
  }
}

async function handleLogout() {
  await auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--c-bg);
}
.login-card {
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: 14px;
  padding: 36px;
  width: 400px;
}
.login-logo {
  text-align: center;
  margin-bottom: 28px;
}
.logo-icon {
  width: 48px; height: 48px; border-radius: 12px;
  background: var(--c-amber, #f59e0b);
  display: inline-flex; align-items: center; justify-content: center;
  font-size: 20px; color: #fff;
  margin-bottom: 12px;
}
.login-title { font-size: 16px; font-weight: 600; }
.login-sub   { font-size: 12px; color: var(--c-muted); margin-top: 6px; line-height: 1.5; }
.login-form  { display: flex; flex-direction: column; gap: 0; }
.error-msg {
  background: rgba(239,68,68,.1);
  border: 1px solid rgba(239,68,68,.25);
  color: var(--c-red);
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 12px;
  margin-bottom: 12px;
}
.login-footer {
  display: flex; gap: 8px; justify-content: center; margin-top: 16px;
}
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
