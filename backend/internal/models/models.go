package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Common errors
var (
	ErrApplicationExists   = errors.New("application with this identifier already exists")
	ErrApplicationNotFound = errors.New("application not found")
)

// SyncType represents the type of sync operation
type SyncType string

const (
	SyncTypeNormal   SyncType = "NORMAL"
	SyncTypeClean    SyncType = "CLEAN"
	SyncTypeCleanAll SyncType = "CLEAN_ALL"
)

// ClientMode represents Santa client operating mode
type ClientMode string

const (
	ClientModeLockdown ClientMode = "LOCKDOWN"
	ClientModeMonitor  ClientMode = "MONITOR"
)

// RuleType represents the type of rule
type RuleType string

const (
	RuleTypeBinary      RuleType = "BINARY"
	RuleTypeCertificate RuleType = "CERTIFICATE"
	RuleTypeSigningID   RuleType = "SIGNINGID"
	RuleTypeTeamID      RuleType = "TEAMID"
	RuleTypeCDHash      RuleType = "CDHASH"
)

// PolicyType represents the policy action for a rule
type PolicyType string

const (
	PolicyAllowlist         PolicyType = "ALLOWLIST"
	PolicyAllowlistCompiler PolicyType = "ALLOWLIST_COMPILER"
	PolicyBlocklist         PolicyType = "BLOCKLIST"
	PolicySilentBlocklist   PolicyType = "SILENT_BLOCKLIST"
	PolicyRemove            PolicyType = "REMOVE"
)

// Decision represents the execution decision
type Decision string

const (
	DecisionAllowBinary      Decision = "ALLOW_BINARY"
	DecisionAllowCertificate Decision = "ALLOW_CERTIFICATE"
	DecisionAllowScope       Decision = "ALLOW_SCOPE"
	DecisionAllowTeamID      Decision = "ALLOW_TEAMID"
	DecisionAllowSigningID   Decision = "ALLOW_SIGNINGID"
	DecisionAllowCDHash      Decision = "ALLOW_CDHASH"
	DecisionAllowUnknown     Decision = "ALLOW_UNKNOWN"
	DecisionBlockBinary      Decision = "BLOCK_BINARY"
	DecisionBlockCertificate Decision = "BLOCK_CERTIFICATE"
	DecisionBlockScope       Decision = "BLOCK_SCOPE"
	DecisionBlockTeamID      Decision = "BLOCK_TEAMID"
	DecisionBlockSigningID   Decision = "BLOCK_SIGNINGID"
	DecisionBlockCDHash      Decision = "BLOCK_CDHASH"
	DecisionBlockUnknown     Decision = "BLOCK_UNKNOWN"
	DecisionBundleBinary     Decision = "BUNDLE_BINARY"
)

// PreflightRequest represents the initial sync request from Santa client
type PreflightRequest struct {
	SerialNum            string     `json:"serial_num"`
	Hostname             string     `json:"hostname"`
	OSVersion            string     `json:"os_version"`
	OSBuild              string     `json:"os_build"`
	ModelIdentifier      string     `json:"model_identifier"`
	SantaVersion         string     `json:"santa_version"`
	PrimaryUser          string     `json:"primary_user"`
	ClientMode           ClientMode `json:"client_mode"`
	BinaryRuleCount      int        `json:"binary_rule_count"`
	CertificateRuleCount int        `json:"certificate_rule_count"`
	TeamIDRuleCount      int        `json:"teamid_rule_count"`
	SigningIDRuleCount   int        `json:"signingid_rule_count"`
	CDHashRuleCount      int        `json:"cdhash_rule_count"`
	TransitiveRuleCount  int        `json:"transitive_rule_count"`
	CompilerRuleCount    int        `json:"compiler_rule_count"`
	RequestCleanSync     bool       `json:"request_clean_sync,omitempty"`

	// Rule drift detection fields
	RuleCountHash       string `json:"rule_count_hash,omitempty"`
	BinaryRuleHash      string `json:"binary_rule_hash,omitempty"`
	CertificateRuleHash string `json:"certificate_rule_hash,omitempty"`
	TeamIDRuleHash      string `json:"teamid_rule_hash,omitempty"`
	SigningIDRuleHash   string `json:"signingid_rule_hash,omitempty"`
	CDHashRuleHash      string `json:"cdhash_rule_hash,omitempty"`
	TransitiveRuleHash  string `json:"transitive_rule_hash,omitempty"`
	CompilerRuleHash    string `json:"compiler_rule_hash,omitempty"`
}

