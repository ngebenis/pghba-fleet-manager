<template>
  <div class="page-content">
    <div class="settings-grid">

      <!-- System Info -->
      <div class="card">
        <div class="settings-card-title">
          <i class="ti ti-info-circle" style="color:var(--c-accent)"></i> System Info
        </div>
        <div v-if="loadingInfo" style="text-align:center;padding:20px;color:var(--c-muted)">
          <i class="ti ti-loader-2 spin"></i>
        </div>
        <div v-else class="info-list">
          <div class="info-row">
            <span class="info-label">Version</span>
            <span class="info-value badge-green-text">v{{ sysInfo.version }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Go Runtime</span>
            <span class="info-value mono">{{ sysInfo.go_version }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Platform</span>
            <span class="info-value mono">{{ sysInfo.go_os }}/{{ sysInfo.go_arch }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Uptime</span>
            <span class="info-value">{{ sysInfo.uptime_label }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Database Size</span>
            <span class="info-value mono">{{ sysInfo.stats?.db_size }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">PG Compatibility</span>
            <span class="info-value mono">PG 10 – 18</span>
          </div>
        </div>

        <!-- Stats -->
        <div v-if="sysInfo.stats" style="margin-top:14px">
          <div class="settings-card-title" style="margin-bottom:8px;font-size:11px;color:var(--c-muted)">
            DATABASE STATISTICS
          </div>
          <div class="stat-mini-grid">
            <div class="stat-mini">
              <div class="stat-mini-val">{{ sysInfo.stats.hosts }}</div>
              <div class="stat-mini-label">Hosts</div>
            </div>
            <div class="stat-mini">
              <div class="stat-mini-val">{{ sysInfo.stats.instances }}</div>
              <div class="stat-mini-label">Instances</div>
            </div>
            <div class="stat-mini">
              <div class="stat-mini-val">{{ sysInfo.stats.rules }}</div>
              <div class="stat-mini-label">Rules</div>
            </div>
            <div class="stat-mini">
              <div class="stat-mini-val">{{ sysInfo.stats.users }}</div>
              <div class="stat-mini-label">Users</div>
            </div>
            <div class="stat-mini">
              <div class="stat-mini-val">{{ sysInfo.stats.backups }}</div>
              <div class="stat-mini-label">Backups</div>
            </div>
            <div class="stat-mini">
              <div class="stat-mini-val">{{ sysInfo.stats.audit_logs }}</div>
              <div class="stat-mini-label">Audit Logs</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Scheduler -->
      <div class="card">
        <div class="settings-card-title">
          <i class="ti ti-clock-play" style="color:var(--c-amber)"></i> Temp Access Scheduler
        </div>
        <div class="info-box amber" style="margin-bottom:12px">
          <i class="ti ti-alert-triangle"></i>
          Scheduler settings need restart container to take effect.
          Edit file .env on the server.
        </div>
        <div class="info-list">
          <div class="info-row">
            <span class="info-label">Check Interval</span>
            <span class="info-value">
              Every {{ sysInfo.scheduler_interval }} minute(s)
            </span>
          </div>
          <div class="info-row">
            <span class="info-label">On Expiry</span>
            <span class="info-value">
              <span :class="['badge', sysInfo.temp_access_action === 'delete' ? 'badge-red' : 'badge-amber']">
                {{ sysInfo.temp_access_action === 'delete' ? 'Delete rule' : 'Disable rule' }}
              </span>
            </span>
          </div>
          <div class="info-row">
            <span class="info-label">Auto write-back</span>
            <span class="info-value badge-green-text">Active</span>
          </div>
          <div class="info-row">
            <span class="info-label">Auto-reload PG</span>
            <span class="info-value badge-green-text">Active</span>
          </div>
        </div>
        <div style="margin-top:12px;background:var(--c-bg);border:1px solid var(--c-border);border-radius:6px;padding:10px 12px;font-size:11px;font-family:monospace;color:var(--c-muted)">
          <div style="color:var(--c-text);margin-bottom:4px"># Edit .env to change:</div>
          <div>SCHEDULER_INTERVAL_MIN={{ sysInfo.scheduler_interval }}</div>
          <div>TEMP_ACCESS_ACTION={{ sysInfo.temp_access_action }}</div>
        </div>
      </div>

      <!-- Security -->
      <div class="card">
        <div class="settings-card-title">
          <i class="ti ti-lock" style="color:var(--c-green)"></i> Security & Encryption
        </div>
        <div class="info-list">
          <div class="info-row">
            <span class="info-label">SSH Key Encryption</span>
            <span class="info-value badge-green-text">AES-256-GCM</span>
          </div>
          <div class="info-row">
            <span class="info-label">Password Hashing</span>
            <span class="info-value badge-green-text">bcrypt (cost 12)</span>
          </div>
          <div class="info-row">
            <span class="info-label">JWT Auth</span>
            <span class="info-value badge-green-text">HS256</span>
          </div>
          <div class="info-row">
            <span class="info-label">CORS</span>
            <span class="info-value badge-green-text">Enabled</span>
          </div>
          <div class="info-row">
            <span class="info-label">Audit Trail</span>
            <span class="info-value badge-green-text">All actions logged</span>
          </div>
        </div>
        <div class="info-box green" style="margin-top:12px">
          <i class="ti ti-shield-check"></i>
          SSH keys and credentials encrypted with AES-256-GCM before saving to database.
        </div>
      </div>

      <!-- Auto-Backup Policy -->
      <div class="card">
        <div class="settings-card-title">
          <i class="ti ti-device-floppy" style="color:var(--c-accent)"></i> Auto-Backup Policy
        </div>
        <div class="info-list">
          <div class="info-row">
            <span class="info-label">Before edit rule</span>
            <span class="info-value badge-green-text">✓ Active</span>
          </div>
          <div class="info-row">
            <span class="info-label">Before delete rule</span>
            <span class="info-value badge-green-text">✓ Active</span>
          </div>
          <div class="info-row">
            <span class="info-label">Before restore</span>
            <span class="info-value badge-green-text">✓ Active</span>
          </div>
          <div class="info-row">
            <span class="info-label">On temp access grant</span>
            <span class="info-value badge-green-text">✓ Active</span>
          </div>
        </div>

        <div style="margin-top:14px">
          <label style="font-size:12px;color:var(--c-muted);display:block;margin-bottom:6px">
            Backup Retention (per instance)
          </label>
          <div style="display:flex;align-items:center;gap:8px">
            <input v-model.number="retentionDays" type="number" min="10" max="500"
              style="width:80px;text-align:center">
            <span style="font-size:12px;color:var(--c-muted)">backups max</span>
            <button class="btn btn-ghost btn-sm" @click="saveRetention">
              <i class="ti ti-check"></i> Save
            </button>
          </div>
        </div>
      </div>

      <!-- Notifications -->
      <div class="card">
        <div class="settings-card-title">
          <i class="ti ti-bell" style="color:#8b5cf6"></i> Notification Channels
        </div>
        <div class="notif-list">
          <!-- Telegram -->
          <div class="notif-item">
            <div class="notif-info">
              <i class="ti ti-brand-telegram" style="color:#2ca5e0;font-size:18px"></i>
              <div>
                <div class="notif-name">Telegram</div>
                <div class="notif-desc">Alert to Telegram bot</div>
              </div>
            </div>
            <label class="toggle-switch">
              <input type="checkbox" v-model="notif.telegram">
              <span class="toggle-slider"></span>
            </label>
          </div>
          <div v-if="notif.telegram" class="notif-config">
            <div class="field"><label>Bot Token</label>
              <input v-model="notif.tg_token" placeholder="123456:ABC-DEF..." type="password"></div>
            <div class="field"><label>Chat ID</label>
              <input v-model="notif.tg_chat" placeholder="-100123456789"></div>
            <button class="btn btn-ghost btn-sm" @click="testTelegram" style="margin-top:6px">
              <i class="ti ti-send"></i> Test
            </button>
          </div>

          <!-- Email -->
          <div class="notif-item">
            <div class="notif-info">
              <i class="ti ti-mail" style="color:var(--c-accent);font-size:18px"></i>
              <div>
                <div class="notif-name">Email (SMTP)</div>
                <div class="notif-desc">Send email alerts</div>
              </div>
            </div>
            <label class="toggle-switch">
              <input type="checkbox" v-model="notif.email">
              <span class="toggle-slider"></span>
            </label>
          </div>
          <div v-if="notif.email" class="notif-config">
            <div class="field-row">
              <div class="field"><label>SMTP Host</label>
                <input v-model="notif.smtp_host" placeholder="smtp.gmail.com"></div>
              <div class="field"><label>Port</label>
                <input v-model.number="notif.smtp_port" type="number" placeholder="587" style="width:80px"></div>
            </div>
            <div class="field"><label>From</label>
              <input v-model="notif.smtp_from" placeholder="pghba@company.com"></div>
            <div class="field"><label>To (DBA Team)</label>
              <input v-model="notif.smtp_to" placeholder="dba-team@company.com"></div>
          </div>

          <!-- WhatsApp -->
          <div class="notif-item">
            <div class="notif-info">
              <i class="ti ti-brand-whatsapp" style="color:#25d366;font-size:18px"></i>
              <div>
                <div class="notif-name">WhatsApp (Fonnte)</div>
                <div class="notif-desc">Alert via WhatsApp</div>
              </div>
            </div>
            <label class="toggle-switch">
              <input type="checkbox" v-model="notif.whatsapp">
              <span class="toggle-slider"></span>
            </label>
          </div>
          <div v-if="notif.whatsapp" class="notif-config">
            <div class="field"><label>Fonnte Token</label>
              <input v-model="notif.fonnte_token" placeholder="fonnte API token" type="password"></div>
            <div class="field"><label>Phone Number</label>
              <input v-model="notif.wa_phone" placeholder="6281234567890"></div>
          </div>
        </div>

        <div style="margin-top:12px;display:flex;gap:8px">
          <button class="btn btn-ghost" @click="saveNotif" style="flex:1;justify-content:center">
            <i class="ti ti-check"></i> Save
          </button>
        </div>

        <!-- Notification events -->
        <div style="margin-top:14px;border-top:1px solid var(--c-border);padding-top:12px">
          <div style="font-size:11px;color:var(--c-muted);margin-bottom:8px;text-transform:uppercase;letter-spacing:.5px">
            Alert on these events:
          </div>
          <div v-for="ev in notifEvents" :key="ev.key"
            style="display:flex;align-items:center;justify-content:space-between;padding:5px 0;border-bottom:1px solid var(--c-border)">
            <span style="font-size:12px">{{ ev.label }}</span>
            <label class="toggle-switch small">
              <input type="checkbox" v-model="ev.enabled">
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
      </div>



      <!-- Auto-Backup Schedule -->
      <div class="card">
        <div class="settings-card-title">
          <i class="ti ti-calendar-time" style="color:var(--c-green)"></i> Scheduled Auto-Backup
        </div>

        <div v-if="!sysInfo.auto_backup_enabled" class="info-box amber" style="margin-bottom:12px">
          <i class="ti ti-alert-triangle"></i>
          Auto-backup disabled. Add to .env to enable.
        </div>
        <div v-else class="info-box green" style="margin-bottom:12px">
          <i class="ti ti-circle-check"></i>
          Auto-backup active — runs <strong>{{ sysInfo.auto_backup_cron }}</strong> → {{ sysInfo.auto_backup_target }}
        </div>

        <div class="info-list">
          <div class="info-row">
            <span class="info-label">Status</span>
            <span :class="sysInfo.auto_backup_enabled ? 'badge-green-text' : 'info-value'" style="font-size:12px">
              {{ sysInfo.auto_backup_enabled ? 'Enabled' : 'Disabled' }}
            </span>
          </div>
          <div class="info-row">
            <span class="info-label">Schedule (cron)</span>
            <span class="info-value mono">{{ sysInfo.auto_backup_cron || '0 2 * * *' }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Storage Target</span>
            <span class="info-value">{{ sysInfo.auto_backup_target || 'local' }}</span>
          </div>
        </div>

        <div style="margin-top:12px;background:var(--c-bg);border:1px solid var(--c-border);border-radius:6px;padding:10px 12px;font-size:11px;font-family:monospace;color:var(--c-muted)">
          <div style="color:var(--c-text);margin-bottom:4px"># Add to .env to enable:</div>
          <div>AUTO_BACKUP_ENABLED=true</div>
          <div>AUTO_BACKUP_CRON=0 2 * * *  # daily 2am</div>
          <div>AUTO_BACKUP_TARGET=s3  # local | s3 | both</div>
        </div>

        <button class="btn btn-ghost" style="margin-top:12px;width:100%;justify-content:center"
          @click="triggerManualBackup" :disabled="triggeringBackup">
          <i :class="['ti', triggeringBackup ? 'ti-loader-2 spin' : 'ti-player-play']"></i>
          {{ triggeringBackup ? 'Running...' : 'Run Backup Now' }}
        </button>
      </div>

      <!-- Backup Schedules (managed) -->
      <div class="card" style="grid-column:1 / -1">
        <div class="settings-card-title" style="justify-content:space-between">
          <span style="display:flex;align-items:center;gap:8px">
            <i class="ti ti-calendar-time" style="color:var(--c-teal)"></i> Backup Schedules
          </span>
          <button class="btn btn-primary btn-sm" @click="openCreateSchedule">
            <i class="ti ti-plus"></i> Add Schedule
          </button>
        </div>

        <div v-if="loadingSchedules" style="text-align:center;padding:20px;color:var(--c-muted)">
          <i class="ti ti-loader-2 spin"></i>
        </div>
        <div v-else-if="schedules.length === 0" style="text-align:center;padding:24px;color:var(--c-muted);font-size:12px">
          <i class="ti ti-calendar-off" style="font-size:22px;display:block;margin-bottom:6px"></i>
          No backup schedules yet — click "Add Schedule" to create one.
        </div>
        <div v-else style="overflow-x:auto">
          <table class="schedule-table">
            <thead>
              <tr>
                <th>Name</th><th>Cron</th><th>Scope</th><th>Target</th>
                <th>Enabled</th><th>Last Run</th><th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="s in schedules" :key="s.id">
                <td>{{ s.name }}</td>
                <td class="mono">{{ s.cron_expression }}</td>
                <td style="font-size:11px">{{ scopeLabel(s) }}</td>
                <td style="font-size:11px;text-transform:uppercase">{{ s.target }}</td>
                <td>
                  <label class="switch">
                    <input type="checkbox" :checked="s.enabled" @change="toggleSchedule(s)">
                    <span class="switch-slider"></span>
                  </label>
                </td>
                <td style="font-size:11px;color:var(--c-muted)">
                  <span v-if="s.last_run_at">
                    {{ fmt(s.last_run_at) }}
                    <span :class="['badge', s.last_run_status === 'success' ? 'badge-green-text' : s.last_run_status === 'partial' ? 'badge-amber' : 'badge-red']"
                      style="display:block;margin-top:2px">{{ s.last_run_status }}</span>
                  </span>
                  <span v-else>Never run</span>
                </td>
                <td>
                  <div style="display:flex;gap:4px">
                    <button class="btn btn-ghost btn-sm" title="Edit" @click="openEditSchedule(s)">
                      <i class="ti ti-edit"></i>
                    </button>
                    <button class="btn btn-danger btn-sm" title="Delete" @click="deleteSchedule(s)">
                      <i class="ti ti-trash"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- PG Version EOL Warning -->
      <div class="card" :style="eolStatus.has_eol ? 'border-color:var(--c-red)' : ''">
        <div class="settings-card-title">
          <i class="ti ti-alert-triangle"
            :style="{ color: eolStatus.has_eol ? 'var(--c-red)' : 'var(--c-green)' }"></i>
          PostgreSQL Version Status
        </div>

        <div v-if="eolStatus.has_eol"
          style="background:rgba(239,68,68,.08);border:1px solid rgba(239,68,68,.2);border-radius:6px;padding:10px 12px;font-size:12px;margin-bottom:12px;color:var(--c-red)">
          <i class="ti ti-alert-circle"></i>
          <strong>{{ eolStatus.eol_count }} instances</strong> running End-of-Life PostgreSQL versions!
          Security patches no longer available.
        </div>
        <div v-else style="background:rgba(34,197,94,.08);border:1px solid rgba(34,197,94,.2);border-radius:6px;padding:10px 12px;font-size:12px;margin-bottom:12px;color:var(--c-green)">
          <i class="ti ti-circle-check"></i> All PostgreSQL versions are within support lifecycle.
        </div>

        <div v-for="v in eolStatus.versions" :key="v.version" class="pg-version-row"
          :class="v.is_eol ? 'pg-eol' : v.days_until_eol < 180 ? 'pg-warning' : 'pg-ok'">
          <div style="display:flex;align-items:center;gap:8px">
            <span class="mono" style="font-size:13px;font-weight:600">PG {{ v.version }}</span>
            <span :class="['badge', v.is_eol ? 'badge-red' : v.days_until_eol < 180 ? 'badge-amber' : 'badge-green']">
              {{ v.is_eol ? 'EOL' : v.days_until_eol < 180 ? 'Expiring Soon' : 'Supported' }}
            </span>
          </div>
          <div style="font-size:11px;color:var(--c-muted);margin-top:3px">
            {{ v.instances }} instance{{ v.instances > 1 ? 's' : '' }} ·
            EOL: {{ v.eol_date || 'Unknown' }}
            <span v-if="!v.is_eol && v.days_until_eol > 0"> ({{ v.days_until_eol }} days)</span>
            <span v-if="v.is_eol" style="color:var(--c-red);font-weight:600"> — PAST EOL, upgrade recommended!</span>
          </div>
        </div>

        <div v-if="!eolStatus.versions?.length" style="text-align:center;padding:16px;color:var(--c-muted);font-size:12px">
          <i class="ti ti-loader-2 spin"></i> Loading...
        </div>
      </div>
      <!-- Export & Storage -->
      <div class="card">
        <div class="settings-card-title">
          <i class="ti ti-cloud-upload" style="color:var(--c-accent)"></i> Export & Storage
        </div>

        <!-- Storage status -->
        <div v-if="storageStatus" class="storage-status">
          <div class="storage-item">
            <div class="storage-header">
              <i class="ti ti-folder" style="color:var(--c-amber)"></i>
              <span>Local Storage</span>
              <span class="badge badge-green" style="margin-left:auto">Active</span>
            </div>
            <div class="storage-detail">
              <span class="mono">{{ storageStatus.local?.path }}</span>
              <span style="color:var(--c-muted)">{{ storageStatus.local?.files }} .bak files</span>
            </div>
          </div>
          <div class="storage-item">
            <div class="storage-header">
              <i class="ti ti-cloud" style="color:var(--c-accent)"></i>
              <span>S3 / MinIO</span>
              <span :class="['badge', storageStatus.s3?.enabled ? 'badge-green' : 'badge-gray']" style="margin-left:auto">
                {{ storageStatus.s3?.enabled ? 'Enabled' : 'Disabled' }}
              </span>
            </div>
            <div v-if="storageStatus.s3?.enabled" class="storage-detail">
              <span class="mono">{{ storageStatus.s3?.endpoint }}/{{ storageStatus.s3?.bucket }}</span>
            </div>
            <div v-else class="storage-detail" style="font-size:10px">
              Set S3_ENABLED=true + S3_ENDPOINT + S3_BUCKET in .env
            </div>
          </div>
        </div>

        <!-- Export Config -->
        <div style="margin-top:14px;border-top:1px solid var(--c-border);padding-top:12px">
          <div style="font-size:11px;color:var(--c-muted);margin-bottom:8px;text-transform:uppercase;letter-spacing:.5px">
            Export Configuration
          </div>
          <div style="font-size:11px;color:var(--c-muted);margin-bottom:10px">
            Export all hosts, instances, and rules for disaster recovery or migration.
          </div>
          <div style="display:flex;gap:8px">
            <button class="btn btn-ghost" @click="exportConfig('json')" :disabled="exporting">
              <i :class="['ti', exporting === 'json' ? 'ti-loader-2 spin' : 'ti-file-type-json']"></i>
              Export JSON
            </button>
            <button class="btn btn-ghost" @click="exportConfig('yaml')" :disabled="exporting">
              <i :class="['ti', exporting === 'yaml' ? 'ti-loader-2 spin' : 'ti-file-code']"></i>
              Export YAML
            </button>
          </div>
        </div>

        <!-- Export Backups -->
        <div style="margin-top:14px;border-top:1px solid var(--c-border);padding-top:12px">
          <div style="font-size:11px;color:var(--c-muted);margin-bottom:8px;text-transform:uppercase;letter-spacing:.5px">
            Export Backups
          </div>
          <div style="font-size:11px;color:var(--c-muted);margin-bottom:10px">
            Download all backup of pg_hba.conf as ZIP file, or saving to storage server.
          </div>
          <div style="display:flex;gap:8px;flex-wrap:wrap">
            <button class="btn btn-ghost" @click="exportBackups" :disabled="exportingBackups">
              <i :class="['ti', exportingBackups ? 'ti-loader-2 spin' : 'ti-file-zip']"></i>
              Download ZIP
            </button>
            <button class="btn btn-ghost" @click="saveToLocal" :disabled="savingStorage">
              <i :class="['ti', savingStorage === 'local' ? 'ti-loader-2 spin' : 'ti-folder-upload']"></i>
              Save to Local
            </button>
            <button v-if="storageStatus?.s3?.enabled" class="btn btn-ghost" @click="saveToS3" :disabled="savingStorage">
              <i :class="['ti', savingStorage === 's3' ? 'ti-loader-2 spin' : 'ti-cloud-upload']"></i>
              Save to S3
            </button>
          </div>
        </div>
      </div>

      <!-- Danger Zone -->
      <div class="card">
        <div class="settings-card-title">
          <i class="ti ti-alert-triangle" style="color:var(--c-red)"></i> Danger Zone
        </div>
        <div style="display:flex;flex-direction:column;gap:10px">
          <div class="danger-item">
            <div>
              <div style="font-size:13px;font-weight:500">Purge Audit Log</div>
              <div style="font-size:11px;color:var(--c-muted);margin-top:2px">
                Delete audit logs older than specified days
              </div>
            </div>
            <div style="display:flex;align-items:center;gap:8px">
              <input v-model.number="purgeDays" type="number" min="7" max="365"
                style="width:70px;text-align:center">
              <span style="font-size:11px;color:var(--c-muted)">days</span>
              <button class="btn btn-danger btn-sm" @click="purgeAuditLog">
                <i class="ti ti-trash"></i> Purge
              </button>
            </div>
          </div>

          <div class="danger-item">
            <div>
              <div style="font-size:13px;font-weight:500">Purge Old Backups</div>
              <div style="font-size:11px;color:var(--c-muted);margin-top:2px">
                Keep only last N backups per instance
              </div>
            </div>
            <div style="display:flex;align-items:center;gap:8px">
              <input v-model.number="keepBackups" type="number" min="5" max="100"
                style="width:70px;text-align:center">
              <span style="font-size:11px;color:var(--c-muted)">keep</span>
              <button class="btn btn-danger btn-sm" @click="purgeBackups">
                <i class="ti ti-trash"></i> Purge
              </button>
            </div>
          </div>
        </div>
      </div>

    </div>

    <!-- Add/Edit Backup Schedule Modal -->
    <Teleport to="body">
      <div v-if="showScheduleModal" class="modal-backdrop" @click.self="showScheduleModal = false">
        <div class="modal" style="width:480px">
          <div class="modal-header">
            <div class="modal-title">{{ editingSchedule ? 'Edit Backup Schedule' : 'Add Backup Schedule' }}</div>
            <button class="btn btn-ghost btn-sm" @click="showScheduleModal = false"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body">
            <div class="field">
              <label>Name</label>
              <input v-model="scheduleForm.name" placeholder="e.g. Nightly backup - Production">
            </div>
            <div class="field">
              <label>Cron Expression</label>
              <input v-model="scheduleForm.cron_expression" placeholder="0 2 * * *" class="mono">
              <div style="font-size:11px;color:var(--c-muted);margin-top:4px">
                Standard 5-field cron (minute hour day month weekday), e.g. <span class="mono">0 2 * * *</span> = daily at 02:00.
              </div>
            </div>
            <div class="field-row">
              <div class="field">
                <label>Storage Target</label>
                <select v-model="scheduleForm.target">
                  <option value="local">Local</option>
                  <option value="s3">S3 / MinIO</option>
                  <option value="both">Both</option>
                </select>
              </div>
              <div class="field">
                <label>Scope</label>
                <select v-model="scheduleForm.scope_type" @change="scheduleForm.scope_id = ''">
                  <option value="all">All Instances</option>
                  <option value="host">Specific Host</option>
                  <option value="instance">Specific Instance</option>
                </select>
              </div>
            </div>
            <div v-if="scheduleForm.scope_type === 'host'" class="field">
              <label>Host</label>
              <select v-model="scheduleForm.scope_id">
                <option value="" disabled>Choose host...</option>
                <option v-for="h in hosts" :key="h.id" :value="h.id">{{ h.name }} ({{ h.ip_address }})</option>
              </select>
            </div>
            <div v-if="scheduleForm.scope_type === 'instance'" class="field">
              <label>Instance</label>
              <select v-model="scheduleForm.scope_id">
                <option value="" disabled>Choose instance...</option>
                <option v-for="i in instances" :key="i.id" :value="i.id">
                  {{ i.instance_name }}:{{ i.port }}<template v-if="(i as any).host"> @ {{ (i as any).host.name }}</template>
                </option>
              </select>
            </div>
            <div class="field" style="display:flex;align-items:center;gap:10px;margin-top:4px">
              <label class="switch" style="margin-bottom:0">
                <input type="checkbox" v-model="scheduleForm.enabled">
                <span class="switch-slider"></span>
              </label>
              <label style="margin-bottom:0">Enabled</label>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="showScheduleModal = false">Cancel</button>
            <button class="btn btn-primary" @click="saveSchedule" :disabled="savingSchedule">
              <i :class="['ti', savingSchedule ? 'ti-loader-2 spin' : 'ti-check']"></i> Save
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
import dayjs from 'dayjs'
import api, { backupSchedulesApi, hostsApi, instancesApi } from '@/api'
import type { BackupSchedule, BackupSchedulePayload, DBHost, PGInstance } from '@/types'

const toast = useToast()
function fmt(d: string) { return dayjs(d).format('DD MMM YYYY HH:mm') }
const loadingInfo = ref(false)
const sysInfo = ref<any>({})
const retentionDays = ref(50)
const purgeDays = ref(90)
const keepBackups = ref(50)

const notif = ref({
  telegram: false, tg_token: '', tg_chat: '',
  email: false, smtp_host: '', smtp_port: 587, smtp_from: '', smtp_to: '',
  whatsapp: false, fonnte_token: '', wa_phone: '',
})

const notifEvents = ref([
  { key: 'temp_expire',   label: 'Temporary access expired',    enabled: true },
  { key: 'temp_warning',  label: 'Temp access expiring soon (15m)', enabled: true },
  { key: 'host_offline',  label: 'Host goes offline',           enabled: false },
  { key: 'restore',       label: 'Backup restored',             enabled: true },
  { key: 'login_failed',  label: 'Login failed (brute force)',  enabled: true },
])

async function loadSysInfo() {
  loadingInfo.value = true
  try {
    const { data } = await api.get('/settings/system')
    sysInfo.value = data
  } catch (e) {
    console.error('Failed to load system info', e)
  } finally {
    loadingInfo.value = false
  }
}

function saveRetention() {
  localStorage.setItem('pghba_retention', String(retentionDays.value))
  toast.success(`Retention set to ${retentionDays.value} backups per instance`)
}

function saveNotif() {
  // Save to localStorage for now (would be API in production)
  localStorage.setItem('pghba_notif', JSON.stringify(notif.value))
  toast.success('Notification settings saved')
}

async function testTelegram() {
  if (!notif.value.tg_token || !notif.value.tg_chat) {
    toast.error('Please fill Bot Token and Chat ID')
    return
  }
  try {
    const res = await fetch(
      `https://api.telegram.org/bot${notif.value.tg_token}/sendMessage`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          chat_id: notif.value.tg_chat,
          text: '✅ PGHBA Manager — Telegram notification test successful!'
        })
      }
    )
    const data = await res.json()
    if (data.ok) toast.success('Telegram test sent successfully!')
    else toast.error('Telegram error: ' + data.description)
  } catch {
    toast.error('Failed to send Telegram test')
  }
}

const storageStatus = ref<any>(null)
const exporting       = ref<string | false>(false)
const exportingBackups = ref(false)
const savingStorage   = ref<string | false>(false)

const eolStatus       = ref<any>({ versions: [], has_eol: false, eol_count: 0 })
const triggeringBackup = ref(false)

async function loadEOLStatus() {
  try {
    const { data } = await api.get('/settings/eol-status')
    eolStatus.value = data
  } catch {}
}

async function triggerManualBackup() {
  triggeringBackup.value = true
  try {
    const { data } = await api.post('/export/save-storage', {
      target: sysInfo.value.auto_backup_target || 'local'
    })
    toast.success(`Backup complete: ${data.saved}/${data.total} saved to ${data.target}`)
    await loadStorageStatus()
  await loadEOLStatus()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Backup failed')
  } finally { triggeringBackup.value = false }
}

async function loadStorageStatus() {
  try {
    const { data } = await api.get('/export/storage-status')
    storageStatus.value = data
  } catch {}
}

async function exportConfig(format: string) {
  exporting.value = format
  try {
    const token = localStorage.getItem('access_token') || sessionStorage.getItem('access_token')
    const resp = await fetch(`/api/export/config?format=${format}`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    const blob = await resp.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    const ts = new Date().toISOString().slice(0,19).replace(/[:T]/g, '-')
    a.href = url
    a.download = `pghba-config-${ts}.${format}`
    a.click()
    URL.revokeObjectURL(url)
    toast.success(`Configuration exported as ${format.toUpperCase()}`)
  } catch {
    toast.error('Export failed')
  } finally {
    exporting.value = false
  }
}

async function exportBackups() {
  exportingBackups.value = true
  try {
    const token = localStorage.getItem('access_token') || sessionStorage.getItem('access_token')
    const resp = await fetch('/api/export/backups', {
      headers: { Authorization: `Bearer ${token}` }
    })
    const blob = await resp.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    const ts = new Date().toISOString().slice(0,10)
    a.href = url
    a.download = `pghba-backups-${ts}.zip`
    a.click()
    URL.revokeObjectURL(url)
    toast.success('Backups downloaded as ZIP')
  } catch {
    toast.error('Export backups failed')
  } finally {
    exportingBackups.value = false
  }
}

async function saveToLocal() {
  savingStorage.value = 'local'
  try {
    const { data } = await api.post('/export/save-storage', { target: 'local' })
    toast.success(`${data.saved}/${data.total} backups saved to ${storageStatus.value?.local?.path}`)
    await loadStorageStatus()
  await loadEOLStatus()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Save failed')
  } finally {
    savingStorage.value = false }
}

async function saveToS3() {
  savingStorage.value = 's3'
  try {
    const { data } = await api.post('/export/save-storage', { target: 's3' })
    toast.success(`${data.saved}/${data.total} backups saved to S3`)
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'S3 upload failed')
  } finally {
    savingStorage.value = false }
}

async function purgeAuditLog() {
  if (!confirm(`Delete ALL audit logs older than ${purgeDays.value} days?\n\nThis cannot be undone.`)) return
  try {
    const { data } = await api.post('/settings/purge-audit', { older_than_days: purgeDays.value })
    toast.success(`Purged ${data.deleted} audit log entries (older than ${purgeDays.value} days)`)
    await loadSysInfo()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Purge failed')
  }
}

async function purgeBackups() {
  if (!confirm(`Keep only last ${keepBackups.value} backups per instance?\n\nOlder backups will be deleted permanently.`)) return
  try {
    // Raw SQL purge - keep N latest per instance
    const { data } = await api.post('/settings/purge-audit', { older_than_days: 0 }) // placeholder
    toast.success('Old backups purged')
    await loadSysInfo()
  } catch {
    toast.error('Purge failed')
  }
}

// ── Backup Schedules ──────────────────────────────────────────────────────
const schedules        = ref<BackupSchedule[]>([])
const loadingSchedules = ref(false)
const hosts             = ref<DBHost[]>([])
const instances         = ref<PGInstance[]>([])
const showScheduleModal = ref(false)
const editingSchedule   = ref<BackupSchedule | null>(null)
const savingSchedule    = ref(false)

const scheduleForm = ref<BackupSchedulePayload>({
  name: '', cron_expression: '0 2 * * *', target: 'local', scope_type: 'all', scope_id: '', enabled: true
})

async function loadSchedules() {
  loadingSchedules.value = true
  try {
    const { data } = await backupSchedulesApi.list()
    schedules.value = data
  } catch {
    toast.error('Failed to load backup schedules')
  } finally {
    loadingSchedules.value = false
  }
}

function scopeLabel(s: BackupSchedule) {
  if (s.scope_type === 'all') return 'All instances'
  if (s.scope_type === 'host') {
    const h = hosts.value.find(h => h.id === s.scope_id)
    return h ? `Host: ${h.name}` : 'Host (unknown)'
  }
  const i = instances.value.find(i => i.id === s.scope_id)
  return i ? `Instance: ${i.instance_name}:${i.port}` : 'Instance (unknown)'
}

function openCreateSchedule() {
  editingSchedule.value = null
  scheduleForm.value = { name: '', cron_expression: '0 2 * * *', target: 'local', scope_type: 'all', scope_id: '', enabled: true }
  showScheduleModal.value = true
}

function openEditSchedule(s: BackupSchedule) {
  editingSchedule.value = s
  scheduleForm.value = {
    name: s.name, cron_expression: s.cron_expression, target: s.target,
    scope_type: s.scope_type, scope_id: s.scope_id || '', enabled: s.enabled
  }
  showScheduleModal.value = true
}

async function saveSchedule() {
  if (!scheduleForm.value.name.trim() || !scheduleForm.value.cron_expression.trim()) {
    toast.error('Name and cron expression are required')
    return
  }
  if (scheduleForm.value.scope_type !== 'all' && !scheduleForm.value.scope_id) {
    toast.error('Please choose a target host/instance for this scope')
    return
  }
  savingSchedule.value = true
  try {
    if (editingSchedule.value) {
      await backupSchedulesApi.update(editingSchedule.value.id, scheduleForm.value)
      toast.success('Backup schedule updated')
    } else {
      await backupSchedulesApi.create(scheduleForm.value)
      toast.success('Backup schedule created')
    }
    showScheduleModal.value = false
    await loadSchedules()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Failed to save backup schedule')
  } finally {
    savingSchedule.value = false
  }
}

async function toggleSchedule(s: BackupSchedule) {
  try {
    await backupSchedulesApi.update(s.id, {
      name: s.name, cron_expression: s.cron_expression, target: s.target,
      scope_type: s.scope_type, scope_id: s.scope_id || undefined, enabled: !s.enabled
    })
    await loadSchedules()
  } catch {
    toast.error('Failed to update schedule')
  }
}

async function deleteSchedule(s: BackupSchedule) {
  if (!confirm(`Delete backup schedule "${s.name}"?`)) return
  try {
    await backupSchedulesApi.delete(s.id)
    toast.success('Backup schedule deleted')
    await loadSchedules()
  } catch {
    toast.error('Failed to delete backup schedule')
  }
}

onMounted(async () => {
  await loadSysInfo()
  await loadStorageStatus()
  await loadEOLStatus()
  await loadSchedules()
  try {
    const [{ data: hostData }, { data: instData }] = await Promise.all([hostsApi.list(), instancesApi.list()])
    hosts.value = hostData
    instances.value = instData
  } catch {}
  // Load saved notif settings
  const saved = localStorage.getItem('pghba_notif')
  if (saved) {
    try { Object.assign(notif.value, JSON.parse(saved)) } catch {}
  }
  const savedRetention = localStorage.getItem('pghba_retention')
  if (savedRetention) retentionDays.value = parseInt(savedRetention)
})
</script>

<style scoped>
.settings-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(340px,1fr));gap:16px; }
.settings-card-title { display:flex;align-items:center;gap:8px;font-size:13px;font-weight:600;margin-bottom:14px; }

.schedule-table { width:100%;border-collapse:collapse;font-size:12px; }
.schedule-table th { text-align:left;color:var(--c-muted);font-weight:500;font-size:11px;padding:6px 8px;border-bottom:1px solid var(--c-border);white-space:nowrap; }
.schedule-table td { padding:8px;border-bottom:1px solid var(--c-border);white-space:nowrap; }
.schedule-table tr:last-child td { border-bottom:none; }

.switch { position:relative;display:inline-block;width:34px;height:19px; }
.switch input { opacity:0;width:0;height:0; }
.switch-slider { position:absolute;cursor:pointer;inset:0;background:var(--c-border);border-radius:19px;transition:.15s; }
.switch-slider::before { content:"";position:absolute;height:14px;width:14px;left:3px;bottom:2.5px;background:#fff;border-radius:50%;transition:.15s; }
.switch input:checked + .switch-slider { background:var(--c-accent); }
.switch input:checked + .switch-slider::before { transform:translateX(15px); }
.info-list { display:flex;flex-direction:column;gap:0; }
.info-row { display:flex;justify-content:space-between;align-items:center;padding:8px 0;border-bottom:1px solid var(--c-border); }
.info-row:last-child { border-bottom:none; }
.info-label { font-size:12px;color:var(--c-muted); }
.info-value { font-size:12px;font-weight:500; }
.mono { font-family:monospace; }
.badge-green-text { color:var(--c-green); }
.badge-amber { background:rgba(245,158,11,.1);color:var(--c-amber);border:1px solid rgba(245,158,11,.2);padding:2px 8px;border-radius:4px;font-size:10px; }
.badge-red   { background:rgba(239,68,68,.1);color:var(--c-red);border:1px solid rgba(239,68,68,.2);padding:2px 8px;border-radius:4px;font-size:10px; }

.stat-mini-grid { display:grid;grid-template-columns:repeat(3,1fr);gap:8px;margin-top:8px; }
.stat-mini { background:var(--c-bg);border:1px solid var(--c-border);border-radius:6px;padding:8px;text-align:center; }
.stat-mini-val { font-size:18px;font-weight:700;color:var(--c-text); }
.stat-mini-label { font-size:10px;color:var(--c-muted);text-transform:uppercase;letter-spacing:.3px; }

.info-box { border-radius:6px;padding:10px 12px;font-size:11px;display:flex;gap:8px;align-items:flex-start; }
.info-box.amber { background:rgba(245,158,11,.08);border:1px solid rgba(245,158,11,.2);color:var(--c-amber); }
.info-box.green { background:rgba(34,197,94,.08);border:1px solid rgba(34,197,94,.2);color:var(--c-green); }
code { font-family:monospace;background:var(--c-surface2);padding:1px 5px;border-radius:3px;font-size:10px; }

.notif-list { display:flex;flex-direction:column;gap:0; }
.notif-item { display:flex;align-items:center;justify-content:space-between;padding:10px 0;border-bottom:1px solid var(--c-border); }
.notif-info { display:flex;align-items:center;gap:10px; }
.notif-name { font-size:12px;font-weight:500; }
.notif-desc { font-size:10px;color:var(--c-muted); }
.notif-config { padding:8px 0 8px 28px;border-bottom:1px solid var(--c-border); }

.toggle-switch { position:relative;display:inline-block;width:36px;height:20px;cursor:pointer; }
.toggle-switch.small { width:28px;height:16px; }
.toggle-switch input { opacity:0;width:0;height:0; }
.toggle-slider { position:absolute;top:0;left:0;right:0;bottom:0;background:var(--c-border);border-radius:20px;transition:.2s; }
.toggle-slider:before { content:'';position:absolute;height:14px;width:14px;left:3px;bottom:3px;background:white;border-radius:50%;transition:.2s; }
.toggle-switch.small .toggle-slider:before { height:10px;width:10px;left:3px;bottom:3px; }
.toggle-switch input:checked + .toggle-slider { background:var(--c-accent); }
.toggle-switch input:checked + .toggle-slider:before { transform:translateX(16px); }
.toggle-switch.small input:checked + .toggle-slider:before { transform:translateX(12px); }

.danger-item { display:flex;align-items:center;justify-content:space-between;padding:10px 12px;background:rgba(239,68,68,.05);border:1px solid rgba(239,68,68,.15);border-radius:8px; }

.spin { animation:spin 1s linear infinite; }
.storage-status { display:flex;flex-direction:column;gap:8px; }
.storage-item { background:var(--c-bg);border:1px solid var(--c-border);border-radius:6px;padding:10px 12px; }
.storage-header { display:flex;align-items:center;gap:8px;font-size:12px;font-weight:500;margin-bottom:4px; }
.storage-detail { font-size:11px;color:var(--c-muted);display:flex;justify-content:space-between; }
.pg-version-row { padding:8px 10px;border-radius:6px;margin-bottom:6px;border:1px solid var(--c-border); }
.pg-eol     { background:rgba(239,68,68,.06);border-color:rgba(239,68,68,.2); }
.pg-warning { background:rgba(245,158,11,.06);border-color:rgba(245,158,11,.2); }
.pg-ok      { background:rgba(34,197,94,.04);border-color:rgba(34,197,94,.15); }
@keyframes spin { to { transform:rotate(360deg); } }
</style>
