package events

import (
	"github.com/woodleighschool/grinch/internal/domain/events"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
	"github.com/woodleighschool/grinch/internal/store/db/sqlc"
)

// mapSigningChain converts signing chain query rows into domain certificates.
func mapSigningChain(rows []sqlc.ListSigningChainEntriesByEventIDRow) []events.Certificate {
	out := make([]events.Certificate, len(rows))
	for i, row := range rows {
		out[i] = events.Certificate{
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
func mapEntitlements(rows []sqlc.ListEntitlementsByEventIDRow) []events.Entitlement {
	out := make([]events.Entitlement, len(rows))
	for i, row := range rows {
		out[i] = events.Entitlement{
			Key:   row.Key,
			Value: row.Value,
		}
	}
	return out
}
