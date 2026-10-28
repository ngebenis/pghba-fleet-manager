<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-logo">
        <div class="logo-icon">PG</div>
        <div class="login-title">PGHBA Manager</div>
        <div class="login-sub">PostgreSQL pg_hba.conf Management</div>
      </div>

      <form class="login-form" @submit.prevent="handleLogin">
        <div class="field">
          <label>Username</label>
          <input v-model="form.username" type="text" placeholder="admin" autocomplete="username" required>
        </div>
        <div class="field">
          <label>Password</label>
          <input v-model="form.password" type="password" placeholder="••••••••" autocomplete="current-password" required>
        </div>

        <div v-if="error" class="error-msg">
          <i class="ti ti-alert-circle"></i> {{ error }}
        </div>

        <button class="btn btn-primary" type="submit" :disabled="loading" style="width:100%;justify-content:center;padding:10px">
          <i v-if="loading" class="ti ti-loader-2 spin"></i>
          <i v-else class="ti ti-login"></i>
          {{ loading ? 'Signing in...' : 'Sign In' }}
        </button>
      </form>

      <div class="login-footer">
        <span class="badge badge-green">PostgreSQL 10 – 18+</span>
        <span class="badge badge-blue">Docker Ready</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth   = useAuthStore()
const router = useRouter()

const form    = ref({ username: '', password: '' })
const loading = ref(false)
const error   = ref('')

async function handleLogin() {
  error.value   = ''
  loading.value = true
  try {
    await auth.login(form.value.username, form.value.password)
    router.push(auth.user?.must_change_password ? '/change-password' : '/dashboard')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Login failed. Check credentials.'
  } finally {
    loading.value = false
  }
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
  width: 380px;
}
.login-logo {
  text-align: center;
  margin-bottom: 28px;
}
.logo-icon {
  width: 48px; height: 48px; border-radius: 12px;
  background: var(--c-accent);
  display: inline-flex; align-items: center; justify-content: center;
  font-size: 18px; font-weight: 600; color: #fff;
  margin-bottom: 12px;
}
.login-title { font-size: 18px; font-weight: 600; }
.login-sub   { font-size: 12px; color: var(--c-muted); margin-top: 4px; }
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
  display: flex; gap: 8px; justify-content: center; margin-top: 20px;
}
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
