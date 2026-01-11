package listing

// Sort describes an ordered field and its sort direction.
type Sort struct {
	Field string
	Desc  bool
}

// Filter represents a single field constraint in a list query.
type Filter struct {
	Field string
	Value any
}

// Query defines a normalised list request after transport specific parsing.
type Query struct {
	Offset  int
	Limit   int
	Sort    []Sort
	Filters []Filter
	Search  string
}

// Page contains pagination metadata for a list response.
type Page struct {
	Total int64
}
