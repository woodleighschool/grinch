package entra

import abstractions "github.com/microsoft/kiota-abstractions-go"

// advancedQueryHeaders returns request headers required for Microsoft Graph
// advanced queries, which rely on eventual consistency.
func advancedQueryHeaders() *abstractions.RequestHeaders {
	h := abstractions.NewRequestHeaders()
	h.Add("ConsistencyLevel", "eventual")
	return h
}
