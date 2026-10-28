package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func newUUID() uuid.UUID { return uuid.New() }

type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *Base) BeforeCreate(*gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = newUUID()
	}
	return nil
}

// ── User ─────────────────────────────────────────────────────────────────────

type UserRole   string
type UserStatus string

const (
	RoleSuperAdmin UserRole = "SuperAdmin"
	RoleDBA        UserRole = "DBA"
	RoleOperator   UserRole = "Operator"
	RoleViewer     UserRole = "Viewer"
	// RoleAccessManager can fully manage Access Rules and Temporary Access
	// (including write-back to the server) and view the Dashboard, but has
	// no access to Hosts, Instances, Backups, Users, Settings, or Audit Log.
	RoleAccessManager UserRole = "AccessManager"
	// RoleTempAccessManager can only grant/revoke Temporary Access and view
	// the Dashboard — narrower than RoleAccessManager, it has no access to
	// Access Rules (add/edit/delete/reorder/clone) at all.
	RoleTempAccessManager UserRole = "TempAccessManager"

	StatusActive   UserStatus = "Active"
	StatusInactive UserStatus = "Inactive"
)

type User struct {
	Base
	Username  string     `gorm:"uniqueIndex;size:64"      json:"username"`
	FullName  string     `gorm:"size:128"                  json:"full_name"`
	Email     string     `gorm:"uniqueIndex;size:256"      json:"email"`
	Password  string     `gorm:"size:256"                  json:"-"`
	Role      UserRole   `gorm:"type:varchar(16)"          json:"role"`
	Status    UserStatus `gorm:"type:varchar(16)"          json:"status"`
	Theme     string     `gorm:"size:16;default:'dark'"    json:"theme"`
	LastLogin *time.Time                                    `json:"last_login"`
	MustChangePassword bool `gorm:"column:must_change_password;default:false" json:"must_change_password"`
}

// ── RefreshToken ─────────────────────────────────────────────────────────────

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index"      json:"user_id"`
	TokenHash string    `gorm:"uniqueIndex"`
	ExpiresAt time.Time
	Revoked   bool      `gorm:"default:false"`
	CreatedAt time.Time
}

func (r *RefreshToken) BeforeCreate(*gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = newUUID()
	}
	return nil
}

// ── DBHost ────────────────────────────────────────────────────────────────────

type EnvType string

const (
	EnvDev  EnvType = "Development"
	EnvUAT  EnvType = "UAT"
	EnvProd EnvType = "Production"
)

type DBHost struct {
	Base
	Name        string    `gorm:"uniqueIndex;size:128"  json:"name"`
	IPAddress   string    `gorm:"size:45"               json:"ip_address"`
	SSHPort     int       `gorm:"default:22"            json:"ssh_port"`
	SSHUser     string    `gorm:"size:64;default:'postgres'" json:"ssh_user"`
	SSHKey      string    `gorm:"type:text"             json:"-"`           // encrypted
	PGVersion   string    `gorm:"size:8"                json:"pg_version"`
	Environment EnvType   `gorm:"type:varchar(16)"      json:"environment"`
	Active      bool      `gorm:"default:true"          json:"active"`
	Notes       string    `gorm:"type:text"             json:"notes"`
	CreatedByID *uuid.UUID `gorm:"type:uuid"            json:"created_by"`
	// SSHHostKeyFingerprint pins this host's SSH key by SHA256 fingerprint
	// after the first successful connection (see services.ConnectToHost) —
	// not a secret, just like a known_hosts entry, so it's fine to expose.
	SSHHostKeyFingerprint string `gorm:"column:ssh_host_key_fingerprint" json:"ssh_host_key_fingerprint,omitempty"`

	Instances []PGInstance `gorm:"foreignKey:HostID"    json:"instances,omitempty"`
}

// ── PGInstance ────────────────────────────────────────────────────────────────

type InstanceStatus string

const (
	StatusOnline  InstanceStatus = "Online"
	StatusOffline InstanceStatus = "Offline"
	StatusUnknown InstanceStatus = "Unknown"
)

type PGInstance struct {
	Base
	HostID        uuid.UUID      `gorm:"type:uuid;index"       json:"host_id"`
	InstanceName  string         `gorm:"size:128"              json:"instance_name"`
	PGVersion     string         `gorm:"size:8"                json:"pg_version"`
	Port          int            `json:"port"`
	DataDirectory string         `gorm:"type:text"             json:"data_directory"`
	HBAPath       string         `gorm:"type:text"             json:"hba_path"`
	Status        InstanceStatus `gorm:"type:varchar(16)"      json:"status"`
	LastScan      *time.Time     `json:"last_scan"`

	Host  DBHost    `gorm:"foreignKey:HostID"         json:"host,omitempty"`
	Rules []HBARule `gorm:"foreignKey:InstanceID"     json:"rules,omitempty"`
}

