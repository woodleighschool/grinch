package apihttp

import (
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

func parseListOptions[T ~string, O ~string](
	limit *int32,
	offset *int32,
	search *T,
	sort *Sort,
	order *O,
	ids *IdsFilter,
) (domain.ListOptions, error) {
	var resolvedLimit int32
	if limit != nil {
		resolvedLimit = *limit
	}

	var resolvedOffset int32
	if offset != nil {
		resolvedOffset = *offset
	}

	switch {
	case limit != nil && resolvedLimit < 1:
		return domain.ListOptions{}, badRequestError("limit must be >= 1")
	case resolvedOffset < 0:
		return domain.ListOptions{}, badRequestError("offset must be >= 0")
	}

	return domain.ListOptions{
		IDs:    cloneUUIDs(ids),
		Limit:  resolvedLimit,
		Offset: resolvedOffset,
		Search: optionalString(search),
		Sort:   optionalString(sort),
		Order:  optionalString(order),
	}, nil
}

func optionalString[T ~string](v *T) string {
	if v == nil {
		return ""
	}

	return string(*v)
}

func cloneUUIDs(v *IdsFilter) []uuid.UUID {
	if v == nil {
		return nil
	}

	out := make([]uuid.UUID, len(*v))
	copy(out, *v)

	return out
}

func cloneBools(v *EnabledFilter) []bool {
	if v == nil {
		return nil
	}

	out := make([]bool, len(*v))
	copy(out, *v)

	return out
}

func parseOptionalValues[T any, V ~string](values *[]V, parse func(string) (T, error)) ([]T, error) {
	if values == nil {
		return nil, nil
	}

	out := make([]T, 0, len(*values))
	for _, v := range *values {
		parsed, err := parse(string(v))
		if err != nil {
			return nil, badRequestError(err.Error())
		}

		out = append(out, parsed)
	}

	return out, nil
}