// PreflightResponse configures client behavior and sync type
type PreflightResponse struct {
	ClientMode               ClientMode `json:"client_mode,omitempty"`
	AllowedPathRegex         *string    `json:"allowed_path_regex,omitempty"`
	BlockedPathRegex         *string    `json:"blocked_path_regex,omitempty"`
	BlockUSBMount            *bool      `json:"block_usb_mount,omitempty"`
	RemountUSBMode           *string    `json:"remount_usb_mode,omitempty"`
	SyncType                 string     `json:"sync_type,omitempty"`
	BatchSize                int        `json:"batch_size,omitempty"`
	EnableBundles            *bool      `json:"enable_bundles,omitempty"`
	EnableTransitiveRules    *bool      `json:"enable_transitive_rules,omitempty"`
	FullSyncInterval         *int       `json:"full_sync_interval,omitempty"`
	OverrideFileAccessAction *string    `json:"override_file_access_action,omitempty"`

	// Additional configuration fields
	FCMProject             string `json:"fcm_project,omitempty"`
	FCMEntity              string `json:"fcm_entity,omitempty"`
	FCMAPIKey              string `json:"fcm_api_key,omitempty"`
	BundleHashingBlockSize int    `json:"bundle_hashing_block_size,omitempty"`
}

// SigningChain represents a certificate in the signing chain
type SigningChain struct {
	SHA256     string `json:"sha256"`
	CN         string `json:"cn"`
	Org        string `json:"org"`
	OU         string `json:"ou"`
	ValidFrom  int64  `json:"valid_from"`
	ValidUntil int64  `json:"valid_until"`
}

// SantaEvent represents a single execution event
type SantaEvent struct {
	FileSHA256                  string         `json:"file_sha256"`
	FilePath                    string         `json:"file_path"`
	FileName                    string         `json:"file_name"`
	ExecutingUser               string         `json:"executing_user"`
	ExecutionTime               float64        `json:"execution_time"`
	LoggedInUsers               []string       `json:"loggedin_users"`
	CurrentSessions             []string       `json:"current_sessions"`
	Decision                    Decision       `json:"decision"`
	FileBundleID                string         `json:"file_bundle_id,omitempty"`
	FileBundlePath              string         `json:"file_bundle_path,omitempty"`
	FileBundleExecutableRelPath string         `json:"file_bundle_executable_rel_path,omitempty"`
	FileBundleName              string         `json:"file_bundle_name,omitempty"`
	FileBundleVersion           string         `json:"file_bundle_version,omitempty"`
	FileBundleVersionString     string         `json:"file_bundle_version_string,omitempty"`
	FileBundleHash              string         `json:"file_bundle_hash,omitempty"`
	FileBundleHashMillis        int            `json:"file_bundle_hash_millis,omitempty"`
	FileBundleBinaryCount       int            `json:"file_bundle_binary_count,omitempty"`
	PID                         int            `json:"pid,omitempty"`
	PPID                        int            `json:"ppid,omitempty"`
	ParentName                  string         `json:"parent_name,omitempty"`
	QuarantineDataURL           string         `json:"quarantine_data_url,omitempty"`
	QuarantineRefererURL        string         `json:"quarantine_referer_url,omitempty"`
	QuarantineTimestamp         int64          `json:"quarantine_timestamp,omitempty"`
	QuarantineAgentBundleID     string         `json:"quarantine_agent_bundle_id,omitempty"`
	SigningChain                []SigningChain `json:"signing_chain,omitempty"`
	SigningID                   string         `json:"signing_id,omitempty"`
	TeamID                      string         `json:"team_id,omitempty"`
	CDHash                      string         `json:"cdhash,omitempty"`
}

