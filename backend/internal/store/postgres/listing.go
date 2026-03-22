package postgres

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
)

func searchPattern(search string) string {
	if search == "" {
		return ""
	}

	return "%" + search + "%"
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}

	return *id
}

func toStrings[T ~string](values []T) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, string(value))
	}

	return result
}

func orderBy(sort string, order string, allowed map[string]string, fallback []string) (string, error) {
	term := strings.TrimSpace(sort)
	if term == "" {
		return strings.Join(fallback, ", "), nil
	}

	column, ok := allowed[term]
	if !ok {
		return "", fmt.Errorf("%w field %q", domain.ErrInvalidSort, term)
	}

	direction := "ASC"
	if strings.EqualFold(strings.TrimSpace(order), "desc") {
		direction = "DESC"
	}

	orderParts := make([]string, 0, len(fallback)+1)
	orderParts = append(orderParts, column+" "+direction)

	for _, defaultPart := range fallback {
		defaultColumn := strings.TrimSpace(strings.Split(defaultPart, " ")[0])
		if defaultColumn == column {
			continue
		}
		orderParts = append(orderParts, defaultPart)
	}

	return strings.Join(orderParts, ", "), nil
}

func collectRows[T any](
	rows pgx.Rows,
	scan func(pgx.Rows) (T, int32, error),
) ([]T, int32, error) {
	defer rows.Close()

	var items []T
	var total int32
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
