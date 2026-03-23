package postgres

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
)

const (
	sortDirectionAsc  = "ASC"
	sortDirectionDesc = "DESC"
)

func searchPattern(search string) string {
	if search == "" {
		return ""
	}

	return "%" + search + "%"
}

func toStrings[T ~string](values []T) []string {
	if len(values) == 0 {
		return nil
	}

	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}

	return out
}

func orderBy(
	sortField string,
	sortOrder string,
	allowed map[string]string,
	defaultOrder []string,
) (string, error) {
	if len(defaultOrder) == 0 {
		return "", errors.New("default order is required")
	}

	if sortField == "" {
		return strings.Join(defaultOrder, ", "), nil
	}

	sortColumn, ok := allowed[sortField]
	if !ok {
		return "", fmt.Errorf("%w field %q", domain.ErrInvalidSort, sortField)
	}

	sortDirection := sortDirectionAsc
	if strings.EqualFold(sortOrder, "desc") {
		sortDirection = sortDirectionDesc
	}

	parts := make([]string, 0, len(defaultOrder)+1)
	parts = append(parts, sortColumn+" "+sortDirection)

	for _, tiebreaker := range defaultOrder {
		tiebreakerColumn, _, _ := strings.Cut(tiebreaker, " ")
		if tiebreakerColumn != sortColumn {
			parts = append(parts, tiebreaker)
		}
	}

	return strings.Join(parts, ", "), nil
}

func collectRows[T any](
	rows pgx.Rows,
	scan func(pgx.Rows) (T, int32, error),
) ([]T, int32, error) {
	defer rows.Close()

	var total int32
	items := make([]T, 0)

	for rows.Next() {
		item, rowTotal, err := scan(rows)
		if err != nil {
			return nil, 0, err
		}

		items = append(items, item)
		total = rowTotal
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
