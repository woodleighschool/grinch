package events

import (
	"context"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	coreevents "github.com/woodleighschool/grinch/internal/core/events"
	"github.com/woodleighschool/grinch/internal/listing"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
	dblisting "github.com/woodleighschool/grinch/internal/store/listing"
)

func listEvents(
	ctx context.Context,
	pool *pgxpool.Pool,
	query listing.Query,
) ([]coreevents.EventListItem, int64, error) {
	cfg := dblisting.Config{
		Table: "events",
		SelectCols: []string{
			"id", "machine_id", "decision", "execution_time",
			"file_path", "file_sha256", "file_name",
			"signing_id", "team_id", "cdhash",
		},
		Columns: map[string]string{
			"id":             "id",
			"machine_id":     "machine_id",
			"decision":       "decision",
			"execution_time": "execution_time",
			"file_path":      "file_path",
			"file_sha256":    "file_sha256",
			"file_name":      "file_name",
			"signing_id":     "signing_id",
			"team_id":        "team_id",
			"cdhash":         "cdhash",
		},
		SearchColumns: []string{"file_path", "file_sha256", "file_name", "signing_id", "team_id"},
		DefaultSort:   listing.Sort{Field: "execution_time", Desc: true},
	}
	return dblisting.List(ctx, pool, cfg, query, scanEventListItem)
}

func scanEventListItem(rows pgx.Rows) (coreevents.EventListItem, error) {
	var (
		item          coreevents.EventListItem
		decision      int32
		executionTime pgtype.Timestamptz
	)

	err := rows.Scan(
		&item.ID, &item.MachineID, &decision, &executionTime,
		&item.FilePath, &item.FileSha256, &item.FileName,
		&item.SigningID, &item.TeamID, &item.Cdhash,
	)
	if err != nil {
		return item, err
	}

	item.Decision = syncv1.Decision(decision)
	item.ExecutionTime = pgconv.TimeVal(executionTime)
	return item, nil
}
