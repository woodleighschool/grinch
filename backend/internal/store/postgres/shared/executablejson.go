package pgutil

import (
	"encoding/json"
	"fmt"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

type signingChainRecord struct {
	CommonName         string    `json:"common_name"`
	Organization       string    `json:"organization"`
	OrganizationalUnit string    `json:"organizational_unit"`
	SHA256             string    `json:"sha256"`
	ValidFrom          time.Time `json:"valid_from"`
	ValidUntil         time.Time `json:"valid_until"`
}

type fileAccessProcessRecord struct {
	Position     int32     `json:"position,omitempty"`
	Pid          int32     `json:"pid"`
	FilePath     string    `json:"file_path"`
	ExecutableID uuid.UUID `json:"executable_id"`
}

func MarshalEntitlements(info *syncv1.EntitlementInfo) ([]byte, error) {
	if info == nil {
		return []byte("{}"), nil
	}

	entitlements := make(map[string]any, len(info.GetEntitlements()))
	for _, entitlement := range info.GetEntitlements() {
		if entitlement == nil {
			continue
		}

		key := entitlement.GetKey()
		if key == "" {
			continue
		}

		rawValue := entitlement.GetValue()
		if rawValue == "" {
			entitlements[key] = nil
			continue
		}

		var decodedValue any
		if err := json.Unmarshal([]byte(rawValue), &decodedValue); err != nil {
			entitlements[key] = rawValue
			continue
		}

		entitlements[key] = decodedValue
	}

	encoded, err := json.Marshal(entitlements)
	if err != nil {
		return nil, fmt.Errorf("marshal entitlements: %w", err)
	}

	return encoded, nil
}

func UnmarshalEntitlements(raw []byte) (map[string]any, error) {
	entitlements := make(map[string]any)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &entitlements); err != nil {
			return nil, fmt.Errorf("decode entitlements: %w", err)
		}
	}
	return entitlements, nil
}

func MarshalSigningChain(certificates []*syncv1.Certificate) ([]byte, error) {
	records := make([]signingChainRecord, 0, len(certificates))
	for _, certificate := range certificates {
		if certificate == nil {
			continue
		}

		records = append(records, signingChainRecord{
			CommonName:         certificate.GetCn(),
			Organization:       certificate.GetOrg(),
			OrganizationalUnit: certificate.GetOu(),
			SHA256:             certificate.GetSha256(),
			ValidFrom:          time.Unix(int64(certificate.GetValidFrom()), 0).UTC(),
			ValidUntil:         time.Unix(int64(certificate.GetValidUntil()), 0).UTC(),
		})
	}

	encoded, err := json.Marshal(records)
	if err != nil {
		return nil, fmt.Errorf("marshal signing chain: %w", err)
	}

	return encoded, nil
}

func UnmarshalSigningChain(raw []byte) ([]domain.SigningChainEntry, error) {
	records := make([]signingChainRecord, 0)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &records); err != nil {
			return nil, fmt.Errorf("decode signing chain: %w", err)
		}
	}

	signingChain := make([]domain.SigningChainEntry, 0, len(records))
	for _, record := range records {
		signingChain = append(signingChain, domain.SigningChainEntry{
			CommonName:         record.CommonName,
			Organization:       record.Organization,
			OrganizationalUnit: record.OrganizationalUnit,
			SHA256:             record.SHA256,
			ValidFrom:          record.ValidFrom.UTC(),
			ValidUntil:         record.ValidUntil.UTC(),
		})
	}

	return signingChain, nil
}

func MarshalFileAccessProcessChain(processes []domain.FileAccessEventProcess) ([]byte, error) {
	records := make([]fileAccessProcessRecord, 0, len(processes))
	for index, process := range processes {
		records = append(records, fileAccessProcessRecord{
			Position:     int32(index),
			Pid:          process.Pid,
			FilePath:     process.FilePath,
			ExecutableID: process.ExecutableID,
		})
	}

	encoded, err := json.Marshal(records)
	if err != nil {
		return nil, fmt.Errorf("marshal file access process chain: %w", err)
	}

	return encoded, nil
}

func UnmarshalFileAccessProcessChain(raw []byte) ([]domain.FileAccessEventProcess, error) {
	records := make([]fileAccessProcessRecord, 0)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &records); err != nil {
			return nil, fmt.Errorf("decode file access process chain: %w", err)
		}
	}

	processes := make([]domain.FileAccessEventProcess, 0, len(records))
	for _, record := range records {
		processes = append(processes, domain.FileAccessEventProcess{
			Pid:          record.Pid,
			FilePath:     record.FilePath,
			ExecutableID: record.ExecutableID,
		})
	}

	return processes, nil
}
