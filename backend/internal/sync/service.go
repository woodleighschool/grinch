package sync

import (
	"context"
	"log/slog"
	"time"

	"github.com/woodleighschool/grinch/backend/internal/entra"
	"github.com/woodleighschool/grinch/backend/internal/store"
)

// PeriodicSyncService handles periodic synchronization tasks
type PeriodicSyncService struct {
	store        *store.Store
	entraService *entra.Service
	logger       *slog.Logger
}

// NewPeriodicSyncService creates a new periodic sync service
func NewPeriodicSyncService(store *store.Store, entraService *entra.Service, logger *slog.Logger) *PeriodicSyncService {
	if logger == nil {
		logger = slog.Default()
	}
	return &PeriodicSyncService{
		store:        store,
		entraService: entraService,
		logger:       logger.With("component", "periodic_sync"),
	}
}

// Start begins the periodic sync operations
func (s *PeriodicSyncService) Start(ctx context.Context) {
	s.logger.Info("periodic sync loop started")
	// Run cleanup every 6 hours
	cleanupTicker := time.NewTicker(6 * time.Hour)
	defer cleanupTicker.Stop()

	// Run local->cloud conversion check every hour
	conversionTicker := time.NewTicker(1 * time.Hour)
	defer conversionTicker.Stop()

	// Run initial cleanup and conversion check
	go s.runCleanup(ctx)
	go s.runLocalToCloudConversion(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Debug("context cancelled, stopping periodic sync loop")
			return
		case <-cleanupTicker.C:
			s.logger.Debug("cleanup ticker fired")
			go s.runCleanup(ctx)
		case <-conversionTicker.C:
			s.logger.Debug("conversion ticker fired")
			go s.runLocalToCloudConversion(ctx)
		}
	}
}

// runCleanup performs cleanup operations for orphaned local users
func (s *PeriodicSyncService) runCleanup(ctx context.Context) {
	s.logger.Debug("starting periodic cleanup")

	// Call the database function to cleanup orphaned local users
	const q = `SELECT cleanup_orphaned_local_users();`
	tag, err := s.store.Pool().Exec(ctx, q)
	if err != nil {
		s.logger.Error("cleanup failed", "error", err)
		return
	}

	s.logger.Info("cleanup completed successfully", "removed_users", tag.RowsAffected())
}

// runLocalToCloudConversion checks for local users that should be converted to cloud users
func (s *PeriodicSyncService) runLocalToCloudConversion(ctx context.Context) {
	s.logger.Debug("starting local-to-cloud conversion check")

	// Get all local users that haven't been checked recently
	localUsers, err := s.getLocalUsersForConversionCheck(ctx)
	if err != nil {
		s.logger.Error("failed to get local users for conversion", "error", err)
		return
	}

	if len(localUsers) == 0 {
		s.logger.Debug("no local users need conversion checking")
		return
	}

	s.logger.Info("checking local users for conversion", "count", len(localUsers))

	// This would require an Entra service token, but since it's periodic,
	// we'll just trigger a full Entra sync which handles conversions
	if err := s.entraService.Sync(ctx); err != nil {
		s.logger.Error("entra sync for conversion failed", "error", err)
	} else {
		s.logger.Info("entra sync for conversion completed")
	}
}

// getLocalUsersForConversionCheck returns local users that need to be checked for conversion
func (s *PeriodicSyncService) getLocalUsersForConversionCheck(ctx context.Context) ([]string, error) {
	const q = `
		UPDATE local_user_metadata 
		SET last_converted_check = NOW(),
		    updated_at = NOW()
		WHERE last_converted_check IS NULL 
		   OR last_converted_check < NOW() - INTERVAL '24 hours'
		RETURNING (SELECT principal_name FROM users WHERE id = user_id);
	`

	rows, err := s.store.Pool().Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var principals []string
	for rows.Next() {
		// Collect principals that need conversion checks
		var principal string
		if err := rows.Scan(&principal); err != nil {
			return nil, err
		}
		principals = append(principals, principal)
	}

	s.logger.Debug("fetched local users for conversion check", "count", len(principals))

	return principals, rows.Err()
}

// CleanupOrphanedUsers manually triggers cleanup of orphaned local users
func (s *PeriodicSyncService) CleanupOrphanedUsers(ctx context.Context) error {
	s.logger.Debug("manual cleanup trigger invoked")
	s.runCleanup(ctx)
	return nil
}

// TriggerLocalToCloudConversion manually triggers local to cloud conversion check
func (s *PeriodicSyncService) TriggerLocalToCloudConversion(ctx context.Context) error {
	s.logger.Debug("manual conversion trigger invoked")
	s.runLocalToCloudConversion(ctx)
	return nil
}
