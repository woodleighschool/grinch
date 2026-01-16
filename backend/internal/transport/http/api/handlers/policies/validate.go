package policies

import (
	"errors"
	"fmt"
	"strings"

	celv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/cel"
	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/ext"
	"github.com/google/uuid"

	coreerrors "github.com/woodleighschool/grinch/internal/core/errors"
	corepolicies "github.com/woodleighschool/grinch/internal/core/policies"
	"github.com/woodleighschool/grinch/internal/transport/http/api/helpers"
)

func validatePolicyPayload(p corepolicies.Policy) error {
	fields := helpers.FieldErrors{}

	if strings.TrimSpace(p.Name) == "" {
		fields.Add("name", "Name is required")
	}
	if p.Priority < 0 {
		fields.Add("priority", "Priority must be zero or greater")
	}

	validateSettings(p, fields)
	validateTargets(p.Targets, fields)
	validateAttachments(p.Attachments, fields)

	if len(fields) == 0 {
		return nil
	}

	return &coreerrors.Error{
		Code:    coreerrors.CodeInvalid,
		Message: "Validation failed",
		Fields:  fields,
	}
}

func validateSettings(p corepolicies.Policy, fields helpers.FieldErrors) {
	if !validClientMode(p.SetClientMode) {
		fields.Add("set_client_mode", "Invalid client mode")
	}
	if p.SetBatchSize < 1 {
		fields.Add("set_batch_size", "Batch size must be at least 1")
	}
	const minSyncInterval = 60
	if p.SetFullSyncIntervalSeconds < minSyncInterval {
		fields.Add("set_full_sync_interval_seconds", "Must be at least 60 seconds")
	}
	if p.SetPushNotificationFullSyncIntervalSeconds < minSyncInterval {
		fields.Add("set_push_notification_full_sync_interval_seconds", "Must be at least 60 seconds")
	}
	if p.SetPushNotificationGlobalRuleSyncDeadlineSeconds < 0 {
		fields.Add("set_push_notification_global_rule_sync_deadline_seconds", "Must be zero or greater")
	}
	if !validFileAccessAction(p.SetOverrideFileAccessAction) {
		fields.Add("set_override_file_access_action", "Invalid file access action")
	}
}

func validClientMode(mode syncv1.ClientMode) bool {
	if _, ok := syncv1.ClientMode_name[int32(mode)]; !ok {
		return false
	}
	return mode != syncv1.ClientMode_UNKNOWN_CLIENT_MODE
}

func validFileAccessAction(action syncv1.FileAccessAction) bool {
	if _, ok := syncv1.FileAccessAction_name[int32(action)]; !ok {
		return false
	}
	return action != syncv1.FileAccessAction_FILE_ACCESS_ACTION_UNSPECIFIED
}

func validateTargets(targets []corepolicies.PolicyTarget, fields helpers.FieldErrors) {
	type targetKey struct {
		kind  corepolicies.PolicyTargetKind
		refID uuid.UUID
	}

	var allCount int
	seen := make(map[targetKey]bool)

	for i, t := range targets {
		field := func(name string) string { return fmt.Sprintf("targets[%d].%s", i, name) }

		switch t.Kind {
		case corepolicies.TargetAll:
			allCount++
			if t.RefID != nil && *t.RefID != uuid.Nil {
				fields.Add(field("ref_id"), "Target 'all' cannot have a reference")
			}
		case corepolicies.TargetUser, corepolicies.TargetGroup, corepolicies.TargetMachine:
			if t.RefID == nil || *t.RefID == uuid.Nil {
				fields.Add(field("ref_id"), "Reference is required")
			}
		case "":
			fields.Add(field("kind"), "Kind is required")
			continue
		default:
			fields.Add(field("kind"), "Invalid kind")
			continue
		}

		refID := uuid.Nil
		if t.RefID != nil {
			refID = *t.RefID
		}
		key := targetKey{kind: t.Kind, refID: refID}
		if seen[key] {
			fields.Add(field("ref_id"), "Duplicate target")
		}
		seen[key] = true
	}

	if allCount > 0 && len(targets) > allCount {
		fields.Add("targets", "Target 'all' cannot be combined with other targets")
	}
	if allCount > 1 {
		fields.Add("targets", "Only one 'all' target allowed")
	}
}

func validateAttachments(attachments []corepolicies.PolicyAttachment, fields helpers.FieldErrors) {
	seen := make(map[uuid.UUID]bool)

	for i, a := range attachments {
		field := func(name string) string { return fmt.Sprintf("attachments[%d].%s", i, name) }

		if a.RuleID != uuid.Nil && seen[a.RuleID] {
			fields.Add(field("rule_id"), "Duplicate rule")
		}
		seen[a.RuleID] = true

		if !validAction(a.Action) {
			fields.Add(field("action"), "Invalid action")
		}

		hasCEL := a.CELExpr != nil && strings.TrimSpace(*a.CELExpr) != ""
		if a.Action == syncv1.Policy_CEL {
			if !hasCEL {
				fields.Add(field("cel_expr"), "CEL expression required for CEL action")
			} else if err := validateCEL(*a.CELExpr); err != nil {
				fields.Add(field("cel_expr"), err.Error())
			}
		} else if hasCEL {
			fields.Add(field("cel_expr"), "CEL expression only allowed for CEL action")
		}
	}
}

func validAction(action syncv1.Policy) bool {
	switch action {
	case syncv1.Policy_ALLOWLIST,
		syncv1.Policy_ALLOWLIST_COMPILER,
		syncv1.Policy_BLOCKLIST,
		syncv1.Policy_SILENT_BLOCKLIST,
		syncv1.Policy_CEL:
		return true
	case syncv1.Policy_POLICY_UNKNOWN, syncv1.Policy_REMOVE:
		return false
	default:
		return false
	}
}

func validateCEL(expr string) error {
	env, err := getCELEnv()
	if err != nil {
		return err
	}

	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return issues.Err()
	}

	out := ast.OutputType()
	if out == nil {
		return errors.New("unknown output type")
	}

	if out == cel.BoolType || out == cel.IntType {
		return nil
	}

	return errors.New("must return bool or int")
}

func getCELEnv() (*cel.Env, error) {
	execDesc := (&celv1.ExecutionContext{}).ProtoReflect().Descriptor()

	env, err := cel.NewEnv(
		cel.DeclareContextProto(execDesc),
		cel.Types(&celv1.ExecutionContext{}, &celv1.ExecutableFile{}),
		ext.Strings(),
		cel.VariableDecls(
			decls.NewConstant("ALLOWLIST", types.IntType, types.Int(int64(celv1.ReturnValue_ALLOWLIST))),
			decls.NewConstant(
				"ALLOWLIST_COMPILER",
				types.IntType,
				types.Int(int64(celv1.ReturnValue_ALLOWLIST_COMPILER)),
			),
			decls.NewConstant("BLOCKLIST", types.IntType, types.Int(int64(celv1.ReturnValue_BLOCKLIST))),
			decls.NewConstant("SILENT_BLOCKLIST", types.IntType, types.Int(int64(celv1.ReturnValue_SILENT_BLOCKLIST))),
		),
	)
	return env, err
}
