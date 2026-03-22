package postgres

import (
	"encoding/json"
	"fmt"

	"github.com/woodleighschool/grinch/internal/domain"
)

func unmarshalEntitlements(raw []byte) (map[string]any, error) {
	entitlements := make(map[string]any)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &entitlements); err != nil {
			return nil, fmt.Errorf("decode entitlements: %w", err)
		}
	}
	return entitlements, nil
}

func unmarshalSigningChain(raw []byte) ([]domain.SigningChainEntry, error) {
	records := make([]domain.SigningChainEntry, 0)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &records); err != nil {
			return nil, fmt.Errorf("decode signing chain: %w", err)
		}
	}

	for index := range records {
		records[index].ValidFrom = records[index].ValidFrom.UTC()
		records[index].ValidUntil = records[index].ValidUntil.UTC()
	}

	return records, nil
}
