package santa

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/backend/internal/models"
	"github.com/woodleighschool/grinch/backend/internal/store"
)

type Service struct {
	store *store.Store
}

func NewService(store *store.Store) *Service {
	return &Service{store: store}
}

// Preflight handles the initial sync request from Santa client
func (s *Service) Preflight(ctx context.Context, machineID string, req models.PreflightRequest) (*models.PreflightResponse, error) {
	if strings.TrimSpace(machineID) == "" {
		return nil, errors.New("machine_id required")
	}

	// Find primary user - create as local user if not found
	var primaryUserID *uuid.UUID
	if req.PrimaryUser != "" {
		user, err := s.store.UserByUsername(ctx, req.PrimaryUser)
		if err != nil {
			return nil, fmt.Errorf("lookup primary user: %w", err)
		}
		if user != nil {
			primaryUserID = &user.ID
		} else {
			// User not found, create as local user
			localUser, err := s.store.UpsertLocalUser(ctx, req.PrimaryUser, "", machineID)
			if err != nil {
				return nil, fmt.Errorf("create local user: %w", err)
			}
			primaryUserID = &localUser.ID
		}
	}

	// Update host information with Santa fields
	if err := s.store.UpsertHostWithMachineID(ctx, machineID, req, primaryUserID); err != nil {
		return nil, fmt.Errorf("update host info: %w", err)
	}

	// Get or create sync state
	syncState, err := s.store.GetSantaSyncState(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("get sync state: %w", err)
	}

	// Determine sync type
	syncType := models.SyncTypeNormal
	if syncState == nil {
		// First sync for this machine should be a clean sync
		syncType = models.SyncTypeClean
	} else if req.RequestCleanSync {
		// Client has requested a clean sync
		syncType = models.SyncTypeClean
	}

	// Check for rule drift (if hashes are provided)
	if syncState != nil && req.RuleCountHash != "" {
		if syncState.RuleCountHash != req.RuleCountHash {
			syncType = models.SyncTypeClean
		}
	}

	// Configure response
	enableBundles := true
	enableTransitive := false
	resp := &models.PreflightResponse{
		SyncType:              string(syncType),
		BatchSize:             50,
		EnableBundles:         &enableBundles,
		EnableTransitiveRules: &enableTransitive,
	}

	return resp, nil
}

// EventUpload handles event upload from Santa client
func (s *Service) EventUpload(ctx context.Context, machineID string, req models.EventUploadRequest) (*models.EventUploadResponse, error) {
	if strings.TrimSpace(machineID) == "" {
		return nil, errors.New("machine_id required")
	}

	// Process each event
	bundleHashes := make(map[string]struct{})
	for _, event := range req.Events {
		if err := s.store.InsertSantaEvent(ctx, machineID, event); err != nil {
			return nil, fmt.Errorf("insert event: %w", err)
		}

		// Collect bundle hashes for bundle binary uploads
		if event.FileBundleHash != "" && event.Decision == models.DecisionBundleBinary {
			bundleHashes[event.FileBundleHash] = struct{}{}
		}
	}

	// Return bundle hashes to request full bundle uploads
	var bundleBinaries []string
	for hash := range bundleHashes {
		bundleBinaries = append(bundleBinaries, hash)
	}

	return &models.EventUploadResponse{
		EventUploadBundleBinaries: bundleBinaries,
	}, nil
}

// RuleDownload handles rule download for Santa client with cursor pagination
func (s *Service) RuleDownload(ctx context.Context, machineID string, req models.RuleDownloadRequest) (*models.RuleDownloadResponse, error) {
	if strings.TrimSpace(machineID) == "" {
		return nil, errors.New("machine_id required")
	}

	const pageSize = 50
	var lastRuleID *uuid.UUID

	// Handle cursor for pagination
	if req.Cursor != "" {
		cursor, ruleID, err := s.store.GetSantaRuleCursor(ctx, req.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		if cursor != machineID {
			return nil, errors.New("cursor machine mismatch")
		}
		lastRuleID = ruleID
	}

	// Get rules for this machine
	rules, err := s.store.GetSantaRulesForMachine(ctx, machineID, lastRuleID, pageSize+1)
	if err != nil {
		return nil, fmt.Errorf("get rules: %w", err)
	}

	// Check if there are more rules
	var nextCursor string
	if len(rules) > pageSize {
		// Trim the extra record and create a cursor for the next page.
		rules = rules[:pageSize]

		if len(rules) > 0 {
			cursor, err := s.store.CreateSantaRuleCursor(ctx, machineID, nil)
			if err != nil {
				return nil, fmt.Errorf("create cursor: %w", err)
			}
			nextCursor = cursor
		}
	}

	// Record rule deliveries
	for _, rule := range rules {
		if err := s.store.RecordSantaRuleDelivery(ctx, machineID, rule); err != nil {
			// Log but don't fail the sync
			continue
		}
	}

	return &models.RuleDownloadResponse{
		Rules:  rules,
		Cursor: nextCursor,
	}, nil
}

// Postflight handles sync completion from Santa client
func (s *Service) Postflight(ctx context.Context, machineID string, req models.PostflightRequest) (*models.PostflightResponse, error) {
	if strings.TrimSpace(machineID) == "" {
		return nil, errors.New("machine_id required")
	}

	now := time.Now()

	// Get the current sync state to use the existing sync type if none provided
	currentState, err := s.store.GetSantaSyncState(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("get current sync state: %w", err)
	}

	// Use the provided sync type, or fall back to the current state, or default to NORMAL
	syncType := req.SyncType
	if syncType == "" {
		if currentState != nil && currentState.LastSyncType != "" {
			syncType = currentState.LastSyncType
		} else {
			syncType = models.SyncTypeNormal
		}
	}

	// Update sync state
	syncState := models.SyncState{
		MachineID:           machineID,
		LastSyncTime:        &now,
		LastSyncType:        syncType,
		RulesDelivered:      req.RulesReceived,
		RulesProcessed:      req.RulesProcessed,
		RuleCountHash:       req.RuleCountHash,
		BinaryRuleHash:      req.BinaryRuleHash,
		CertificateRuleHash: req.CertificateRuleHash,
		TeamIDRuleHash:      req.TeamIDRuleHash,
		SigningIDRuleHash:   req.SigningIDRuleHash,
		CDHashRuleHash:      req.CDHashRuleHash,
		TransitiveRuleHash:  req.TransitiveRuleHash,
		CompilerRuleHash:    req.CompilerRuleHash,
	}

	if _, err := s.store.UpsertSantaSyncState(ctx, machineID, syncState); err != nil {
		return nil, fmt.Errorf("update sync state: %w", err)
	}

	return &models.PostflightResponse{}, nil
}
