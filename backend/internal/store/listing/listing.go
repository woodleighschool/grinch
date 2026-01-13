// Package listing provides generic list query execution for store repositories.
package listing

import (
	"context"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woodleighschool/grinch/internal/listing"
)

// Config describes the persistence mapping for a listable resource.
type Config struct {
	Table         string
	SelectCols    []string
	Columns       map[string]string
	SearchColumns []string
	DefaultSort   listing.Sort
}

// List executes a list query and scans results using scanRow.
func List[T any](
	ctx context.Context,
	pool *pgxpool.Pool,
	cfg Config,
	query listing.Query,
	scanRow func(pgx.Rows) (T, error),
) ([]T, int64, error) {
	dialect := goqu.Dialect("postgres")

	ds := dialect.From(cfg.Table)
	if len(cfg.SelectCols) > 0 {
		cols := make([]any, len(cfg.SelectCols))
		for i, c := range cfg.SelectCols {
			cols[i] = goqu.I(c)
		}
		ds = ds.Select(cols...)
	}

	ds = applyFilters(ds, cfg, query)
	ds = applySearch(ds, cfg, query)

	countDS := goqu.From(ds.As("sub")).Select(goqu.COUNT(goqu.L("1")))
	countSQL, args, err := countDS.ToSQL()
	if err != nil {
		return nil, 0, fmt.Errorf("build count sql: %w", err)
	}

	var total int64
	if err = pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	ds = applySort(ds, cfg, query)
	if query.Limit > 0 {
		ds = ds.Limit(uint(query.Limit))
	}
	if query.Offset > 0 {
		ds = ds.Offset(uint(query.Offset))
	}

	listSQL, args, err := ds.ToSQL()
	if err != nil {
		return nil, 0, fmt.Errorf("build list sql: %w", err)
	}

	rows, err := pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query list: %w", err)
	}
	defer rows.Close()

	results := make([]T, 0)
	var item T
	for rows.Next() {
		item, err = scanRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan row: %w", err)
		}
		results = append(results, item)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return results, total, nil
}

func applyFilters(ds *goqu.SelectDataset, cfg Config, query listing.Query) *goqu.SelectDataset {
	for _, f := range query.Filters {
		col, ok := cfg.Columns[f.Field]
		if !ok {
			continue
		}

		switch v := f.Value.(type) {
		case []any:
			if len(v) > 0 {
				ds = ds.Where(goqu.I(col).In(v...))
			}
		case []string:
			if len(v) > 0 {
				vals := make([]any, len(v))
				for i := range v {
					vals[i] = v[i]
				}
				ds = ds.Where(goqu.I(col).In(vals...))
			}
		default:
			ds = ds.Where(goqu.I(col).Eq(v))
		}
	}
	return ds
}

func applySearch(ds *goqu.SelectDataset, cfg Config, query listing.Query) *goqu.SelectDataset {
	search := strings.TrimSpace(query.Search)
	if search == "" {
		return ds
	}

	cols := cfg.SearchColumns
	if len(cols) == 0 {
		cols = cfg.SelectCols
	}
	if len(cols) == 0 {
		return ds
	}

	ilikes := make([]goqu.Expression, 0, len(cols))
	for _, col := range cols {
		ilikes = append(ilikes, goqu.Cast(goqu.I(col), "TEXT").ILike("%"+search+"%"))
	}
	return ds.Where(goqu.Or(ilikes...))
}

func applySort(ds *goqu.SelectDataset, cfg Config, query listing.Query) *goqu.SelectDataset {
	sorts := query.Sort
	if len(sorts) == 0 && cfg.DefaultSort.Field != "" {
		sorts = []listing.Sort{cfg.DefaultSort}
	}

	for _, s := range sorts {
		col, ok := cfg.Columns[s.Field]
		if !ok {
			continue
		}
		if s.Desc {
			ds = ds.Order(goqu.I(col).Desc())
		} else {
			ds = ds.Order(goqu.I(col).Asc())
		}
	}

	return ds
}
