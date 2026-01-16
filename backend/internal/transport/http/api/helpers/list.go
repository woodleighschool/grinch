package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/woodleighschool/grinch/internal/listing"
)

// ParseListQuery parses React Admin list query parameters into a listing.Query.
func ParseListQuery(r *http.Request) listing.Query {
	q := r.URL.Query()
	filters, search := parseFilters(q.Get("filter"))

	return listing.Query{
		Offset:  ParseRangeStart(q.Get("range")),
		Limit:   ParseRangeLimit(q.Get("range")),
		Sort:    parseSort(q.Get("sort")),
		Filters: filters,
		Search:  search,
	}
}

func parseSort(raw string) []listing.Sort {
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err != nil || len(arr) < 2 {
		return nil
	}

	return []listing.Sort{{
		Field: arr[0],
		Desc:  arr[1] == "DESC",
	}}
}

// ParseRangeStart parses range=[start,end] and returns start.
func ParseRangeStart(raw string) int {
	var arr []int
	if err := json.Unmarshal([]byte(raw), &arr); err != nil || len(arr) < 1 {
		return 0
	}
	if arr[0] < 0 {
		return 0
	}
	return arr[0]
}

// ParseRangeLimit parses range=[start,end] and returns (end minus start plus 1).
func ParseRangeLimit(raw string) int {
	var arr []int
	if err := json.Unmarshal([]byte(raw), &arr); err != nil || len(arr) < 2 {
		return 0
	}

	start, end := arr[0], arr[1]
	if start < 0 {
		start = 0
	}
	if end < 0 || end < start {
		return 0
	}

	return end - start + 1
}

// ParseRange parses range=[start,end] and returns start and limit.
func ParseRange(raw string) (int, int) {
	start := ParseRangeStart(raw)
	limit := ParseRangeLimit(raw)
	if limit == 0 {
		limit = 25
	}
	return start, limit
}

func parseFilters(raw string) ([]listing.Filter, string) {
	if raw == "" {
		return nil, ""
	}

	var filterMap map[string]any
	if err := json.Unmarshal([]byte(raw), &filterMap); err != nil {
		return nil, ""
	}

	var search string
	if q, ok := filterMap["q"].(string); ok {
		search = q
		delete(filterMap, "q")
	}

	filters := make([]listing.Filter, 0, len(filterMap))
	for field, value := range filterMap {
		filters = append(filters, listing.Filter{Field: field, Value: value})
	}

	return filters, search
}

// WriteList writes a React Admin list response and sets the Content Range header.
func WriteList[T any](w http.ResponseWriter, r *http.Request, resource string, items []T, total int64) {
	start := ParseRangeStart(r.URL.Query().Get("range"))
	end := max(start+len(items)-1, start)

	w.Header().Set("Content-Range", fmt.Sprintf("%s %d-%d/%d", resource, start, end, total))
	w.Header().Set("Access-Control-Expose-Headers", "Content-Range")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if items == nil {
		items = make([]T, 0)
	}
	_ = json.NewEncoder(w).Encode(items)
}
