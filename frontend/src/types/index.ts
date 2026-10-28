// ── Enums ─────────────────────────────────────────────────────────────────────

export type UserRole   = 'SuperAdmin' | 'DBA' | 'Operator' | 'Viewer' | 'AccessManager' | 'TempAccessManager'
export type UserStatus = 'Active' | 'Inactive'
export type EnvType    = 'Development' | 'UAT' | 'Production'
export type InstanceStatus = 'Online' | 'Offline' | 'Unknown'
export type RuleType   = 'local' | 'host' | 'hostssl' | 'hostnossl'
export type AuthMethod = 'trust' | 'reject' | 'md5' | 'scram-sha-256' | 'password' | 'gss' | 'ident' | 'peer' | 'ldap' | 'cert'
export type TempDurationUnit = 'Minutes' | 'Hours' | 'Days' | 'Weeks' | 'Months'

// ── Auth ──────────────────────────────────────────────────────────────────────

export interface AuthUser {
  id: string
  username: string
  full_name: string
  email: string
  role: UserRole
  theme: string
  must_change_password?: boolean
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: AuthUser
}

// ── User ─────────────────────────────────────────────────────────────────────

export interface User {
  id: string
  username: string
  full_name: string
  email: string
  role: UserRole
  status: UserStatus
  theme: string
  last_login: string | null
  created_at: string
  updated_at: string
}

// ── DBHost ────────────────────────────────────────────────────────────────────

export interface DBHost {
  id: string
  name: string
  ip_address: string
  ssh_port: number
  ssh_user: string
  pg_version: string
  environment: EnvType
  active: boolean
  notes: string
  created_by: string | null
  created_at: string
  updated_at: string
  instances?: PGInstance[]
}

export interface CreateHostPayload {
  name: string
  ip_address: string
  ssh_port?: number
  ssh_user?: string
  ssh_key?: string
  pg_version: string
  environment: EnvType
  notes?: string
}

// ── PGInstance ────────────────────────────────────────────────────────────────

export interface PGInstance {
  id: string
  host_id: string
  instance_name: string
  pg_version: string
  port: number
  data_directory: string
  hba_path: string
  status: InstanceStatus
  last_scan: string | null
  created_at: string
  updated_at: string
  host?: DBHost
}

// ── HBARule ───────────────────────────────────────────────────────────────────

export interface HBARule {
  id: string
  instance_id: string
  rule_order: number
  rule_type: RuleType
  database: string
  db_user: string
  address: string
  auth_method: AuthMethod
  auth_options: string
  comment: string
  enabled: boolean
  is_temporary: boolean
  created_by: string | null
  created_at: string
  updated_at: string
  instance?: PGInstance
}

export interface CreateRulePayload {
  instance_id: string
  rule_type: RuleType
  database?: string
  db_user?: string
  address?: string
  auth_method: AuthMethod
  auth_options?: string
  comment?: string
  enabled?: boolean
  rule_order?: number
}

// ── TempAccess ────────────────────────────────────────────────────────────────

export interface TempAccess {
  id: string
  rule_id: string
  instance_id: string
  granted_by: string
  duration_value: number
  duration_unit: TempDurationUnit
  expires_at: string
  reason: string
  revoked: boolean
  revoked_by: string | null
  revoked_at: string | null
  created_at: string
  granted_by_user?: User
}

export interface GrantTempPayload {
  instance_id: string
  ip_address: string
  database?: string
  db_user?: string
  auth_method?: AuthMethod
  duration_value: number
  duration_unit: TempDurationUnit
  reason?: string
}

// ── Backup ────────────────────────────────────────────────────────────────────

export interface HBABackup {
  id: string
  instance_id: string
  filename: string
  file_path: string
  file_size: number
  trigger_event: string
  note: string
  created_by: string | null
  created_at: string
  instance?: PGInstance
}

// ── BackupSchedule ────────────────────────────────────────────────────────────

export type BackupTarget = 'local' | 's3' | 'both'
export type BackupScheduleScope = 'all' | 'host' | 'instance'

export interface BackupSchedule {
  id: string
  name: string
  cron_expression: string
  target: BackupTarget
  scope_type: BackupScheduleScope
  scope_id: string | null
  enabled: boolean
  created_by: string | null
  last_run_at: string | null
  last_run_status: string
  last_run_message: string
  created_at: string
  updated_at: string
}

export interface BackupSchedulePayload {
  name: string
  cron_expression: string
  target: BackupTarget
  scope_type: BackupScheduleScope
  scope_id?: string
  enabled?: boolean
}

// ── Audit ─────────────────────────────────────────────────────────────────────

export interface AuditLog {
  id: number
  user_id: string | null
  username: string
  event: string
  target: string
  instance_id: string | null
  instance_label: string
  client_ip: string
  detail: string
  created_at: string
}

export interface AuditListResponse {
  data: AuditLog[]
  total: number
  page: number
  limit: number
}

// ── Dashboard ─────────────────────────────────────────────────────────────────

export interface VersionDist {
  pg_version: string
  count: number
}

export interface DashboardStats {
  total_hosts: number
  total_instances: number
  online_instances: number
  offline_instances: number
  total_rules: number
  temp_rules_active: number
  expired_rules: number
  version_dist: VersionDist[]
  recent_activity: AuditLog[]
}
