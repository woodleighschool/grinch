package policies

import (
	"errors"
	"fmt"

	celv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/cel"
	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/ext"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain/errx"
)

func validate(policy Policy) error {
	if err := errx.ValidateStruct(policy); err != nil {
		return err
	}

	fields := make(map[string]string)

	if !isValidClientMode(policy.SetClientMode) {
		fields["set_client_mode"] = "Invalid client mode"
	}

	if !isValidFileAccessAction(policy.SetOverrideFileAccessAction) {
		fields["set_override_file_access_action"] = "Invalid file access action"
	}

	validateTargets(policy.Targets, fields)
	validateAttachments(policy.Attachments, fields)

	if len(fields) > 0 {
		return &errx.Error{
			Code:    errx.CodeInvalid,
			Message: "Validation failed",
			Fields:  fields,
		}
	}

	return nil
}

func isValidClientMode(mode syncv1.ClientMode) bool {
	if _, ok := syncv1.ClientMode_name[int32(mode)]; !ok {
		return false
	}
	return mode != syncv1.ClientMode_UNKNOWN_CLIENT_MODE
}

func isValidFileAccessAction(action syncv1.FileAccessAction) bool {
	if _, ok := syncv1.FileAccessAction_name[int32(action)]; !ok {
		return false
	}
	return action != syncv1.FileAccessAction_FILE_ACCESS_ACTION_UNSPECIFIED
}

func validateTargets(targets []Target, fields map[string]string) {
	type targetKey struct {
		kind  TargetKind
		refID uuid.UUID
	}

	var allCount int
	seen := make(map[targetKey]bool)

	for i, t := range targets {
		field := func(name string) string { return fmt.Sprintf("targets[%d].%s", i, name) }

		switch t.Kind {
		case TargetAll:
			allCount++
			if t.RefID != nil && *t.RefID != uuid.Nil {
				fields[field("ref_id")] = "Target 'all' cannot have a reference"
			}
		case TargetUser, TargetGroup, TargetMachine:
			if t.RefID == nil || *t.RefID == uuid.Nil {
				fields[field("ref_id")] = "Reference is required"
			}
		case "":
			fields[field("kind")] = "Kind is required"
			continue
		default:
			fields[field("kind")] = "Invalid kind"
			continue
		}

		refID := uuid.Nil
		if t.RefID != nil {
			refID = *t.RefID
		}
		key := targetKey{kind: t.Kind, refID: refID}
		if seen[key] {
			fields[field("ref_id")] = "Duplicate target"
		}
		seen[key] = true
	}

	if allCount > 0 && len(targets) > allCount {
		fields["targets"] = "Target 'all' cannot be combined with other targets"
	}
	if allCount > 1 {
		fields["targets"] = "Only one 'all' target allowed"
	}
}

func validateAttachments(attachments []Attachment, fields map[string]string) {
	seen := make(map[uuid.UUID]bool)

	for i, a := range attachments {
		field := func(name string) string { return fmt.Sprintf("attachments[%d].%s", i, name) }

		if a.RuleID != uuid.Nil && seen[a.RuleID] {
			fields[field("rule_id")] = "Duplicate rule"
		}
		seen[a.RuleID] = true

		if !isValidAction(a.Action) {
			fields[field("action")] = "Invalid action"
		}

		hasCEL := a.CELExpr != nil && *a.CELExpr != ""
		if a.Action == syncv1.Policy_CEL {
			if !hasCEL {
				fields[field("cel_expr")] = "CEL expression required for CEL action"
			} else if err := validateCEL(*a.CELExpr); err != nil {
				fields[field("cel_expr")] = err.Error()
			}
		} else if hasCEL {
			fields[field("cel_expr")] = "CEL expression only allowed for CEL action"
		}
	}
}

func isValidAction(action syncv1.Policy) bool {
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

	// Santa policies accept CEL expressions that return bool or a ReturnValue int.
	if out == cel.BoolType || out == cel.IntType {
		return nil
	}

	return errors.New("must return bool or int")
}
