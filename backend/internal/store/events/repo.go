// Package events provides persistence operations for event data.
package events

import (
	"context"
	"fmt"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woodleighschool/grinch/internal/domain/errx"
	"github.com/woodleighschool/grinch/internal/domain/events"
	"github.com/woodleighschool/grinch/internal/listing"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
	"github.com/woodleighschool/grinch/internal/store/db/sqlc"
)

// Repo persists events and related signing metadata.
type Repo struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo backed by the provided connection pool.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: sqlc.New(pool), pool: pool}
}

// Get returns an event by ID, including signing chain and entitlements.
func (r *Repo) Get(ctx context.Context, id uuid.UUID) (events.Event, error) {
	row, err := r.q.GetEventByID(ctx, id)
	if err != nil {
		return events.Event{}, errx.FromStore(err, nil)
	}

	ev := mapEvent(row)

	chain, err := r.q.ListSigningChainEntriesByEventID(ctx, id)
	if err != nil {
		return events.Event{}, errx.FromStore(err, nil)
	}
	ev.SigningChain = mapSigningChain(chain)

	ents, err := r.q.ListEntitlementsByEventID(ctx, id)
	if err != nil {
		return events.Event{}, errx.FromStore(err, nil)
	}
	ev.Entitlements = mapEntitlements(ents)

	return ev, nil
}

// List returns events matching the query.
func (r *Repo) List(ctx context.Context, query listing.Query) ([]events.ListItem, listing.Page, error) {
	items, total, err := listEvents(ctx, r.pool, query)
	if err != nil {
		return nil, listing.Page{}, errx.FromStore(err, nil)
	}
	return items, listing.Page{Total: total}, nil
}

// InsertBatch inserts events and their related signing metadata in a single transaction.
func (r *Repo) InsertBatch(ctx context.Context, items []events.Event) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return errx.FromStore(err, nil)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := r.q.WithTx(tx)
	var row sqlc.Event

	for _, ev := range items {
		row, err = qtx.CreateEvent(ctx, toCreateParams(ev))
		if err != nil {
			return errx.FromStore(err, nil)
		}

		if err = saveSigningChain(ctx, qtx, row.ID, ev.SigningChain); err != nil {
			return fmt.Errorf("save signing chain: %w", err)
		}

		if err = saveEntitlements(ctx, qtx, row.ID, ev.Entitlements); err != nil {
			return fmt.Errorf("save entitlements: %w", err)
		}
	}

	return errx.FromStore(tx.Commit(ctx), nil)
}

func saveSigningChain(ctx context.Context, q *sqlc.Queries, eventID uuid.UUID, chain []events.Certificate) error {
	for _, cert := range chain {
		if err := q.UpsertCertificate(ctx, sqlc.UpsertCertificateParams{
			Sha256:     cert.SHA256,
			Cn:         cert.CN,
			Org:        cert.Org,
			Ou:         cert.OU,
			ValidFrom:  pgconv.TimeOrNull(cert.ValidFrom),
			ValidUntil: pgconv.TimeOrNull(cert.ValidUntil),
		}); err != nil {
			return errx.FromStore(err, nil)
		}
	}

	for i, cert := range chain {
		if err := q.CreateSigningChainEntry(ctx, sqlc.CreateSigningChainEntryParams{
			EventID:           eventID,
			Ordinal:           pgconv.IntToInt32(i),
			CertificateSha256: cert.SHA256,
		}); err != nil {
			return errx.FromStore(err, nil)
		}
	}

	return nil
}

func saveEntitlements(ctx context.Context, q *sqlc.Queries, eventID uuid.UUID, ents []events.Entitlement) error {
	for i, ent := range ents {
		entID, err := q.UpsertEntitlement(ctx, sqlc.UpsertEntitlementParams{
			Key:   ent.Key,
			Value: ent.Value,
		})
		if err != nil {
			return errx.FromStore(err, nil)
		}

		if err = q.CreateEventEntitlement(ctx, sqlc.CreateEventEntitlementParams{
			EventID:       eventID,
			Ordinal:       pgconv.IntToInt32(i),
			EntitlementID: entID,
		}); err != nil {
			return errx.FromStore(err, nil)
		}
	}

	return nil
}

func mapEvent(row sqlc.Event) events.Event {
	return events.Event{
		ID:                          row.ID,
		MachineID:                   row.MachineID,
		Decision:                    syncv1.Decision(row.Decision),
		FilePath:                    row.FilePath,
		FileSha256:                  row.FileSha256,
		FileName:                    row.FileName,
		ExecutingUser:               row.ExecutingUser,
		ExecutionTime:               pgconv.TimeVal(row.ExecutionTime),
		LoggedInUsers:               pgconv.TextArray(row.LoggedInUsers),
		CurrentSessions:             pgconv.TextArray(row.CurrentSessions),
		FileBundleID:                row.FileBundleID,
		FileBundlePath:              row.FileBundlePath,
		FileBundleExecutableRelPath: row.FileBundleExecutableRelPath,
		FileBundleName:              row.FileBundleName,
		FileBundleVersion:           row.FileBundleVersion,
		FileBundleVersionString:     row.FileBundleVersionString,
		FileBundleHash:              row.FileBundleHash,
		FileBundleHashMillis:        row.FileBundleHashMillis,
		FileBundleBinaryCount:       row.FileBundleBinaryCount,
		Pid:                         row.Pid,
		Ppid:                        row.Ppid,
		ParentName:                  row.ParentName,
		TeamID:                      row.TeamID,
		SigningID:                   row.SigningID,
		Cdhash:                      row.Cdhash,
		CsFlags:                     row.CsFlags,
		SigningStatus:               syncv1.SigningStatus(row.SigningStatus),
		SecureSigningTime:           pgconv.TimeVal(row.SecureSigningTime),
		SigningTime:                 pgconv.TimeVal(row.SigningTime),
	}
}

func toCreateParams(ev events.Event) sqlc.CreateEventParams {
	return sqlc.CreateEventParams{
		MachineID:                   ev.MachineID,
		Decision:                    int32(ev.Decision),
		FilePath:                    ev.FilePath,
		FileSha256:                  ev.FileSha256,
		FileName:                    ev.FileName,
		ExecutingUser:               ev.ExecutingUser,
		ExecutionTime:               pgconv.TimeOrNull(ev.ExecutionTime),
		LoggedInUsers:               pgconv.TextArray(ev.LoggedInUsers),
		CurrentSessions:             pgconv.TextArray(ev.CurrentSessions),
		FileBundleID:                ev.FileBundleID,
		FileBundlePath:              ev.FileBundlePath,
		FileBundleExecutableRelPath: ev.FileBundleExecutableRelPath,
		FileBundleName:              ev.FileBundleName,
		FileBundleVersion:           ev.FileBundleVersion,
		FileBundleVersionString:     ev.FileBundleVersionString,
		FileBundleHash:              ev.FileBundleHash,
		FileBundleHashMillis:        ev.FileBundleHashMillis,
		FileBundleBinaryCount:       ev.FileBundleBinaryCount,
		Pid:                         ev.Pid,
		Ppid:                        ev.Ppid,
		ParentName:                  ev.ParentName,
		TeamID:                      ev.TeamID,
		SigningID:                   ev.SigningID,
		Cdhash:                      ev.Cdhash,
		CsFlags:                     ev.CsFlags,
		SigningStatus:               int32(ev.SigningStatus),
		SecureSigningTime:           pgconv.TimeOrNull(ev.SecureSigningTime),
		SigningTime:                 pgconv.TimeOrNull(ev.SigningTime),
	}
}