// EventUploadRequest contains batched events from client
type EventUploadRequest struct {
	Events []SantaEvent `json:"events"`
}

// EventUploadResponse tells client which bundle events to upload
type EventUploadResponse struct {
	EventUploadBundleBinaries []string `json:"event_upload_bundle_binaries,omitempty"`
}

// SantaRule represents a rule to be applied by the client
type SantaRule struct {
	Identifier            string     `json:"identifier"`
	RuleType              RuleType   `json:"rule_type"`
	Policy                PolicyType `json:"policy"`
	CustomMsg             string     `json:"custom_msg,omitempty"`
	CustomURL             string     `json:"custom_url,omitempty"`
	CreationTime          float64    `json:"creation_time,omitempty"`
	FileBundleBinaryCount int        `json:"file_bundle_binary_count,omitempty"`
	FileBundleHash        string     `json:"file_bundle_hash,omitempty"`
}

// RuleDownloadRequest contains cursor for pagination
type RuleDownloadRequest struct {
	Cursor string `json:"cursor,omitempty"`
}

// RuleDownloadResponse contains rules and pagination cursor
type RuleDownloadResponse struct {
	Rules  []SantaRule `json:"rules"`
	Cursor string      `json:"cursor,omitempty"`
}

// PostflightRequest contains sync completion information
type PostflightRequest struct {
	RulesReceived  int      `json:"rules_received"`
	RulesProcessed int      `json:"rules_processed"`
	SyncType       SyncType `json:"sync_type,omitempty"`

	// Rule drift detection fields
	RuleCountHash       string `json:"rule_count_hash,omitempty"`
	BinaryRuleHash      string `json:"binary_rule_hash,omitempty"`
	CertificateRuleHash string `json:"certificate_rule_hash,omitempty"`
	TeamIDRuleHash      string `json:"teamid_rule_hash,omitempty"`
	SigningIDRuleHash   string `json:"signingid_rule_hash,omitempty"`
	CDHashRuleHash      string `json:"cdhash_rule_hash,omitempty"`
	TransitiveRuleHash  string `json:"transitive_rule_hash,omitempty"`
	CompilerRuleHash    string `json:"compiler_rule_hash,omitempty"`
}

// PostflightResponse is typically empty
type PostflightResponse struct {
	// Usually empty - 200 OK is sufficient
}

