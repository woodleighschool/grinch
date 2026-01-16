package helpers

// FieldErrors collects validation messages keyed by field name.
type FieldErrors map[string]string

// Add stores the message for the field if it has not already been set.
func (f FieldErrors) Add(field, msg string) {
	if f == nil {
		return
	}
	if _, exists := f[field]; !exists {
		f[field] = msg
	}
}