// ── HBARule ───────────────────────────────────────────────────────────────────

type RuleType   string
type AuthMethod string

const (
	TypeLocal     RuleType = "local"
	TypeHost      RuleType = "host"
	TypeHostSSL   RuleType = "hostssl"
	TypeHostNoSSL RuleType = "hostnossl"

	MethodTrust       AuthMethod = "trust"
	MethodReject      AuthMethod = "reject"
	MethodMD5         AuthMethod = "md5"
	MethodScram       AuthMethod = "scram-sha-256"
	MethodPassword    AuthMethod = "password"
)

type HBARule struct {
	Base
	InstanceID  uuid.UUID  `gorm:"type:uuid;index"       json:"instance_id"`
	RuleOrder   int        `gorm:"default:0"             json:"rule_order"`
	RuleType    RuleType   `gorm:"type:varchar(16)"      json:"rule_type"`
	Database    string     `gorm:"size:256;default:'all'" json:"database"`
	DBUser      string     `gorm:"size:256;default:'all'" json:"db_user"`
	Address     string     `gorm:"size:64"               json:"address"`
	AuthMethod  AuthMethod `gorm:"type:varchar(32)"      json:"auth_method"`
	AuthOptions string     `gorm:"size:256"              json:"auth_options"`
	Comment     string     `gorm:"size:512"              json:"comment"`
	Enabled     bool       `gorm:"default:true"          json:"enabled"`
	IsTemporary bool       `gorm:"default:false"         json:"is_temporary"`
	// CreatedForTempAccess is true only when this rule was created brand-new by the
	// Grant Temp Access flow (as opposed to a pre-existing rule later marked temporary
	// via Set Temporary). Used to decide whether it's safe to hard-delete on expiry.
	CreatedForTempAccess bool       `gorm:"default:false"         json:"created_for_temp_access"`
	CreatedByID          *uuid.UUID `gorm:"type:uuid"             json:"created_by"`

	Instance PGInstance  `gorm:"foreignKey:InstanceID" json:"instance,omitempty"`
	TempData *TempAccess `gorm:"foreignKey:RuleID"     json:"temp_data,omitempty"`
}

// ToHBALine renders the rule as a pg_hba.conf line
func (r *HBARule) ToHBALine() string {
	prefix := ""
	if !r.Enabled {
		prefix = "# "
	}
	if r.RuleType == TypeLocal {
		return prefix + fmt.Sprintf("local\t%s\t%s\t\t%s", r.Database, r.DBUser, r.AuthMethod)
	}
	return prefix + fmt.Sprintf("%s\t%s\t%s\t%s\t%s", r.RuleType, r.Database, r.DBUser, r.Address, r.AuthMethod)
}

// ── TempAccess ────────────────────────────────────────────────────────────────

type TempDurationUnit string

const (
	DurMinutes TempDurationUnit = "Minutes"
	DurHours   TempDurationUnit = "Hours"
	DurDays    TempDurationUnit = "Days"
	DurWeeks   TempDurationUnit = "Weeks"
	DurMonths  TempDurationUnit = "Months"
)

type TempAccess struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	RuleID        uuid.UUID        `gorm:"type:uuid;uniqueIndex" json:"rule_id"`
	InstanceID    uuid.UUID        `gorm:"type:uuid;index"       json:"instance_id"`
	GrantedByID   uuid.UUID        `gorm:"type:uuid"             json:"granted_by"`
	DurationValue int              `json:"duration_value"`
	DurationUnit  TempDurationUnit `gorm:"type:varchar(16)"      json:"duration_unit"`
	ExpiresAt     time.Time        `gorm:"index"                 json:"expires_at"`
	Reason        string           `gorm:"type:text"             json:"reason"`
	Revoked       bool             `gorm:"default:false"         json:"revoked"`
	RevokedByID   *uuid.UUID       `gorm:"type:uuid"             json:"revoked_by"`
	RevokedAt     *time.Time       `json:"revoked_at"`
	CreatedAt     time.Time        `json:"created_at"`

	GrantedBy User    `gorm:"foreignKey:GrantedByID" json:"granted_by_user,omitempty"`
	Rule      *HBARule `gorm:"foreignKey:RuleID"     json:"rule,omitempty"`
}

func (TempAccess) TableName() string { return "temp_access" }

func (t *TempAccess) BeforeCreate(*gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = newUUID()
	}
	return nil
}

// ── HBABackup ─────────────────────────────────────────────────────────────────

