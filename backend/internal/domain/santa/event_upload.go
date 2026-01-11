package santa

import (
	"context"
	"fmt"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
	"github.com/woodleighschool/grinch/internal/domain/events"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
)

// EventUpload persists execution events reported by a Santa client.
func (s SyncService) EventUpload(
	ctx context.Context,
	machineID uuid.UUID,
	req *syncv1.EventUploadRequest,
) (*syncv1.EventUploadResponse, error) {
	if err := s.events.InsertBatch(ctx, convertEvents(machineID, req.GetEvents())); err != nil {
		return nil, fmt.Errorf("eventupload: insert: %w", err)
	}

	machine, err := s.machines.Get(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("eventupload: get machine: %w", err)
	}
	machine.LastSeen = time.Now().UTC()
	if _, err = s.machines.Upsert(ctx, machine); err != nil {
		return nil, fmt.Errorf("eventupload: upsert machine: %w", err)
	}

	return &syncv1.EventUploadResponse{}, nil
}

func convertEvents(machineID uuid.UUID, in []*syncv1.Event) []events.Event {
	if len(in) == 0 {
		return nil
	}

	out := make([]events.Event, 0, len(in))
	for _, ev := range in {
		if ev == nil {
			continue
		}
		out = append(out, convertEvent(machineID, ev))
	}
	return out
}

func convertEvent(machineID uuid.UUID, ev *syncv1.Event) events.Event {
	loggedInUsers := ev.GetLoggedInUsers()
	if loggedInUsers == nil {
		loggedInUsers = []string{}
	}
	currentSessions := ev.GetCurrentSessions()
	if currentSessions == nil {
		currentSessions = []string{}
	}

	return events.Event{
		MachineID:                   machineID,
		Decision:                    ev.GetDecision(),
		FilePath:                    ev.GetFilePath(),
		FileSha256:                  ev.GetFileSha256(),
		FileName:                    ev.GetFileName(),
		ExecutingUser:               ev.GetExecutingUser(),
		ExecutionTime:               timePtrFromFloat(ev.GetExecutionTime()),
		LoggedInUsers:               loggedInUsers,
		CurrentSessions:             currentSessions,
		FileBundleID:                ev.GetFileBundleId(),
		FileBundlePath:              ev.GetFileBundlePath(),
		FileBundleExecutableRelPath: ev.GetFileBundleExecutableRelPath(),
		FileBundleName:              ev.GetFileBundleName(),
		FileBundleVersion:           ev.GetFileBundleVersion(),
		FileBundleVersionString:     ev.GetFileBundleVersionString(),
		FileBundleHash:              ev.GetFileBundleHash(),
		FileBundleHashMillis:        pgconv.Uint32ToInt32(ev.GetFileBundleHashMillis()),
		FileBundleBinaryCount:       pgconv.Uint32ToInt32(ev.GetFileBundleBinaryCount()),
		Pid:                         ev.GetPid(),
		Ppid:                        ev.GetPpid(),
		ParentName:                  ev.GetParentName(),
		TeamID:                      ev.GetTeamId(),
		SigningID:                   ev.GetSigningId(),
		Cdhash:                      ev.GetCdhash(),
		SigningChain:                convertCertificates(ev.GetSigningChain()),
		Entitlements:                convertEntitlements(ev.GetEntitlementInfo()),
		CsFlags:                     pgconv.Uint32ToInt32(ev.GetCsFlags()),
		SigningStatus:               ev.GetSigningStatus(),
		SecureSigningTime:           timePtrFromUnix(ev.GetSecureSigningTime()),
		SigningTime:                 timePtrFromUnix(ev.GetSigningTime()),
	}
}

func convertCertificates(chain []*syncv1.Certificate) []events.Certificate {
	if len(chain) == 0 {
		return []events.Certificate{}
	}

	out := make([]events.Certificate, 0, len(chain))
	for _, c := range chain {
		if c == nil {
			continue
		}
		out = append(out, events.Certificate{
			SHA256:     c.GetSha256(),
			CN:         c.GetCn(),
			Org:        c.GetOrg(),
			OU:         c.GetOu(),
			ValidFrom:  timePtrFromUnix(c.GetValidFrom()),
			ValidUntil: timePtrFromUnix(c.GetValidUntil()),
		})
	}
	return out
}

func convertEntitlements(info *syncv1.EntitlementInfo) []events.Entitlement {
	if info == nil {
		return []events.Entitlement{}
	}

	ents := info.GetEntitlements()
	if len(ents) == 0 {
		return []events.Entitlement{}
	}

	out := make([]events.Entitlement, 0, len(ents))
	for _, e := range ents {
		if e == nil {
			continue
		}
		out = append(out, events.Entitlement{
			Key:   e.GetKey(),
			Value: e.GetValue(),
		})
	}
	return out
}

func timePtrFromFloat(seconds float64) *time.Time {
	if seconds == 0 {
		return nil
	}
	nsec := int64(seconds * float64(time.Second))
	t := time.Unix(0, nsec).UTC()
	return &t
}

func timePtrFromUnix(seconds uint32) *time.Time {
	if seconds == 0 {
		return nil
	}
	t := time.Unix(int64(seconds), 0).UTC()
	return &t
}