// SyncState tracks the sync state for a machine
type SyncState struct {
	ID             uuid.UUID  `json:"id"`
	MachineID      string     `json:"machine_id"`
	LastSyncTime   *time.Time `json:"last_sync_time"`
	LastSyncType   SyncType   `json:"last_sync_type"`
	RulesDelivered int        `json:"rules_delivered"`
	RulesProcessed int        `json:"rules_processed"`

	// Rule drift tracking
	RuleCountHash       string `json:"rule_count_hash,omitempty"`
	BinaryRuleHash      string `json:"binary_rule_hash,omitempty"`
	CertificateRuleHash string `json:"certificate_rule_hash,omitempty"`
	TeamIDRuleHash      string `json:"teamid_rule_hash,omitempty"`
	SigningIDRuleHash   string `json:"signingid_rule_hash,omitempty"`
	CDHashRuleHash      string `json:"cdhash_rule_hash,omitempty"`
	TransitiveRuleHash  string `json:"transitive_rule_hash,omitempty"`
	CompilerRuleHash    string `json:"compiler_rule_hash,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserType represents the type of user account
type UserType string

const (
	UserTypeLocal UserType = "local"
	UserTypeCloud UserType = "cloud"
)

type User struct {
	ID               uuid.UUID  `json:"id"`
	ExternalID       *string    `json:"external_id,omitempty"` // Only cloud users have external IDs
	PrincipalName    string     `json:"principal_name"`
	DisplayName      string     `json:"display_name,omitempty"`
	Email            string     `json:"email,omitempty"`
	UserType         UserType   `json:"user_type"`
	PasswordHash     *string    `json:"-"` // Never expose password hash in JSON
	IsProtectedLocal bool       `json:"is_protected_local"`
	SyncedAt         *time.Time `json:"synced_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ApplicationSetting represents a configurable application setting
type ApplicationSetting struct {
	ID          uuid.UUID       `json:"id"`
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	Description string          `json:"description,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// LocalUserMetadata contains metadata specific to local users
type LocalUserMetadata struct {
	UserID              uuid.UUID  `json:"user_id"`
	FirstSeenAt         time.Time  `json:"first_seen_at"`
	SantaAgentMachineID string     `json:"santa_agent_machine_id,omitempty"`
	LastConvertedCheck  *time.Time `json:"last_converted_check,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type Group struct {
	ID          uuid.UUID  `json:"id"`
	ExternalID  string     `json:"external_id"`
	DisplayName string     `json:"display_name"`
	Description string     `json:"description,omitempty"`
	SyncedAt    *time.Time `json:"synced_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type GroupMembership struct {
	GroupID uuid.UUID `json:"group_id"`
	UserID  uuid.UUID `json:"user_id"`
}

type Application struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	RuleType    string    `json:"rule_type"`
	Identifier  string    `json:"identifier"`
	Description string    `json:"description,omitempty"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ApplicationScope struct {
	ID            uuid.UUID `json:"id"`
	ApplicationID uuid.UUID `json:"application_id"`
	TargetType    string    `json:"target_type"`
	TargetID      uuid.UUID `json:"target_id"`
	Action        string    `json:"action"`
	CreatedAt     time.Time `json:"created_at"`
}

type BlockedEvent struct {
	ID            int64      `json:"id"`
	HostID        *uuid.UUID `json:"host_id,omitempty"`
	UserID        *uuid.UUID `json:"user_id,omitempty"`
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	ProcessPath   string     `json:"process_path"`
	ProcessHash   string     `json:"process_hash,omitempty"`
	Signer        string     `json:"signer,omitempty"`
	BlockedReason string     `json:"blocked_reason,omitempty"`
	EventPayload  any        `json:"event_payload,omitempty"`
	OccurredAt    time.Time  `json:"occurred_at"`
	IngestedAt    time.Time  `json:"ingested_at"`
}

type Host struct {
	ID                     uuid.UUID  `json:"id"`
	Hostname               string     `json:"hostname"`
	SerialNumber           string     `json:"serial_number,omitempty"`
	MachineID              string     `json:"machine_id"`
	PrimaryUserID          *uuid.UUID `json:"primary_user_id,omitempty"`
	PrimaryUserPrincipal   string     `json:"primary_user_principal,omitempty"`
	PrimaryUserDisplayName string     `json:"primary_user_display_name,omitempty"`
	LastSeen               *time.Time `json:"last_seen,omitempty"`
	OSVersion              string     `json:"os_version,omitempty"`
	OSBuild                string     `json:"os_build,omitempty"`
	ModelIdentifier        string     `json:"model_identifier,omitempty"`
	SantaVersion           string     `json:"santa_version,omitempty"`
	ClientMode             ClientMode `json:"client_mode,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type UserPolicy struct {
	ScopeID         uuid.UUID `json:"scope_id"`
	ApplicationID   uuid.UUID `json:"application_id"`
	ApplicationName string    `json:"application_name"`
	RuleType        string    `json:"rule_type"`
	Identifier      string    `json:"identifier"`
	Action          string    `json:"action"`
	TargetType      string    `json:"target_type"`
	TargetID        uuid.UUID `json:"target_id"`
	TargetName      string    `json:"target_name,omitempty"`
	ViaGroup        bool      `json:"via_group"`
	CreatedAt       time.Time `json:"created_at"`
}

type UserEvent struct {
	ID            int64      `json:"id"`
	HostID        *uuid.UUID `json:"host_id,omitempty"`
	Hostname      string     `json:"hostname,omitempty"`
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	ProcessPath   string     `json:"process_path"`
	BlockedReason string     `json:"blocked_reason,omitempty"`
	Decision      string     `json:"decision,omitempty"`
	OccurredAt    time.Time  `json:"occurred_at"`
}
