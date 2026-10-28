<template>
  <div class="page-content">
    <!-- Toolbar -->
    <div class="toolbar" style="flex-wrap:wrap;gap:8px">
      <select class="filter-select" v-model="filters.environment" @change="onFilterChange" style="width:130px">
        <option value="">All Environments</option>
        <option>Production</option><option>Staging</option><option>UAT</option><option>Development</option>
      </select>
      <select class="filter-select" v-model="filters.instance_id" @change="onFilterChange" style="width:200px">
        <option value="">All Instances</option>
        <option v-for="i in filteredInstances" :key="i.id" :value="i.id">
          {{ i.instance_name }}:{{ i.port }} PG{{ i.pg_version }}
        </option>
      </select>
      <select class="filter-select" v-model="filters.rule_type" @change="onFilterChange" style="width:110px">
        <option value="">All Types</option>
        <option>host</option><option>hostssl</option><option>hostnossl</option><option>local</option>
      </select>
      <select class="filter-select" v-model="filters.enabled" @change="onFilterChange" style="width:110px">
        <option value="">All Status</option>
        <option value="true">Enabled</option>
        <option value="false">Disabled</option>
      </select>
      <div style="flex:1;max-width:260px">
        <div class="search-inline">
          <i class="ti ti-search" style="font-size:12px;color:var(--c-muted)"></i>
          <input v-model="globalSearch" placeholder="Search IP, database, instance...">
        </div>
      </div>
      <div class="toolbar-right">
        <button class="btn btn-ghost" @click="reloadAll" title="Refresh the rule list from the database (does not touch the server)">
          <i class="ti ti-refresh"></i> Refresh List
        </button>
        <button v-if="auth.canManageRules && singleVisibleInstance" class="btn btn-ghost" @click="openWriteReload"
          style="color:var(--c-amber);border-color:var(--c-amber)"
          title="Push saved rules to pg_hba.conf on the server and reload PostgreSQL">
          <i class="ti ti-upload"></i> Write & Reload
        </button>
        <button class="btn btn-ghost"
          :style="conflicts.errors > 0 ? 'color:var(--c-red);border-color:var(--c-red)' : conflicts.warnings > 0 ? 'color:var(--c-amber);border-color:var(--c-amber)' : 'color:#8b5cf6;border-color:#8b5cf6'"
          @click="checkConflicts">
          <i class="ti ti-alert-triangle"></i> Check Conflicts
          <span v-if="conflicts.total > 0"
            style="font-size:9px;font-weight:700;padding:1px 5px;border-radius:10px;margin-left:4px;background:var(--c-red);color:#fff">
            {{ conflicts.total }}
          </span>
        </button>
        <button v-if="auth.canManageRules" class="btn btn-ghost" @click="showClone = true"
          style="color:var(--c-teal);border-color:var(--c-teal)">
          <i class="ti ti-copy"></i> Clone
        </button>
        <button v-if="auth.canManageRules && filters.instance_id" class="btn btn-ghost"
          :style="dragMode ? 'background:rgba(139,92,246,.15);border-color:#8b5cf6;color:#8b5cf6' : 'color:#8b5cf6;border-color:#8b5cf6'"
          @click="toggleDragMode">
          <i class="ti ti-arrows-sort"></i> {{ dragMode ? 'Exit Reorder' : 'Reorder' }}
          <span v-if="dragMode && savingOrder" style="margin-left:4px"><i class="ti ti-loader-2 spin" style="font-size:11px"></i></span>
        </button>
        <button v-if="auth.canManageRules" class="btn btn-primary" @click="openCreate">
          <i class="ti ti-plus"></i> Add Rule
        </button>
      </div>
    </div>

    <!-- Drag mode banner -->
    <div v-if="dragMode"
      style="background:rgba(139,92,246,.1);border:1px solid rgba(139,92,246,.3);border-radius:8px;padding:10px 14px;margin-bottom:12px;font-size:12px;color:#8b5cf6;display:flex;align-items:center;gap:8px">
      <i class="ti ti-arrows-sort"></i>
      <strong>Reorder Mode</strong> — Drag rows up/down to change rule order. Changes save automatically and pg_hba.conf will be updated.
      <button class="btn btn-ghost btn-sm" @click="toggleDragMode" style="margin-left:auto;color:#8b5cf6">
        <i class="ti ti-x"></i> Exit
      </button>
    </div>

    <div class="table-wrap">
      <div class="table-header">
        <span class="table-title">
          Access Rules
          <span class="total-count">{{ filtered.length }} rules</span>
        </span>
        <div style="display:flex;align-items:center;gap:8px">
          <span style="font-size:12px;color:var(--c-muted)">Sort:</span>
          <select class="filter-select" v-model="sortBy" :disabled="dragMode" style="width:140px">
            <option value="rule_order_asc">Rule Order ↑</option>
            <option value="rule_order_desc">Rule Order ↓</option>
            <option value="disabled_first">Disabled First</option>
            <option value="enabled_first">Enabled First</option>
            <option value="address_asc">IP Address ↑</option>
            <option value="address_desc">IP Address ↓</option>
            <option value="created_desc">Newest First</option>
            <option value="created_asc">Oldest First</option>
          </select>
          <span style="font-size:12px;color:var(--c-muted)">Per page:</span>
          <select class="filter-select" v-model="perPage" @change="currentPage = 1" :disabled="dragMode" style="width:70px">
            <option :value="10">10</option><option :value="20">20</option>
            <option :value="30">30</option><option :value="50">50</option><option :value="100">100</option>
          </select>
        </div>
      </div>

      <table>
        <thead>
          <tr>
            <th v-if="dragMode" style="width:36px"></th>
            <th>Type</th><th>Database</th><th>User</th>
            <th>CIDR Address</th><th>Method</th><th>Instance</th>
            <th>Host</th><th>Comment</th><th>Status</th><th></th>
          </tr>
        </thead>
        <tbody ref="tbodyRef">
          <tr v-if="loading">
            <td :colspan="dragMode ? 11 : 10" style="text-align:center;padding:24px;color:var(--c-muted)">
              <i class="ti ti-loader-2 spin"></i> Loading...
            </td>
          </tr>
          <tr v-for="(rule, idx) in displayRules" :key="rule.id"
            :class="['rule-row', { 'row-disabled': !rule.enabled, 'row-temp': rule.is_temporary, 'drag-over': dragMode && dragOverIdx === idx, 'row-dragging': dragMode && dragSrcIdx === idx }]"
            :draggable="dragMode"
            @dragstart="dragMode ? onDragStart($event, idx) : null"
            @dragover.prevent="dragMode ? onDragOver($event, idx) : null"
            @dragleave="dragMode ? onDragLeave() : null"
            @drop.prevent="dragMode ? onDrop($event, idx) : null"
            @dragend="dragMode ? onDragEnd() : null">
            <!-- Drag handle -->
            <td v-if="dragMode" class="drag-handle-cell">
              <i class="ti ti-grip-vertical drag-icon"></i>
            </td>
            <td>
              <span :class="['rule-type', 'type-' + rule.rule_type]">{{ rule.rule_type }}</span>
            </td>
            <td>{{ rule.database }}</td>
            <td>{{ rule.db_user }}</td>
            <td class="mono">{{ rule.address || '—' }}</td>
            <td class="method">{{ rule.auth_method }}</td>
            <td style="font-size:11px">
              {{ rule.instance?.instance_name }}:{{ rule.instance?.port }}
              <span style="color:var(--c-muted)">(PG {{ rule.instance?.pg_version }})</span>
            </td>
            <td style="font-size:11px;color:var(--c-muted)">
              <span v-if="rule.instance?.host?.name">
                {{ rule.instance.host.name }}<br>
                <span style="font-family:monospace;font-size:10px">{{ rule.instance.host.ip_address }}</span>
              </span>
              <span v-else>—</span>
            </td>
            <td style="font-size:11px;color:var(--c-muted);max-width:150px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" :title="rule.comment">
              {{ rule.comment || '—' }}
            </td>
            <td>
              <label class="toggle" :title="rule.enabled ? 'Disable' : 'Enable'">
                <input type="checkbox" :checked="rule.enabled" @change="toggleRule(rule)">
                <span class="slider"></span>
              </label>
            </td>
            <td>
              <div style="display:flex;gap:3px">
                <button v-if="auth.canManageRules && !dragMode" class="btn btn-ghost btn-sm" @click="openEdit(rule)" title="Edit">
                  <i class="ti ti-edit"></i>
                </button>
                <button v-if="auth.isOperator && !rule.is_temporary && !dragMode" class="btn btn-ghost btn-sm"
                  @click="openSetTemp(rule)" title="Set temporary" style="color:var(--c-amber)">
                  <i class="ti ti-clock-plus"></i>
                </button>
                <button v-if="auth.canManageRules && !dragMode" class="btn btn-danger btn-sm" @click="deleteRule(rule)" title="Delete">
                  <i class="ti ti-trash"></i>
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && filtered.length === 0">
            <td :colspan="dragMode ? 11 : 10" style="text-align:center;padding:32px;color:var(--c-muted)">
              No rules found
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination (hidden in drag mode) -->
      <div class="pagination" v-if="totalPages > 1 && !dragMode">
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
      <div v-if="dragMode" style="padding:10px 16px;font-size:11px;color:var(--c-muted);border-top:1px solid var(--c-border);text-align:center">
        Showing all {{ dragRules.length }} rules for reordering
      </div>
    </div>

    <!-- Conflict Modal -->
    <Teleport to="body">
      <div v-if="showConflicts" class="modal-backdrop" @click.self="showConflicts = false">
        <div class="modal" style="width:700px;max-height:80vh;display:flex;flex-direction:column">
          <div class="modal-header">
            <div class="modal-title">
              <i class="ti ti-alert-triangle" :style="{ color: conflicts.errors > 0 ? 'var(--c-red)' : 'var(--c-amber)' }"></i>
              Rule Conflicts — {{ conflicts.total }} found
            </div>
            <button class="btn btn-ghost btn-sm" @click="showConflicts = false"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body" style="overflow-y:auto;flex:1">
            <div v-if="conflicts.total === 0" style="text-align:center;padding:32px;color:var(--c-green)">
              <i class="ti ti-circle-check" style="font-size:32px;display:block;margin-bottom:8px"></i>
              No conflicts detected!
            </div>
            <div v-else>
              <div style="display:flex;gap:8px;margin-bottom:14px">
                <span v-if="conflicts.errors > 0" style="display:inline-flex;align-items:center;gap:4px;background:rgba(239,68,68,.1);color:var(--c-red);border:1px solid rgba(239,68,68,.2);padding:3px 10px;border-radius:5px;font-size:11px">
                  <i class="ti ti-x-circle"></i> {{ conflicts.errors }} errors
                </span>
                <span v-if="conflicts.warnings > 0" style="display:inline-flex;align-items:center;gap:4px;background:rgba(245,158,11,.1);color:var(--c-amber);border:1px solid rgba(245,158,11,.2);padding:3px 10px;border-radius:5px;font-size:11px">
                  <i class="ti ti-alert-triangle"></i> {{ conflicts.warnings }} warnings
                </span>
              </div>
              <div v-for="(cf, i) in conflicts.items" :key="i"
                :style="{ background: cf.severity==='error' ? 'rgba(239,68,68,.08)' : 'rgba(245,158,11,.08)', border: '1px solid ' + (cf.severity==='error' ? 'rgba(239,68,68,.2)' : 'rgba(245,158,11,.2)'), borderRadius: '8px', padding: '12px', marginBottom: '10px' }">
                <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px">
                  <span :style="{ background: cf.severity==='error' ? 'rgba(239,68,68,.15)' : 'rgba(245,158,11,.15)', color: cf.severity==='error' ? 'var(--c-red)' : 'var(--c-amber)', padding: '2px 7px', borderRadius: '4px', fontSize: '10px', fontWeight: 600 }">
                    {{ cf.type.toUpperCase() }}
                  </span>
                  <span style="font-size:12px;font-weight:600">{{ cf.description }}</span>
                </div>
                <div style="font-size:11px;font-family:monospace;background:var(--c-bg);border:1px solid var(--c-border);border-radius:4px;padding:8px;margin-bottom:6px">
                  <div style="color:var(--c-muted);font-size:10px;margin-bottom:4px">
                    Instance: <strong style="color:var(--c-text)">{{ cf.instance_label || cf.rule_a?.instance?.instance_name }}</strong>
                  </div>
                  <div><span style="color:var(--c-muted)">Rule A (order {{ cf.order_a }}):</span> {{ cf.rule_a?.rule_type }} {{ cf.rule_a?.database }} {{ cf.rule_a?.db_user }} {{ cf.rule_a?.address }} <strong>{{ cf.rule_a?.auth_method }}</strong></div>
                  <div><span style="color:var(--c-muted)">Rule B (order {{ cf.order_b }}):</span> {{ cf.rule_b?.rule_type }} {{ cf.rule_b?.database }} {{ cf.rule_b?.db_user }} {{ cf.rule_b?.address }} <strong>{{ cf.rule_b?.auth_method }}</strong></div>
                </div>
                <div style="font-size:11px;color:var(--c-muted)">
                  <i class="ti ti-bulb" style="color:var(--c-amber)"></i> {{ cf.suggestion }}
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="showConflicts = false">Close</button>
            <button class="btn btn-ghost" @click="checkConflicts"><i class="ti ti-refresh"></i> Re-check</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Clone Modal -->
    <Teleport to="body">
      <div v-if="showClone" class="modal-backdrop" @click.self="showClone = false; cloneResult = null">
        <div class="modal" style="width:500px">
          <div class="modal-header">
            <div class="modal-title" style="color:var(--c-teal)"><i class="ti ti-copy"></i> Clone Rules</div>
            <button class="btn btn-ghost btn-sm" @click="showClone = false; cloneResult = null"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body">
            <div v-if="!cloneResult">
              <div class="field">
                <label>Source (copy from)</label>
                <select v-model="cloneForm.source_instance_id" style="width:100%">
                  <option value="" disabled>Choose source...</option>
                  <option v-for="i in instances" :key="i.id" :value="i.id" :disabled="i.id === cloneForm.target_instance_id">
                    {{ i.instance_name }}:{{ i.port }}
                  </option>
                </select>
              </div>
              <div style="text-align:center;color:var(--c-muted);margin:6px 0"><i class="ti ti-arrow-down"></i></div>
              <div class="field">
                <label>Target (copy to)</label>
                <select v-model="cloneForm.target_instance_id" style="width:100%">
                  <option value="" disabled>Choose target...</option>
                  <option v-for="i in instances" :key="i.id" :value="i.id" :disabled="i.id === cloneForm.source_instance_id">
                    {{ i.instance_name }}:{{ i.port }}
                  </option>
                </select>
              </div>
              <div style="margin-top:10px;background:var(--c-bg);border:1px solid var(--c-border);border-radius:8px;padding:10px 12px;display:flex;flex-direction:column;gap:8px">
                <label style="display:flex;align-items:center;gap:8px;cursor:pointer;font-size:12px">
                  <input type="checkbox" v-model="cloneForm.overwrite_existing">
                  <span><strong>Overwrite existing rules</strong></span>
                </label>
                <label style="display:flex;align-items:center;gap:8px;cursor:pointer;font-size:12px">
                  <input type="checkbox" v-model="cloneForm.write_back">
                  <span><strong>Write & Reload after clone</strong></span>
                </label>
              </div>
            </div>
            <div v-if="cloneResult" style="background:rgba(34,197,94,.08);border:1px solid rgba(34,197,94,.2);border-radius:8px;padding:14px">
              <div style="font-weight:600;margin-bottom:4px"><i class="ti ti-circle-check" style="color:var(--c-green)"></i> Clone Complete</div>
              <div style="font-size:12px;color:var(--c-muted)">{{ cloneResult.source }} → {{ cloneResult.target }}</div>
              <div style="margin-top:6px;font-size:12px">{{ cloneResult.cloned }} cloned, {{ cloneResult.skipped }} skipped</div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="showClone = false; cloneResult = null">{{ cloneResult ? 'Close' : 'Cancel' }}</button>
            <button v-if="!cloneResult" class="btn btn-primary" @click="doClone"
              :disabled="cloning || !cloneForm.source_instance_id || !cloneForm.target_instance_id"
              style="background:var(--c-teal);border-color:var(--c-teal)">
              <i :class="['ti', cloning ? 'ti-loader-2 spin' : 'ti-copy']"></i>
              {{ cloning ? 'Cloning...' : 'Clone Rules' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Set Temporary Modal -->
    <Teleport to="body">
      <div v-if="tempTarget" class="modal-backdrop" @click.self="tempTarget = null">
        <div class="modal" style="width:420px">
          <div class="modal-header">
            <div class="modal-title"><i class="ti ti-clock-plus" style="color:var(--c-amber)"></i> Set Temporary</div>
            <button class="btn btn-ghost btn-sm" @click="tempTarget = null"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body">
            <div style="background:rgba(245,158,11,.08);border:1px solid rgba(245,158,11,.2);border-radius:6px;padding:10px;font-size:11px;color:var(--c-amber);margin-bottom:12px">
              <i class="ti ti-alert-triangle"></i> Rule akan di-disable otomatis saat waktu habis + pg_hba.conf di-update.
            </div>
            <div style="font-family:monospace;font-size:11px;background:var(--c-bg);border:1px solid var(--c-border);border-radius:4px;padding:8px;margin-bottom:12px">
              {{ tempTarget?.rule_type }} {{ tempTarget?.database }} {{ tempTarget?.db_user }} {{ tempTarget?.address }} {{ tempTarget?.auth_method }}
            </div>
            <div class="field-row">
              <div class="field"><label>Duration</label><input v-model.number="tempForm.duration_value" type="number" min="1" placeholder="15"></div>
              <div class="field"><label>Unit</label>
                <select v-model="tempForm.duration_unit">
                  <option>Minutes</option><option>Hours</option><option>Days</option><option>Weeks</option>
                </select>
              </div>
            </div>
            <div class="field"><label>Reason</label><input v-model="tempForm.reason" placeholder="Temporary access for..."></div>
            <div style="font-size:11px;color:var(--c-muted);margin-top:6px">
              Expires: <strong style="color:var(--c-amber)">{{ previewTempExpiry }}</strong>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="tempTarget = null">Cancel</button>
            <button class="btn btn-primary" @click="doSetTemp" :disabled="savingTemp"
              style="background:var(--c-amber);border-color:var(--c-amber)">
              <i :class="['ti', savingTemp ? 'ti-loader-2 spin' : 'ti-clock']"></i> Set Temporary
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Add/Edit Rule Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
        <div class="modal" style="width:480px">
          <div class="modal-header">
            <div class="modal-title">{{ editing ? 'Edit Rule' : 'Add Rule' }}</div>
            <button class="btn btn-ghost btn-sm" @click="showModal = false"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body">
            <div class="field">
              <label>Instance</label>
              <select v-model="form.instance_id" :disabled="!!editing">
                <option v-for="i in instances" :key="i.id" :value="i.id">
                  {{ i.instance_name }}:{{ i.port }}{{ i.host?.ip_address ? ` — ${i.host.ip_address}` : '' }}
                </option>
              </select>
              <div v-if="selectedInstance?.host" class="text-muted" style="font-size:12px;margin-top:4px">
                <i class="ti ti-server"></i> {{ selectedInstance.host.name }} &middot; {{ selectedInstance.host.ip_address }}
              </div>
            </div>
            <div class="field-row">
              <div class="field"><label>Type</label>
                <select v-model="form.rule_type">
                  <option>host</option><option>hostssl</option><option>hostnossl</option><option>local</option>
                </select>
              </div>
              <div class="field"><label>Method</label>
                <select v-model="form.auth_method">
                  <option>md5</option><option>scram-sha-256</option><option>trust</option><option>reject</option><option>peer</option><option>ident</option>
                </select>
              </div>
            </div>
            <div class="field-row">
              <div class="field"><label>Database</label><input v-model="form.database" placeholder="all"></div>
              <div class="field"><label>User</label><input v-model="form.db_user" placeholder="all"></div>
            </div>
            <div v-if="form.rule_type !== 'local'" class="field">
              <label>CIDR Address</label><input v-model="form.address" placeholder="0.0.0.0/0">
            </div>
            <div class="field"><label>Comment</label><input v-model="form.comment" placeholder="Optional..."></div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="showModal = false">Cancel</button>
            <button class="btn btn-primary" @click="saveRule" :disabled="saving">
              <i :class="['ti', saving ? 'ti-loader-2 spin' : 'ti-check']"></i> Save
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Write & Reload Modal -->
    <Teleport to="body">
      <div v-if="writeTarget" class="modal-backdrop" @click.self="writeTarget = null">
        <div class="modal" style="width:480px">
          <div class="modal-header">
            <div class="modal-title" style="color:var(--c-amber)">
              <i class="ti ti-upload"></i> Write & Reload — {{ writeTarget.instance_name }}:{{ writeTarget.port }}
            </div>
            <button class="btn btn-ghost btn-sm" @click="writeTarget = null"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body">
            <div class="warning-box">
              <i class="ti ti-alert-triangle"></i>
              <div>
                <div style="font-weight:600;margin-bottom:6px">Tindakan ini akan:</div>
                <ol style="margin-left:16px;line-height:1.9;font-size:12px;color:var(--c-muted)">
                  <li>Menulis semua rule tersimpan untuk instance ini ke <span class="mono">pg_hba.conf</span> di server</li>
                  <li>Reload PostgreSQL supaya perubahan berlaku</li>
                </ol>
              </div>
            </div>
            <div style="margin-top:12px;font-size:12px;color:var(--c-muted)">
              Untuk melihat perbandingan sebelum-sesudah, gunakan tombol "Write & Reload" di halaman
              <router-link to="/instances" style="color:var(--c-accent)">PG Instances</router-link>.
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="writeTarget = null">Cancel</button>
            <button class="btn btn-primary" @click="doWriteReload" :disabled="writingReload"
              style="background:var(--c-amber);border-color:var(--c-amber)">
              <i :class="['ti', writingReload ? 'ti-loader-2 spin' : 'ti-upload']"></i>
              {{ writingReload ? 'Writing...' : 'Write & Reload' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useToast } from 'vue-toastification'
import { rulesApi, instancesApi } from '@/api'
import api from '@/api'
import { globalSearch } from '@/stores/search'
import { useAuthStore } from '@/stores/auth'
import type { HBARule, PGInstance, CreateRulePayload } from '@/types'

const toast   = useToast()
const auth    = useAuthStore()
const route   = useRoute()

const rules     = ref<HBARule[]>([])
const instances = ref<PGInstance[]>([])
const loading   = ref(false)
const showModal = ref(false)
const editing   = ref<HBARule | null>(null)
const saving    = ref(false)

// Write & Reload
const writeTarget   = ref<PGInstance | null>(null)
const writingReload  = ref(false)

// Filters
const filters = ref({
  environment: '',
  instance_id: (route.query.instance_id as string) || '',
  rule_type: '',
  enabled: '',
})

// Sort & Pagination
const sortBy      = ref('rule_order_asc')
const perPage     = ref(20)
const currentPage = ref(1)

// Drag & Drop
const dragMode    = ref(false)
const dragRules   = ref<HBARule[]>([])
const dragSrcIdx  = ref<number | null>(null)
const dragOverIdx = ref<number | null>(null)
const savingOrder = ref(false)

// Clone
const showClone   = ref(false)
const cloning     = ref(false)
const cloneResult = ref<any>(null)
const cloneForm   = ref({ source_instance_id: '', target_instance_id: '', overwrite_existing: false, write_back: true })

// Set Temporary
const tempTarget  = ref<HBARule | null>(null)
const savingTemp  = ref(false)
const tempForm    = ref({ duration_value: 15, duration_unit: 'Minutes', reason: '' })

// Conflicts
const conflicts     = ref({ total: 0, errors: 0, warnings: 0, items: [] as any[] })
const showConflicts = ref(false)

// Form
const form = ref<Partial<CreateRulePayload & { address: string }>>({
  instance_id: '', rule_type: 'host', database: 'all', db_user: 'all',
  address: '', auth_method: 'md5', comment: '', rule_order: 0,
})

// Instances filtered by environment
const filteredInstances = computed(() => {
  if (!filters.value.environment) return instances.value
  return instances.value.filter(i => (i as any).host?.environment === filters.value.environment)
})

// Instance selected in the Add/Edit Rule modal
const selectedInstance = computed(() => instances.value.find(i => i.id === form.value.instance_id))

// Filtered rules
const filtered = computed(() => {
  const q = globalSearch.value.toLowerCase().trim()
  return rules.value.filter(r => {
    if (filters.value.environment) {
      const env = r.instance?.host?.environment
      if (env !== filters.value.environment) return false
    }
    if (filters.value.instance_id && r.instance_id !== filters.value.instance_id) return false
    if (filters.value.rule_type && r.rule_type !== filters.value.rule_type) return false
    if (filters.value.enabled !== '') {
      if (r.enabled !== (filters.value.enabled === 'true')) return false
    }
    if (q) {
      const inst = r.instance ? `${r.instance.instance_name}:${r.instance.port}` : ''
      const host = r.instance?.host ? `${r.instance.host.name} ${r.instance.host.ip_address}` : ''
      const haystack = [r.address, r.database, r.db_user, r.auth_method, r.rule_type, r.comment, inst, host].join(' ').toLowerCase()
      if (!haystack.includes(q)) return false
    }
    return true
  })
})

// Instance to target for Write & Reload — resolved from whatever's actually
// showing in the table right now (dropdown filter, free-text search, or
// both), not just the Instance dropdown. Otherwise narrowing to one
// instance purely via search (e.g. by CIDR) hides the button entirely, and
// a rule can sit unpushed on the server with no way to notice from here.
const singleVisibleInstance = computed(() => {
  const ids = new Set(filtered.value.map(r => r.instance_id))
  if (ids.size !== 1) return null
  const [id] = ids
  return instances.value.find(i => i.id === id) || null
})

// Sorted rules
const sorted = computed(() => {
  const arr = [...filtered.value]
  switch (sortBy.value) {
    case 'rule_order_asc':  return arr.sort((a, b) => a.rule_order - b.rule_order)
    case 'rule_order_desc': return arr.sort((a, b) => b.rule_order - a.rule_order)
    case 'disabled_first':  return arr.sort((a, b) => (a.enabled ? 1 : 0) - (b.enabled ? 1 : 0))
    case 'enabled_first':   return arr.sort((a, b) => (b.enabled ? 1 : 0) - (a.enabled ? 1 : 0))
    case 'address_asc':     return arr.sort((a, b) => (a.address||'').localeCompare(b.address||''))
    case 'address_desc':    return arr.sort((a, b) => (b.address||'').localeCompare(a.address||''))
    case 'created_desc':    return arr.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    case 'created_asc':     return arr.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())
    default: return arr
  }
})

const totalPages = computed(() => Math.max(1, Math.ceil(filtered.value.length / perPage.value)))
const paginated  = computed(() => sorted.value.slice((currentPage.value - 1) * perPage.value, currentPage.value * perPage.value))

// In drag mode show all filtered rules; otherwise show paginated
const displayRules = computed(() => dragMode.value ? dragRules.value : paginated.value)

const pageNumbers = computed(() => {
  const total = totalPages.value, cur = currentPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages: (number | string)[] = [1]
  if (cur > 3) pages.push('...')
  for (let i = Math.max(2, cur - 1); i <= Math.min(total - 1, cur + 1); i++) pages.push(i)
  if (cur < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

const previewTempExpiry = computed(() => {
  const val = tempForm.value.duration_value || 1
  const unit = tempForm.value.duration_unit
  const now = new Date()
  if (unit === 'Minutes') now.setMinutes(now.getMinutes() + val)
  else if (unit === 'Hours') now.setHours(now.getHours() + val)
  else if (unit === 'Days') now.setDate(now.getDate() + val)
  else if (unit === 'Weeks') now.setDate(now.getDate() + val * 7)
  return now.toLocaleString('id-ID')
})

function onFilterChange() { currentPage.value = 1 }

// ── Drag & Drop ────────────────────────────────────────────────
function toggleDragMode() {
  if (!filters.value.instance_id) {
    toast.warning('Select a specific instance first to reorder rules')
    return
  }
  dragMode.value = !dragMode.value
  if (dragMode.value) {
    dragRules.value = [...filtered.value].sort((a, b) => a.rule_order - b.rule_order)
    toast.info(`Drag rows to reorder — ${dragRules.value.length} rules loaded`)
  } else {
    dragRules.value = []
  }
  dragSrcIdx.value = null
  dragOverIdx.value = null
}

function onDragStart(e: DragEvent, idx: number) {
  dragSrcIdx.value = idx
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', String(idx))
  }
}

function onDragOver(e: DragEvent, idx: number) {
  e.preventDefault()
  if (dragSrcIdx.value === null || dragSrcIdx.value === idx) return
  dragOverIdx.value = idx
  // Live reorder preview
  const arr = [...dragRules.value]
  const [moved] = arr.splice(dragSrcIdx.value, 1)
  arr.splice(idx, 0, moved)
  dragRules.value = arr
  dragSrcIdx.value = idx
}

function onDragLeave() {
  dragOverIdx.value = null
}

async function onDrop(e: DragEvent, idx: number) {
  e.preventDefault()
  dragOverIdx.value = null
}

async function onDragEnd() {
  dragSrcIdx.value = null
  dragOverIdx.value = null
  await saveOrder()
}

async function saveOrder() {
  if (!dragRules.value.length) return
  savingOrder.value = true
  try {
    const payload = dragRules.value.map((r, i) => ({ id: r.id, order: i }))
    await api.post('/rules/reorder', { rules: payload })
    dragRules.value.forEach((r, i) => { r.rule_order = i })
    toast.success('Rule order saved — pg_hba.conf updated')
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Failed to save order')
  } finally {
    savingOrder.value = false
  }
}

// ── Other functions ────────────────────────────────────────────
async function checkConflicts() {
  try {
    const params: any = {}
    if (filters.value.instance_id) params.instance_id = filters.value.instance_id
    const { data } = await api.get('/rules/conflicts', { params })
    conflicts.value = { total: data.total || 0, errors: data.errors || 0, warnings: data.warnings || 0, items: data.conflicts || [] }
    showConflicts.value = true
    if (data.total === 0) { toast.success('No conflicts!'); showConflicts.value = false }
    else if (data.errors > 0) toast.error(`${data.errors} errors, ${data.warnings} warnings found`)
    else toast.warning(`${data.warnings} warnings found`)
  } catch { toast.error('Failed to check conflicts') }
}

async function doClone() {
  cloning.value = true
  try {
    const { data } = await api.post('/rules/clone', cloneForm.value)
    cloneResult.value = data
    toast.success(`Cloned ${data.cloned} rules`)
    await load()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Clone failed')
  } finally { cloning.value = false }
}

function openSetTemp(rule: HBARule) {
  tempTarget.value = rule
  tempForm.value = { duration_value: 15, duration_unit: 'Minutes', reason: '' }
}

async function doSetTemp() {
  if (!tempTarget.value) return
  savingTemp.value = true
  try {
    await api.post(`/rules/${tempTarget.value.id}/set-temporary`, tempForm.value)
    toast.success(`Rule set as temporary — expires in ${tempForm.value.duration_value} ${tempForm.value.duration_unit}`)
    tempTarget.value = null
    await load()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Failed to set temporary')
  } finally { savingTemp.value = false }
}

async function load() {
  loading.value = true
  try {
    const { data } = await rulesApi.list(filters.value.instance_id ? { instance_id: filters.value.instance_id } : undefined)
    rules.value = data
  } catch { toast.error('Failed to load rules') }
  finally { loading.value = false }
}

async function toggleRule(rule: HBARule) {
  try {
    if (rule.enabled) {
      await rulesApi.disable(rule.id)
      rule.enabled = false
    } else {
      await rulesApi.enable(rule.id)
      rule.enabled = true
    }
  } catch { toast.error('Failed to toggle rule'); await load() }
}

function openCreate() {
  editing.value = null
  form.value = { instance_id: instances.value[0]?.id || '', rule_type: 'host', database: 'all', db_user: 'all', address: '', auth_method: 'md5', comment: '', rule_order: 0 }
  showModal.value = true
}

function openEdit(rule: HBARule) {
  editing.value = rule
  form.value = { instance_id: rule.instance_id, rule_type: rule.rule_type, database: rule.database, db_user: rule.db_user, address: rule.address, auth_method: rule.auth_method, comment: rule.comment, rule_order: rule.rule_order }
  showModal.value = true
}

async function saveRule() {
  saving.value = true
  try {
    if (editing.value) {
      await rulesApi.update(editing.value.id, form.value as CreateRulePayload)
      toast.success('Rule updated')
    } else {
      await rulesApi.create(form.value as CreateRulePayload)
      toast.success('Rule added')
    }
    showModal.value = false
    await load()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Save failed')
  } finally { saving.value = false }
}

async function deleteRule(rule: HBARule) {
  if (!confirm(`Delete rule ${rule.address}?`)) return
  try {
    await rulesApi.delete(rule.id)
    toast.success('Rule deleted')
    await load()
  } catch { toast.error('Delete failed') }
}

async function reloadAll() {
  toast.info('Refreshing rule list...')
  await load()
  toast.success(`Loaded ${rules.value.length} rules`)
}

function openWriteReload() {
  if (!singleVisibleInstance.value) {
    toast.error('Narrow the list to a single instance first (Instance filter or search)')
    return
  }
  writeTarget.value = singleVisibleInstance.value
}

async function doWriteReload() {
  if (!writeTarget.value) return
  const inst = writeTarget.value
  writingReload.value = true
  try {
    await api.post(`/instances/${inst.id}/write-hba`)
    toast.success(`✓ pg_hba.conf updated & reloaded — ${inst.instance_name}:${inst.port}`)
    writeTarget.value = null
  } catch (e: any) {
    toast.error(e.response?.data?.detail || e.response?.data?.error || 'Write or reload failed')
  } finally {
    writingReload.value = false
  }
}

onMounted(async () => {
  const { data } = await instancesApi.list()
  instances.value = data
  await load()
})
</script>

<style scoped>
.search-inline { display:flex;align-items:center;gap:8px;background:var(--c-bg);border:1px solid var(--c-border);border-radius:6px;padding:6px 10px; }
.search-inline input { background:none;border:none;outline:none;color:var(--c-text);font-size:12px;font-family:var(--font);width:100%; }
.total-count { font-size:12px;color:var(--c-muted);font-weight:400;margin-left:8px; }
.rule-row { transition:background .1s; }
.rule-row:hover { background:var(--c-surface2); }
.row-disabled { opacity:.55; }
.row-temp td:first-child { border-left:2px solid var(--c-amber); }
.rule-type { font-family:monospace;background:var(--c-bg);padding:2px 6px;border-radius:4px;font-size:11px;color:var(--c-accent); }
.mono { font-family:monospace;font-size:12px; }
.method { font-family:monospace;font-size:11px;color:var(--c-teal); }
.warning-box { background:rgba(245,158,11,.08);border:1px solid rgba(245,158,11,.2);border-radius:8px;padding:14px;display:flex;gap:12px;color:var(--c-amber);font-size:12px; }
/* Drag & Drop */
.drag-handle-cell { width:36px;text-align:center;cursor:grab;color:var(--c-muted);user-select:none; }
.drag-handle-cell:active { cursor:grabbing; }
.drag-icon { font-size:16px;color:#8b5cf6; }
.drag-over { background:rgba(139,92,246,.15) !important;border-top:2px solid #8b5cf6 !important; }
.row-dragging { opacity:.3; }
tr[draggable="true"] { cursor:grab; }
tr[draggable="true"]:hover { background:rgba(139,92,246,.06); }
/* Pagination */
.pagination { display:flex;align-items:center;gap:4px;padding:12px 16px;border-top:1px solid var(--c-border);flex-wrap:wrap; }
.page-info { font-size:12px;color:var(--c-muted);margin-left:auto; }
.page-ellipsis { font-size:12px;color:var(--c-muted);padding:0 4px; }
.spin { animation:spin 1s linear infinite; }
@keyframes spin { to { transform:rotate(360deg); } }
</style>
