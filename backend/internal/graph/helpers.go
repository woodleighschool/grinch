package graph

// deref safely returns the string pointed to by value or an empty string.
func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
