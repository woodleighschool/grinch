package helpers

import "encoding/json"

// ParseFilterMap returns the decoded filter JSON map or nil when absent/invalid.
func ParseFilterMap(raw string) map[string]any {
	if raw == "" {
		return nil
	}

	var m map[string]any
	_ = json.Unmarshal([]byte(raw), &m)
	return m
}
