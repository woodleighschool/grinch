package rules

import (
	"regexp"
	"strings"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"

	coreerrors "github.com/woodleighschool/grinch/internal/core/errors"
	corerules "github.com/woodleighschool/grinch/internal/core/rules"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

var (
	sha256Pattern    = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
	cdhashPattern    = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
	teamIDPattern    = regexp.MustCompile(`^[A-Z0-9]{10}$`)
	signingIDPattern = regexp.MustCompile(`^(?:[A-Z0-9]{10}|platform):[a-zA-Z0-9.-]+$`)
)

func validateRulePayload(r corerules.Rule) error {
	fields := helpers.FieldErrors{}

	if strings.TrimSpace(r.Name) == "" {
		fields.Add("name", "Name is required")
	}
	if strings.TrimSpace(r.Identifier) == "" {
		fields.Add("identifier", "Identifier is required")
	}

	validateRuleType(r, fields)

	if len(fields) == 0 {
		return nil
	}

	return &coreerrors.Error{
		Code:    coreerrors.CodeInvalid,
		Message: "Validation failed",
		Fields:  fields,
	}
}

func validateRuleType(r corerules.Rule, fields helpers.FieldErrors) {
	switch r.RuleType {
	case syncv1.RuleType_BINARY, syncv1.RuleType_CERTIFICATE:
		if !sha256Pattern.MatchString(r.Identifier) {
			fields.Add("identifier", "Identifier must be a SHA256 hash")
		}
	case syncv1.RuleType_TEAMID:
		if !teamIDPattern.MatchString(r.Identifier) {
			fields.Add("identifier", "Identifier must be a Team ID")
		}
	case syncv1.RuleType_SIGNINGID:
		if !signingIDPattern.MatchString(r.Identifier) {
			fields.Add("identifier", "Identifier must be TEAMID:bundle.id or platform:id")
		}
	case syncv1.RuleType_CDHASH:
		if !cdhashPattern.MatchString(r.Identifier) {
			fields.Add("identifier", "Identifier must be a CDHash (40 hex chars)")
		}
	case syncv1.RuleType_RULETYPE_UNKNOWN:
		fields.Add("rule_type", "Rule type is required")
	default:
		fields.Add("rule_type", "Invalid rule type")
	}
}
