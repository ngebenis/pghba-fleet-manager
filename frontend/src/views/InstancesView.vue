<template>
  <div class="page-content">
    <div class="toolbar">
      <select class="filter-select" v-model="filters.environment" @change="onFilterChange">
        <option value="">All Environments</option>
        <option>Production</option>
        <option>Staging</option>
        <option>UAT</option>
        <option>Development</option>
      </select>
      <select class="filter-select" v-model="filters.host_id" @change="onFilterChange">
        <option value="">All Hosts</option>
        <option v-for="h in filteredHosts" :key="h.id" :value="h.id">{{ h.name }} ({{ h.ip_address }})</option>
      </select>
      <select class="filter-select" v-model="filters.status" @change="onFilterChange">
        <option value="">All Status</option>
        <option>Online</option>
        <option>Offline</option>
        <option>Unknown</option>
      </select>
      <div class="toolbar-right">
        <button class="btn btn-ghost" @click="refreshAll">
          <i class="ti ti-refresh"></i> Refresh Status
        </button>
        <button class="btn btn-ghost" @click="discoverAll">
          <i class="ti ti-radar"></i> Auto-Discover
        </button>
        <button class="btn btn-primary" @click="showBulkImport = true"
          style="background:rgba(79,142,247,.15);border-color:var(--c-accent);color:var(--c-accent)">
          <i class="ti ti-download"></i> Bulk Import
        </button>
        <button class="btn btn-primary" @click="showBulkApply = true"
          style="background:rgba(245,158,11,.15);border-color:var(--c-amber);color:var(--c-amber)">
          <i class="ti ti-upload"></i> Bulk Apply
        </button>
      </div>
    </div>

    <!-- Summary cards -->
    <div class="inst-summary">
      <div class="sum-card">
        <div class="sum-num green">{{ onlineCount }}</div>
        <div class="sum-label">Online</div>
      </div>
      <div class="sum-card">
        <div class="sum-num red">{{ offlineCount }}</div>
        <div class="sum-label">Offline</div>
      </div>
      <div class="sum-card">
        <div class="sum-num amber">{{ unknownCount }}</div>
        <div class="sum-label">Unknown</div>
      </div>
      <div class="sum-card">
        <div class="sum-num accent">{{ filtered.length }}</div>
        <div class="sum-label">Total</div>
      </div>
    </div>

    <div class="table-wrap">
      <!-- Table header with per-page selector -->
      <div class="table-header">
        <span class="table-title">
          PostgreSQL Instances
          <span class="total-count">{{ filtered.length }} instances</span>
        </span>
        <div style="display:flex;align-items:center;gap:8px">
          <span style="font-size:12px;color:var(--c-muted)">Per page:</span>
          <select class="filter-select" v-model="perPage" @change="currentPage = 1" style="width:70px">
            <option :value="10">10</option>
            <option :value="20">20</option>
            <option :value="30">30</option>
            <option :value="50">50</option>
            <option :value="100">100</option>
          </select>
        </div>
      </div>

      <table>
        <thead>
          <tr>
            <th>Instance</th>
            <th>Host</th>
            <th>Port</th>
            <th>PG Ver</th>
            <th>pg_hba.conf Path</th>
            <th>Last Scan</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="8" style="text-align:center;padding:24px;color:var(--c-muted)">
              <i class="ti ti-loader-2 spin"></i> Loading...
            </td>
          </tr>
          <tr v-for="inst in paginated" :key="inst.id">
            <td style="font-weight:600">{{ inst.instance_name }}</td>
            <td>
              <span class="host-chip">
                <i class="ti ti-server" style="font-size:12px"></i>
                {{ inst.host?.name ?? '—' }}
                <span v-if="inst.host?.ip_address" style="color:var(--c-muted);font-size:11px;font-family:monospace">({{ inst.host.ip_address }})</span>
              </span>
            </td>
            <td><span class="port-badge">:{{ inst.port }}</span></td>
            <td><span class="badge badge-gray">PG {{ inst.pg_version }}</span></td>
            <td class="path-cell" :title="inst.hba_path">{{ inst.hba_path }}</td>
            <td style="font-size:11px;color:var(--c-muted);white-space:nowrap">
              {{ inst.last_scan ? fmt(inst.last_scan) : 'Never' }}
            </td>
            <td>
              <span class="dot" :class="statusDot(inst.status)"></span>{{ inst.status }}
            </td>
            <td>
              <div class="action-group">
                <button class="btn btn-ghost btn-sm" @click="viewRules(inst)" title="View Rules">
                  <i class="ti ti-shield-lock"></i>
                </button>
                <button class="btn btn-ghost btn-sm" @click="scanInstance(inst)"
                  :disabled="scanning === inst.id" title="Scan Status">
                  <i :class="['ti', scanning === inst.id ? 'ti-loader-2 spin' : 'ti-radar']"></i>
                </button>
                <button class="btn btn-ghost btn-sm" @click="importHBA(inst)"
                  :disabled="importing === inst.id" title="Import pg_hba.conf"
                  style="color:var(--c-accent)">
                  <i :class="['ti', importing === inst.id ? 'ti-loader-2 spin' : 'ti-download']"></i>
                  Import
                </button>
                <button class="btn btn-ghost btn-sm" @click="confirmWriteReload(inst)"
                  :disabled="writing === inst.id" title="Write & Reload PostgreSQL"
                  style="color:var(--c-amber)">
                  <i :class="['ti', writing === inst.id ? 'ti-loader-2 spin' : 'ti-upload']"></i>
                  Apply
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && filtered.length === 0">
            <td colspan="8" style="text-align:center;padding:32px;color:var(--c-muted)">
              No instances found.
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <div class="pagination" v-if="totalPages > 1">
        <button class="btn btn-ghost btn-sm" @click="currentPage = 1" :disabled="currentPage === 1">
          <i class="ti ti-chevrons-left"></i>
        </button>
        <button class="btn btn-ghost btn-sm" @click="currentPage--" :disabled="currentPage === 1">
          <i class="ti ti-chevron-left"></i>
        </button>

        <template v-for="p in pageNumbers" :key="p">
          <span v-if="p === '...'" class="page-ellipsis">...</span>
          <button v-else
            :class="['btn', 'btn-sm', p === currentPage ? 'btn-primary' : 'btn-ghost']"
            @click="currentPage = p as number">
            {{ p }}
          </button>
        </template>

        <button class="btn btn-ghost btn-sm" @click="currentPage++" :disabled="currentPage === totalPages">
          <i class="ti ti-chevron-right"></i>
        </button>
        <button class="btn btn-ghost btn-sm" @click="currentPage = totalPages" :disabled="currentPage === totalPages">
          <i class="ti ti-chevrons-right"></i>
        </button>

        <span class="page-info">
          {{ (currentPage - 1) * perPage + 1 }}–{{ Math.min(currentPage * perPage, filtered.length) }}
          of {{ filtered.length }}
        </span>
      </div>
    </div>

    <!-- Import Result Modal -->
    <Teleport to="body">
      <div v-if="importResult" class="modal-backdrop" @click.self="importResult = null">
        <div class="modal" style="width:560px">
          <div class="modal-header">
            <div class="modal-title">
              <i class="ti ti-download" style="color:var(--c-accent)"></i>
              Import — {{ importResult.instance }}:{{ importResult.port }}
            </div>
            <button class="btn btn-ghost btn-sm" @click="importResult = null"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body">
            <div class="import-stats">
              <div class="stat-pill green"><i class="ti ti-check"></i> {{ importResult.imported }} imported</div>
              <div class="stat-pill gray"><i class="ti ti-minus"></i> {{ importResult.skipped }} skipped</div>
            </div>
            <div style="margin-top:10px;font-size:11px;color:var(--c-muted)">
              PATH: <span style="font-family:monospace;color:var(--c-text)">{{ importResult.path }}</span>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="importResult = null">Close</button>
            <button class="btn btn-primary" @click="goToRules(importResult.instance_id)">
              <i class="ti ti-shield-lock"></i> View Rules
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Write & Reload + Diff Modal -->
    <Teleport to="body">
      <div v-if="writeTarget" class="modal-backdrop" @click.self="writeTarget = null">
        <div class="modal" style="width:780px;max-height:85vh;display:flex;flex-direction:column">
          <div class="modal-header">
            <div class="modal-title">
              <i class="ti ti-upload" style="color:var(--c-amber)"></i>
              Apply — {{ writeTarget.instance_name }}:{{ writeTarget.port }}
              <span style="font-size:11px;color:var(--c-muted);font-weight:400;margin-left:6px">
                {{ writeTarget.hba_path }}
              </span>
            </div>
            <button class="btn btn-ghost btn-sm" @click="writeTarget = null; diffData = null"><i class="ti ti-x"></i></button>
          </div>

          <!-- Loading diff -->
          <div v-if="loadingDiff" style="padding:40px;text-align:center;color:var(--c-muted)">
            <i class="ti ti-loader-2 spin" style="font-size:24px;display:block;margin-bottom:8px"></i>
            Comparing pg_hba.conf with database...
          </div>

          <!-- Diff error -->
          <div v-else-if="diffError" class="modal-body">
            <div style="background:rgba(239,68,68,.1);border:1px solid rgba(239,68,68,.2);border-radius:8px;padding:14px;color:var(--c-red);font-size:12px">
              <i class="ti ti-alert-circle"></i> {{ diffError }}
            </div>
            <div style="margin-top:12px">
              <div class="rules-count-box">
                <span class="rules-count">{{ ruleCount }}</span> rules
                <span style="color:var(--c-muted);font-size:11px;margin-left:6px">({{ enabledCount }} enabled, {{ disabledCount }} disabled)</span>
              </div>
            </div>
          </div>

          <!-- Diff View -->
          <div v-else-if="diffData" style="display:flex;flex-direction:column;flex:1;overflow:hidden">
            <!-- Diff stats -->
            <div style="display:flex;gap:8px;padding:10px 16px;border-bottom:1px solid var(--c-border);background:var(--c-bg);flex-shrink:0">
              <div class="diff-stat-pill green">
                <i class="ti ti-plus"></i> {{ diffData.stats.added }} added
              </div>
              <div class="diff-stat-pill red">
                <i class="ti ti-minus"></i> {{ diffData.stats.removed }} removed
              </div>
              <div class="diff-stat-pill gray">
                <i class="ti ti-equal"></i> {{ diffData.stats.unchanged }} unchanged
              </div>
              <div v-if="!diffData.stats.changed" style="margin-left:auto;font-size:12px;color:var(--c-green);display:flex;align-items:center;gap:4px">
                <i class="ti ti-circle-check"></i> No changes — file already up to date
              </div>
              <div v-else style="margin-left:auto;font-size:12px;color:var(--c-amber);display:flex;align-items:center;gap:4px">
                <i class="ti ti-alert-triangle"></i> {{ diffData.stats.added + diffData.stats.removed }} lines will change
              </div>
            </div>

            <!-- Diff tabs -->
            <div style="display:flex;gap:0;border-bottom:1px solid var(--c-border);flex-shrink:0">
              <button :class="['diff-tab', diffView === 'unified' ? 'active' : '']" @click="diffView = 'unified'">
                <i class="ti ti-layout-rows"></i> Unified Diff
              </button>
              <button :class="['diff-tab', diffView === 'split' ? 'active' : '']" @click="diffView = 'split'">
                <i class="ti ti-columns"></i> Split View
              </button>
            </div>

            <!-- Unified diff -->
            <div v-if="diffView === 'unified'" class="diff-body">
              <div v-for="(line, i) in diffData.diff" :key="i"
                :class="['diff-line', `diff-${line.type}`]">
                <span class="diff-gutter">{{ line.line_num }}</span>
                <span class="diff-sign">{{ line.type === 'added' ? '+' : line.type === 'removed' ? '-' : ' ' }}</span>
                <span class="diff-content">{{ line.content }}</span>
              </div>
            </div>

            <!-- Split diff -->
            <div v-if="diffView === 'split'" class="diff-split-body">
              <div class="diff-split-pane">
                <div class="diff-pane-header">Current (server)</div>
                <div class="diff-pane-content">
                  <pre class="diff-pre">{{ diffData.current }}</pre>
                </div>
              </div>
              <div class="diff-split-divider"></div>
              <div class="diff-split-pane">
                <div class="diff-pane-header">New (from DB)</div>
                <div class="diff-pane-content">
                  <pre class="diff-pre">{{ diffData.new }}</pre>
                </div>
              </div>
            </div>
          </div>

          <div class="modal-footer" style="flex-shrink:0">
            <button class="btn btn-ghost" @click="writeTarget = null; diffData = null">Cancel</button>
            <button v-if="diffData && !diffData.stats.changed" class="btn btn-ghost" disabled
              style="color:var(--c-green)">
              <i class="ti ti-circle-check"></i> Already up to date
            </button>
            <button v-else class="btn btn-primary" @click="doWriteReload" :disabled="writing === writeTarget?.id"
              style="background:var(--c-amber);border-color:var(--c-amber)">
              <i :class="['ti', writing === writeTarget?.id ? 'ti-loader-2 spin' : 'ti-upload']"></i>
              Write & Reload
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>

    
    <!-- Bulk Apply Modal -->
    <Teleport to="body">
      <div v-if="showBulkApply" class="modal-backdrop" @click.self="showBulkApply = false; bulkApplyResult = null">
        <div class="modal" style="width:620px;max-height:80vh;display:flex;flex-direction:column">
          <div class="modal-header">
            <div class="modal-title">
              <i class="ti ti-upload" style="color:var(--c-amber)"></i>
              Bulk Apply — Write & Reload pg_hba.conf
            </div>
            <button class="btn btn-ghost btn-sm" @click="showBulkApply = false; bulkApplyResult = null">
              <i class="ti ti-x"></i>
            </button>
          </div>
          <div class="modal-body" style="overflow-y:auto;flex:1">
            <div v-if="!bulkApplyResult">
              <div class="warning-box" style="margin-bottom:14px">
                <i class="ti ti-alert-triangle"></i>
                <div>
                  Tindakan ini akan menulis pg_hba.conf ke <strong>semua instances</strong> yang dipilih
                  dan me-reload PostgreSQL. Instance yang belum pernah di-<strong>Import</strong> (rules-nya
                  di database masih kosong atau tidak lengkap) akan otomatis ditolak demi keamanan,
                  bukan ditimpa diam-diam — jalankan <strong>Import</strong> pada instance tersebut
                  dulu satu per satu kalau ada yang gagal, lalu Bulk Apply lagi.
                </div>
              </div>

              <div class="field">
                <label>Apply To</label>
                <div style="display:flex;gap:8px;flex-wrap:wrap">
                  <label class="radio-opt" :class="{ active: bulkApplyFilter.type === 'host' }">
                    <input type="radio" v-model="bulkApplyFilter.type" value="host"> By Host
                  </label>
                  <label class="radio-opt" :class="{ active: bulkApplyFilter.type === 'environment' }">
                    <input type="radio" v-model="bulkApplyFilter.type" value="environment"> By Environment
                  </label>
                </div>
              </div>

              <div v-if="bulkApplyFilter.type === 'host'" class="field">
                <label>Select Host</label>
                <select v-model="bulkApplyFilter.host_id" style="width:100%">
                  <option value="" disabled>Choose a host...</option>
                  <option v-for="h in hosts" :key="h.id" :value="h.id">
                    {{ h.name }} ({{ h.ip_address }})
                  </option>
                </select>
                <div style="font-size:11px;color:var(--c-muted);margin-top:4px">
                  {{ instanceCountForHost(bulkApplyFilter.host_id) }} instances will be updated
                </div>
              </div>

              <div v-if="bulkApplyFilter.type === 'environment'" class="field">
                <label>Select Environment</label>
                <select v-model="bulkApplyFilter.environment" style="width:100%">
                  <option value="" disabled>Choose environment...</option>
                  <option>Production</option>
                  <option>Staging</option>
                  <option>UAT</option>
                  <option>Development</option>
                </select>
                <div style="font-size:11px;color:var(--c-muted);margin-top:4px">
                  {{ instanceCountForEnv(bulkApplyFilter.environment) }} instances will be updated
                </div>
              </div>
            </div>

            <!-- Result -->
            <div v-if="bulkApplyResult">
              <div class="import-stats" style="margin-bottom:14px">
                <div class="stat-pill green"><i class="ti ti-check"></i> {{ bulkApplyResult.success }} success</div>
                <div class="stat-pill red" v-if="bulkApplyResult.failed > 0">
                  <i class="ti ti-x"></i> {{ bulkApplyResult.failed }} failed
                </div>
                <div class="stat-pill gray"><i class="ti ti-database"></i> {{ bulkApplyResult.total }} total</div>
              </div>

              <table style="width:100%;font-size:11px;border-collapse:collapse">
                <thead>
                  <tr>
                    <th style="text-align:left;padding:4px 8px;color:var(--c-muted);background:var(--c-bg)">Instance</th>
                    <th style="text-align:left;padding:4px 8px;color:var(--c-muted);background:var(--c-bg)">Host</th>
                    <th style="text-align:center;padding:4px 8px;color:var(--c-muted);background:var(--c-bg)">Status</th>
                    <th style="text-align:left;padding:4px 8px;color:var(--c-muted);background:var(--c-bg)">Error</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="r in bulkApplyResult.results" :key="r.instance+r.port"
                    style="border-top:1px solid var(--c-border)">
                    <td style="padding:5px 8px;font-weight:500">{{ r.instance }}:{{ r.port }}</td>
                    <td style="padding:5px 8px;color:var(--c-muted)">{{ r.host }}</td>
                    <td style="padding:5px 8px;text-align:center">
                      <span :class="['badge', r.status === 'success' ? 'badge-green' : 'badge-red']">
                        {{ r.status }}
                      </span>
                    </td>
                    <td style="padding:5px 8px;color:var(--c-red);font-size:10px">{{ r.error || '' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div class="modal-footer" style="flex-shrink:0">
            <button class="btn btn-ghost" @click="showBulkApply = false; bulkApplyResult = null">
              {{ bulkApplyResult ? 'Close' : 'Cancel' }}
            </button>
            <button v-if="!bulkApplyResult" class="btn btn-primary" @click="doBulkApply"
              :disabled="bulkApplying || (!bulkApplyFilter.host_id && !bulkApplyFilter.environment)"
              style="background:var(--c-amber);border-color:var(--c-amber)">
              <i :class="['ti', bulkApplying ? 'ti-loader-2 spin' : 'ti-upload']"></i>
              {{ bulkApplying ? 'Applying...' : 'Bulk Apply' }}
            </button>
            <button v-if="bulkApplyResult" class="btn btn-ghost" @click="bulkApplyResult = null">
              <i class="ti ti-refresh"></i> Apply Again
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Bulk Import Modal -->
    <Teleport to="body">
      <div v-if="showBulkImport" class="modal-backdrop" @click.self="showBulkImport = false">
        <div class="modal" style="width:600px;max-height:80vh;display:flex;flex-direction:column">
          <div class="modal-header">
            <div class="modal-title">
              <i class="ti ti-download" style="color:var(--c-accent)"></i>
              Bulk Import pg_hba.conf
            </div>
            <button class="btn btn-ghost btn-sm" @click="showBulkImport = false"><i class="ti ti-x"></i></button>
          </div>
          <div class="modal-body" style="overflow-y:auto;flex:1">
            <div v-if="!bulkResult">
              <div style="background:var(--c-bg);border:1px solid var(--c-border);border-radius:8px;padding:14px;margin-bottom:14px;font-size:12px;color:var(--c-muted)">
                <i class="ti ti-info-circle" style="color:var(--c-accent)"></i>
                Import pg_hba.conf dari semua <strong style="color:var(--c-text)">{{ filtered.length }} instances</strong> yang sedang ditampilkan.
                Rules yang sudah ada akan di-skip (tidak duplikat).
              </div>
              <div class="field">
                <label>Filter Host (opsional)</label>
                <select v-model="bulkHostFilter" class="filter-select" style="width:100%">
                  <option value="">All Hosts ({{ filtered.length }} instances)</option>
                  <option v-for="h in hosts" :key="h.id" :value="h.id">
                    {{ h.name }} ({{ h.ip_address }})
                  </option>
                </select>
              </div>
            </div>

            <!-- Result -->
            <div v-if="bulkResult">
              <div class="import-stats" style="margin-bottom:14px">
                <div class="stat-pill green"><i class="ti ti-check"></i> {{ bulkResult.total_imported }} rules imported</div>
                <div class="stat-pill gray"><i class="ti ti-minus"></i> {{ bulkResult.total_skipped }} skipped</div>
                <div class="stat-pill" :class="bulkResult.total_errors > 0 ? 'red' : 'gray'">
                  <i class="ti ti-alert-circle"></i> {{ bulkResult.total_errors }} errors
                </div>
              </div>

              <table style="width:100%;font-size:11px;border-collapse:collapse">
                <thead>
                  <tr>
                    <th style="text-align:left;padding:4px 8px;color:var(--c-muted);background:var(--c-bg)">Instance</th>
                    <th style="text-align:left;padding:4px 8px;color:var(--c-muted);background:var(--c-bg)">Host</th>
                    <th style="text-align:center;padding:4px 8px;color:var(--c-muted);background:var(--c-bg)">Imported</th>
                    <th style="text-align:center;padding:4px 8px;color:var(--c-muted);background:var(--c-bg)">Skipped</th>
                    <th style="text-align:left;padding:4px 8px;color:var(--c-muted);background:var(--c-bg)">Status</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="r in bulkResult.results" :key="r.instance+r.port"
                    style="border-top:1px solid var(--c-border)"
                    :style="{ opacity: r.status === 'skipped' ? 0.5 : 1 }">
                    <td style="padding:5px 8px;font-weight:500">{{ r.instance }}:{{ r.port }}</td>
                    <td style="padding:5px 8px;color:var(--c-muted)">{{ r.host }}</td>
                    <td style="padding:5px 8px;text-align:center;color:var(--c-green)">
                      {{ r.status === 'success' ? r.imported : '—' }}
                    </td>
                    <td style="padding:5px 8px;text-align:center;color:var(--c-muted)">
                      {{ r.status === 'success' ? r.skipped : '—' }}
                    </td>
                    <td style="padding:5px 8px">
                      <span :class="['badge', r.status === 'success' ? 'badge-green' : r.status === 'error' ? 'badge-red' : 'badge-gray']">
                        {{ r.status }}
                      </span>
                      <span v-if="r.error" style="color:var(--c-red);font-size:10px;margin-left:4px">{{ r.error }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="showBulkImport = false; bulkResult = null">
              {{ bulkResult ? 'Close' : 'Cancel' }}
            </button>
            <button v-if="!bulkResult" class="btn btn-primary" @click="doBulkImport" :disabled="bulkImporting">
              <i :class="['ti', bulkImporting ? 'ti-loader-2 spin' : 'ti-download']"></i>
              {{ bulkImporting ? 'Importing...' : 'Start Bulk Import' }}
            </button>
            <button v-if="bulkResult" class="btn btn-ghost" @click="bulkResult = null">
              <i class="ti ti-refresh"></i> Import Again
            </button>
          </div>
        </div>
      </div>
    </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from 'vue-toastification'
import { instancesApi, hostsApi } from '@/api'
import api from '@/api'
import { globalSearch } from '@/stores/search'
import type { PGInstance, DBHost, InstanceStatus } from '@/types'
import dayjs from 'dayjs'

const router = useRouter()
const toast  = useToast()

const instances = ref<PGInstance[]>([])
const hosts     = ref<DBHost[]>([])
const loading   = ref(false)
const scanning  = ref<string | null>(null)
const importing = ref<string | null>(null)
const writing    = ref<string | null>(null)
const loadingDiff = ref(false)
const diffData    = ref<any>(null)
const diffView    = ref<'unified' | 'split' | 'new'>('unified')
const showBulkImport = ref(false)
const showBulkApply  = ref(false)
const bulkApplying   = ref(false)
const bulkApplyResult = ref<any>(null)
const bulkApplyFilter = ref({ type: 'host', host_id: '', environment: '' })
const bulkImporting = ref(false)
const bulkResult = ref<any>(null)
const bulkHostFilter = ref('')
const filters   = ref({ environment: '', host_id: '', status: '' })
const importResult = ref<any>(null)
const writeTarget  = ref<PGInstance | null>(null)
const ruleCount    = ref(0)
const enabledCount  = ref(0)
const disabledCount = ref(0)

// Pagination
const perPage     = ref(20)
const currentPage = ref(1)

// Hosts filtered by environment
const filteredHosts = computed(() => {
  if (!filters.value.environment) return hosts.value
  return hosts.value.filter(h => (h as any).environment === filters.value.environment)
})

// Split diff lines
const currentLines = computed(() => {
  if (!diffData.value) return []
  return diffData.value.current.split('\n').map((l: string, i: number) => ({
    content: l,
    removed: diffData.value.diff.some((d: any) => d.type === 'removed' && d.content === l)
  }))
})

const newLines = computed(() => {
  if (!diffData.value) return []
  return diffData.value.new.split('\n').map((l: string) => ({
    content: l,
    added: diffData.value.diff.some((d: any) => d.type === 'added' && d.content === l)
  }))
})

// Filtered list
const filtered = computed(() => {
  const q = globalSearch.value.toLowerCase().trim()
  return instances.value.filter(i => {
    if (filters.value.host_id && i.host_id !== filters.value.host_id) return false
    if (filters.value.status && i.status !== filters.value.status) return false
    if (q) {
      const haystack = [
        i.instance_name, i.hba_path, i.pg_version,
        i.host?.name, i.host?.ip_address, String(i.port)
      ].join(' ').toLowerCase()
      if (!haystack.includes(q)) return false
    }
    return true
  })
})

const totalPages = computed(() => Math.max(1, Math.ceil(filtered.value.length / perPage.value)))

const paginated = computed(() => {
  const start = (currentPage.value - 1) * perPage.value
  return filtered.value.slice(start, start + perPage.value)
})

const pageNumbers = computed(() => {
  const total = totalPages.value
  const cur   = currentPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages: (number | string)[] = [1]
  if (cur > 3) pages.push('...')
  for (let i = Math.max(2, cur - 1); i <= Math.min(total - 1, cur + 1); i++) pages.push(i)
  if (cur < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

function onFilterChange() { currentPage.value = 1 }

const onlineCount  = computed(() => instances.value.filter(i => i.status === 'Online').length)
const offlineCount = computed(() => instances.value.filter(i => i.status === 'Offline').length)
const unknownCount = computed(() => instances.value.filter(i => i.status === 'Unknown').length)

function statusDot(s: InstanceStatus) {
  return s === 'Online' ? 'dot-green' : s === 'Offline' ? 'dot-red' : 'dot-amber'
}
function fmt(d: string) { return dayjs(d).format('DD MMM YYYY HH:mm') }

async function load() {
  loading.value = true
  try {
    const { data } = await instancesApi.list()
    instances.value = data
  } catch { toast.error('Failed to load instances') }
  finally { loading.value = false }
}

function viewRules(inst: PGInstance) {
  router.push({ name: 'rules', query: { instance_id: inst.id } })
}
function goToRules(instanceId: string) {
  importResult.value = null
  router.push({ name: 'rules', query: { instance_id: instanceId } })
}

async function scanInstance(inst: PGInstance) {
  scanning.value = inst.id
  try {
    const { data } = await api.post(`/instances/${inst.id}/scan`)
    inst.status    = data.status
    inst.last_scan = data.scanned_at
    toast.success(`${inst.instance_name} — ${data.status}`)
  } catch { toast.error('Scan failed') }
  finally { scanning.value = null }
}

async function importHBA(inst: PGInstance) {
  importing.value = inst.id
  try {
    const { data } = await api.post(`/instances/${inst.id}/import-hba`)
    importResult.value = { ...data, instance_id: inst.id }
    if (data.imported > 0) toast.success(`Imported ${data.imported} rules from ${inst.instance_name}`)
    else toast.info(`No new rules (${data.skipped} already exist)`)
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Import failed')
  } finally { importing.value = null }
}

async function confirmWriteReload(inst: PGInstance) {
  writeTarget.value = inst
  diffData.value    = null
  diffError.value   = ''
  diffView.value    = 'unified'

  // Load rule counts
  try {
    const { data } = await api.get('/rules', { params: { instance_id: inst.id } })
    ruleCount.value     = data.length
    enabledCount.value  = data.filter((r: any) => r.enabled).length
    disabledCount.value = data.filter((r: any) => !r.enabled).length
  } catch { ruleCount.value = 0 }

  // Load diff preview
  loadingDiff.value = true
  try {
    const { data } = await api.get(`/instances/${inst.id}/preview-diff`)
    diffData.value = data
  } catch (e: any) {
    diffError.value = e.response?.data?.error || 'Failed to load diff'
  } finally {
    loadingDiff.value = false
  }
}

async function doWriteReload() {
  if (!writeTarget.value) return
  const inst = writeTarget.value
  writing.value = inst.id
  writeTarget.value = null
  diffData.value = null
  try {
    await api.post(`/instances/${inst.id}/write-hba`)
    toast.success(`✓ pg_hba.conf updated & reloaded — ${inst.instance_name}:${inst.port}`)
  } catch (e: any) {
    toast.error(e.response?.data?.detail || 'Write or reload failed')
  } finally { writing.value = null }
}

async function refreshAll() {
  toast.info('Scanning all instances...')
  for (const inst of instances.value) {
    try {
      const { data } = await api.post(`/instances/${inst.id}/scan`)
      inst.status    = data.status
      inst.last_scan = data.scanned_at
    } catch {}
  }
  toast.success('Status refreshed')
}

function discoverAll() {
  toast.info('Use Discover Instances button on each host card')
}

async function doBulkImport() {
  bulkImporting.value = true
  try {
    const payload = bulkHostFilter.value ? { host_id: bulkHostFilter.value } : {}
    const { data } = await api.post('/instances/bulk-import', payload)
    bulkResult.value = data
    if (data.total_imported > 0) {
      toast.success(`Bulk import: ${data.total_imported} rules imported from ${data.total_instances} instances`)
    } else {
      toast.info(`Bulk import: all rules already exist (${data.total_skipped} skipped)`)
    }
    await load()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Bulk import failed')
  } finally {
    bulkImporting.value = false
  }
}

function instanceCountForHost(hostId: string): number {
  return instances.value.filter(i => i.host_id === hostId).length
}

function instanceCountForEnv(env: string): number {
  return instances.value.filter(i => (i as any).host?.environment === env).length
}

async function doBulkApply() {
  bulkApplying.value = true
  try {
    const payload: any = {}
    if (bulkApplyFilter.value.type === 'host') {
      payload.host_id = bulkApplyFilter.value.host_id
    } else {
      payload.environment = bulkApplyFilter.value.environment
    }
    const { data } = await api.post('/instances/bulk-apply', payload)
    bulkApplyResult.value = data
    if (data.failed === 0) {
      toast.success(`Bulk apply: ${data.success} instances updated successfully`)
    } else {
      toast.warning(`Bulk apply: ${data.success} success, ${data.failed} failed`)
    }
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Bulk apply failed')
  } finally {
    bulkApplying.value = false
  }
}

onMounted(async () => {
  const { data } = await hostsApi.list()
  hosts.value = data
  await load()
})
</script>

<style scoped>
.inst-summary { display: flex; gap: 12px; margin-bottom: 16px; }
.sum-card { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: 8px; padding: 12px 20px; text-align: center; min-width: 80px; }
.sum-num { font-size: 24px; font-weight: 600; line-height: 1; }
.sum-num.green  { color: var(--c-green); }
.sum-num.red    { color: var(--c-red); }
.sum-num.amber  { color: var(--c-amber); }
.sum-num.accent { color: var(--c-accent); }
.sum-label { font-size: 11px; color: var(--c-muted); margin-top: 4px; text-transform: uppercase; letter-spacing: .3px; }
.host-chip { display: inline-flex; align-items: center; gap: 4px; font-size: 12px; color: var(--c-muted); }
.port-badge { font-family: monospace; background: var(--c-bg); border: 1px solid var(--c-border); border-radius: 4px; padding: 2px 6px; font-size: 11px; color: var(--c-accent); }
.path-cell { font-family: monospace; font-size: 11px; color: var(--c-muted); max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.action-group { display: flex; gap: 4px; align-items: center; }
.total-count { font-size: 12px; color: var(--c-muted); font-weight: 400; margin-left: 8px; }
.pagination { display: flex; align-items: center; gap: 4px; padding: 12px 16px; border-top: 1px solid var(--c-border); flex-wrap: wrap; }
.page-info { font-size: 12px; color: var(--c-muted); margin-left: auto; }
.page-ellipsis { font-size: 12px; color: var(--c-muted); padding: 0 4px; }
.import-stats { display: flex; gap: 10px; }
.stat-pill { display: flex; align-items: center; gap: 6px; padding: 6px 12px; border-radius: 6px; font-size: 12px; font-weight: 500; }
.stat-pill.green { background: rgba(34,197,94,.1); color: var(--c-green); border: 1px solid rgba(34,197,94,.2); }
.stat-pill.gray  { background: rgba(255,255,255,.04); color: var(--c-muted); border: 1px solid var(--c-border); }
.warning-box { background: rgba(245,158,11,.08); border: 1px solid rgba(245,158,11,.25); border-radius: 8px; padding: 12px 14px; display: flex; gap: 12px; color: var(--c-amber); font-size: 12px; }
.rules-count-box { background: var(--c-bg); border: 1px solid var(--c-border); border-radius: 6px; padding: 10px 14px; display: flex; align-items: center; }
.rules-count { font-size: 20px; font-weight: 600; color: var(--c-accent); margin-right: 6px; }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