type HBABackup struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	InstanceID  uuid.UUID  `gorm:"type:uuid;index"      json:"instance_id"`
	Filename    string     `gorm:"size:256"             json:"filename"`
	FilePath    string     `gorm:"type:text"            json:"file_path"`
	FileSize    int64      `json:"file_size"`
	Content     string     `gorm:"type:text"            json:"-"`
	TriggerEvent string    `gorm:"size:64"              json:"trigger_event"`
	Note        string     `gorm:"size:512"             json:"note"`
	CreatedByID *uuid.UUID `gorm:"type:uuid"            json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`

	Instance PGInstance `gorm:"foreignKey:InstanceID" json:"instance,omitempty"`
}

func (b *HBABackup) BeforeCreate(*gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = newUUID()
	}
	return nil
}

// ── BackupSchedule ────────────────────────────────────────────────────────────

type BackupSchedule struct {
	Base
	Name           string     `gorm:"size:128"                json:"name"`
	CronExpression string     `gorm:"size:64"                 json:"cron_expression"`
	Target         string     `gorm:"size:16;default:'local'" json:"target"`    // local | s3 | both
	ScopeType      string     `gorm:"size:16;default:'all'"   json:"scope_type"` // all | host | instance
	ScopeID        *uuid.UUID `gorm:"type:uuid"                json:"scope_id"`
	Enabled        bool       `gorm:"default:true"             json:"enabled"`
	CreatedByID    *uuid.UUID `gorm:"type:uuid"                json:"created_by"`
	LastRunAt      *time.Time `json:"last_run_at"`
	LastRunStatus  string     `gorm:"size:16"                  json:"last_run_status"`
	LastRunMessage string     `gorm:"size:256"                 json:"last_run_message"`
}

func (BackupSchedule) TableName() string { return "backup_schedules" }

// ── AuditLog ──────────────────────────────────────────────────────────────────

type AuditEvent string

const (
	EventLogin             AuditEvent = "Login"
	EventLogout            AuditEvent = "Logout"
	EventLoginFailed       AuditEvent = "LoginFailed"
	EventAddHost           AuditEvent = "AddHost"
	EventEditHost          AuditEvent = "EditHost"
	EventDeleteHost        AuditEvent = "DeleteHost"
	EventAddInstance       AuditEvent = "AddInstance"
	EventAddRule           AuditEvent = "AddRule"
	EventEditRule          AuditEvent = "EditRule"
	EventDeleteRule        AuditEvent = "DeleteRule"
	EventEnableRule        AuditEvent = "EnableRule"
	EventDisableRule       AuditEvent = "DisableRule"
	EventGrantTempAccess   AuditEvent = "GrantTempAccess"
	EventRevokeTempAccess  AuditEvent = "RevokeTempAccess"
	EventExpireTempAccess  AuditEvent = "ExpireTempAccess"
	EventCreateBackup      AuditEvent = "CreateBackup"
	EventRestoreBackup     AuditEvent = "RestoreBackup"
	EventAddUser           AuditEvent = "AddUser"
	EventEditUser          AuditEvent = "EditUser"
	EventDeleteUser        AuditEvent = "DeleteUser"
	EventReloadConf        AuditEvent = "ReloadConf"
	EventAddBackupSchedule    AuditEvent = "AddBackupSchedule"
	EventEditBackupSchedule   AuditEvent = "EditBackupSchedule"
	EventDeleteBackupSchedule AuditEvent = "DeleteBackupSchedule"
)

type AuditLog struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        *uuid.UUID `gorm:"type:uuid;index"          json:"user_id"`
	Username      string     `gorm:"size:64"                  json:"username"`
	Event         AuditEvent `gorm:"type:varchar(32);index"   json:"event"`
	Target        string     `gorm:"type:text"                json:"target"`
	InstanceID    *uuid.UUID `gorm:"type:uuid;index"          json:"instance_id"`
	InstanceLabel string     `gorm:"size:256"                 json:"instance_label"`
	ClientIP      string     `gorm:"size:45"                  json:"client_ip"`
	Detail        string     `gorm:"type:jsonb"               json:"detail"`
	CreatedAt     time.Time  `gorm:"index"                    json:"created_at"`
}

// HBADiff — stores before/after content for pg_hba.conf changes
type HBADiff struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AuditLogID uint64     `gorm:"index"                    json:"audit_log_id"`
	InstanceID *uuid.UUID `gorm:"type:uuid;index"          json:"instance_id"`
	BeforeContent string  `gorm:"type:text"                json:"before_content"`
	AfterContent  string  `gorm:"type:text"                json:"after_content"`
	DiffLines     string  `gorm:"type:text"                json:"diff_lines"` // JSON array of diff lines
	CreatedAt  time.Time  `gorm:"index"                    json:"created_at"`
}

func (HBADiff) TableName() string { return "hba_diffs" }
