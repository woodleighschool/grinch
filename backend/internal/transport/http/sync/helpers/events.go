package helpers

import (
	"strings"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	coreevents "github.com/woodleighschool/grinch/internal/core/events"
	"github.com/woodleighschool/grinch/internal/store/db/pgconv"
)

func ConvertEvents(machineID uuid.UUID, req *syncv1.EventUploadRequest) []coreevents.Event {
	events := req.GetEvents()
	if len(events) == 0 {
		return nil
	}

	out := make([]coreevents.Event, 0, len(events))
	for _, ev := range events {
		if ev == nil {
			continue
		}
		out = append(out, convertEvent(machineID, ev))
	}
	return out
}

func convertEvent(machineID uuid.UUID, ev *syncv1.Event) coreevents.Event {
	return coreevents.Event{
		MachineID:                   machineID,
		Decision:                    ev.GetDecision(),
		FilePath:                    ev.GetFilePath(),
		FileSha256:                  ev.GetFileSha256(),
		FileName:                    ev.GetFileName(),
		ExecutingUser:               ev.GetExecutingUser(),
		ExecutionTime:               timePtrFromFloat(ev.GetExecutionTime()),
		LoggedInUsers:               sanitiseStrings(ev.GetLoggedInUsers()),
		CurrentSessions:             sanitiseStrings(ev.GetCurrentSessions()),
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

func convertCertificates(chain []*syncv1.Certificate) []coreevents.Certificate {
	if len(chain) == 0 {
		return []coreevents.Certificate{}
	}

	out := make([]coreevents.Certificate, 0, len(chain))
	for _, cert := range chain {
		if cert == nil {
			continue
		}
		out = append(out, coreevents.Certificate{
			SHA256:     cert.GetSha256(),
			CN:         cert.GetCn(),
			Org:        cert.GetOrg(),
			OU:         cert.GetOu(),
			ValidFrom:  timePtrFromUnix(cert.GetValidFrom()),
			ValidUntil: timePtrFromUnix(cert.GetValidUntil()),
		})
	}
	return out
}

func convertEntitlements(info *syncv1.EntitlementInfo) []coreevents.Entitlement {
	if info == nil {
		return []coreevents.Entitlement{}
	}

	ents := info.GetEntitlements()
	if len(ents) == 0 {
		return []coreevents.Entitlement{}
	}

	out := make([]coreevents.Entitlement, 0, len(ents))
	for _, ent := range ents {
		if ent == nil {
			continue
		}
		out = append(out, coreevents.Entitlement{
			Key:   ent.GetKey(),
			Value: ent.GetValue(),
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

func sanitiseStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, sanitiseString(v))
	}
	return out
}

func sanitiseString(value string) string {
	if value == "" {
		return ""
	}
	return strings.ReplaceAll(value, "\x00", "")
}
