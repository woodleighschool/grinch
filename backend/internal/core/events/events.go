package events

import (
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
)

// Certificate describes a signing certificate in a signing chain.
type Certificate struct {
	SHA256     string     `json:"sha256"`
	CN         string     `json:"cn"`
	Org        string     `json:"org"`
	OU         string     `json:"ou"`
	ValidFrom  *time.Time `json:"valid_from"`
	ValidUntil *time.Time `json:"valid_until"`
}

// Entitlement represents a single entitlement key:value pair.
type Entitlement struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Event represents a single Santa execution event.
type Event struct {
	ID        uuid.UUID       `json:"id"`
	MachineID uuid.UUID       `json:"machine_id"`
	Decision  syncv1.Decision `json:"decision"`

	FileSha256 string `json:"file_sha256"`
	FilePath   string `json:"file_path"`
	FileName   string `json:"file_name"`

	ExecutingUser string     `json:"executing_user"`
	ExecutionTime *time.Time `json:"execution_time"`

	LoggedInUsers   []string `json:"logged_in_users"`
	CurrentSessions []string `json:"current_sessions"`

	FileBundleID                string `json:"file_bundle_id"`
	FileBundlePath              string `json:"file_bundle_path"`
	FileBundleExecutableRelPath string `json:"file_bundle_executable_rel_path"`
	FileBundleName              string `json:"file_bundle_name"`
	FileBundleVersion           string `json:"file_bundle_version"`
	FileBundleVersionString     string `json:"file_bundle_version_string"`
	FileBundleHash              string `json:"file_bundle_hash"`
	FileBundleHashMillis        int32  `json:"file_bundle_hash_millis"`
	FileBundleBinaryCount       int32  `json:"file_bundle_binary_count"`

	Pid        int32  `json:"pid"`
	Ppid       int32  `json:"ppid"`
	ParentName string `json:"parent_name"`
	TeamID     string `json:"team_id"`
	SigningID  string `json:"signing_id"`
	Cdhash     string `json:"cdhash"`
	CsFlags    int32  `json:"cs_flags"`

	SigningStatus     syncv1.SigningStatus `json:"signing_status"`
	SigningChain      []Certificate        `json:"signing_chain"`
	Entitlements      []Entitlement        `json:"entitlements"`
	SecureSigningTime *time.Time           `json:"secure_signing_time"`
	SigningTime       *time.Time           `json:"signing_time"`
}

// EventListItem provides a reduced view of an event for list queries.
type EventListItem struct {
	ID            uuid.UUID       `json:"id"`
	MachineID     uuid.UUID       `json:"machine_id"`
	Decision      syncv1.Decision `json:"decision"`
	ExecutionTime *time.Time      `json:"execution_time"`
	FilePath      string          `json:"file_path"`
	FileSha256    string          `json:"file_sha256"`
	FileName      string          `json:"file_name"`
	SigningID     string          `json:"signing_id"`
	TeamID        string          `json:"team_id"`
	Cdhash        string          `json:"cdhash"`
}
