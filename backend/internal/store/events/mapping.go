package events

import (
	coreevents "github.com/woodleighschool/grinch/internal/core/events"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
	"github.com/woodleighschool/grinch/internal/store/db/sqlc"
)

// mapSigningChain converts signing chain query rows into domain certificates.
func mapSigningChain(rows []sqlc.ListSigningChainEntriesByEventIDRow) []coreevents.Certificate {
	out := make([]coreevents.Certificate, len(rows))
	for i, row := range rows {
		out[i] = coreevents.Certificate{
			SHA256:     row.Sha256,
			CN:         row.Cn,
			Org:        row.Org,
			OU:         row.Ou,
			ValidFrom:  pgconv.TimeVal(row.ValidFrom),
			ValidUntil: pgconv.TimeVal(row.ValidUntil),
		}
	}
	return out
}

// mapEntitlements converts entitlement query rows into domain entitlements.
func mapEntitlements(rows []sqlc.ListEntitlementsByEventIDRow) []coreevents.Entitlement {
	out := make([]coreevents.Entitlement, len(rows))
	for i, row := range rows {
		out[i] = coreevents.Entitlement{
			Key:   row.Key,
			Value: row.Value,
		}
	}
	return out
}
