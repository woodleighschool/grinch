// Package constraints maps database constraint names to domain field names.
package constraints

// RuleFields returns the mapping of rule constraint names to validation field names.
func RuleFields() map[string]string {
	return map[string]string{
		"rules_name_key":       "name",
		"rules_identifier_key": "identifier",
	}
}

// PolicyFields returns the mapping of policy constraint names to validation field names.
func PolicyFields() map[string]string {
	return map[string]string{
		"policies_name_key":     "name",
		"policies_priority_key": "priority",
	}
}
