package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/google/cel-go/cel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/rules"
	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

const (
	errCodeValidationFailed    = "VALIDATION_FAILED"
	errCodeDuplicateIdentifier = "DUPLICATE_IDENTIFIER"
	errCodeDuplicateScope      = "DUPLICATE_SCOPE"
)

type identifierValidator struct {
	regex   *regexp.Regexp
	message string
}

var applicationIdentifierValidators = map[string]identifierValidator{
	"BINARY": {
		regex:   regexp.MustCompile(`^[a-fA-F0-9]{64}$`),
		message: "Identifier must be a valid 64-character SHA-256 hash",
	},
	"CERTIFICATE": {
		regex:   regexp.MustCompile(`^[a-fA-F0-9]{64}$`),
		message: "Certificate identifiers must be a 64-character SHA-256 hash",
	},
	"SIGNINGID": {
		regex:   regexp.MustCompile(`^(?:[A-Z0-9]{10}|platform):[a-zA-Z0-9.-]+$`),
		message: "Signing IDs must follow TEAMID/platform:bundle.identifier format",
	},
	"TEAMID": {
		regex:   regexp.MustCompile(`^[A-Z0-9]{10}$`),
		message: "Team IDs must be 10 uppercase alphanumeric characters",
	},
	"CDHASH": {
		regex:   regexp.MustCompile(`^[a-fA-F0-9]{40}$`),
		message: "CDHashes must be a 40-character hexadecimal value",
	},
}

type fieldErrors map[string]string

// applicationValidationResult is returned to callers for display.
type applicationValidationResult struct {
	Name          string `json:"name"`
	RuleType      string `json:"rule_type"`
	Identifier    string `json:"identifier"`
	Description   string `json:"description,omitempty"`
	BlockMessage  string `json:"block_message,omitempty"`
	CelEnabled    bool   `json:"cel_enabled"`
	CelExpression string `json:"cel_expression,omitempty"`
}

// applicationValidationInput contains raw fields before trimming/validation.
type applicationValidationInput struct {
	Name          string
	RuleType      string
	Identifier    string
	Description   string
	BlockMessage  string
	CelEnabled    bool
	CelExpression string
}

// scopeValidationResult is returned when validating new scopes.
type scopeValidationResult struct {
	ApplicationID uuid.UUID `json:"application_id"`
	TargetType    string    `json:"target_type"`
	TargetID      uuid.UUID `json:"target_id"`
	Action        string    `json:"action"`
}

var (
	celEnvOnce sync.Once
	celEnv     *cel.Env
	celEnvErr  error
)

// celValidationEnv lazily initialises the CEL parser.
func celValidationEnv() (*cel.Env, error) {
	celEnvOnce.Do(func() {
		celEnv, celEnvErr = cel.NewEnv()
	})
	return celEnv, celEnvErr
}

// validateCELExpression ensures expression parses before persisting.
func validateCELExpression(expr string) error {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return errors.New("CEL expression cannot be empty")
	}
	env, err := celValidationEnv()
	if err != nil {
		return fmt.Errorf("cel parser unavailable: %w", err)
	}
	if _, issues := env.Parse(trimmed); issues != nil && issues.Err() != nil {
		return fmt.Errorf("invalid CEL expression: %w", issues.Err())
	}
	return nil
}

// apiErrorResponse standardises validation errors sent back to the UI.
type apiErrorResponse struct {
	Error               string            `json:"error"`
	Message             string            `json:"message"`
	FieldErrors         map[string]string `json:"field_errors,omitempty"`
	ExistingApplication *applicationDTO   `json:"existing_application,omitempty"`
}

// validationSuccessResponse wraps successful validation responses.
type validationSuccessResponse[T any] struct {
	Valid      bool `json:"valid"`
	Normalised T    `json:"normalised"`
}

// respondValidationError emits a structured validation error payload.
func respondValidationError(w http.ResponseWriter, status int, code, message string, fields fieldErrors, existing *applicationDTO) {
	resp := apiErrorResponse{
		Error:   code,
		Message: message,
	}
	if len(fields) > 0 {
		resp.FieldErrors = fields
	}
	if existing != nil {
		resp.ExistingApplication = existing
	}
	respondJSON(w, status, resp)
}

// respondValidationSuccess mirrors the happy-path result for validation endpoints.
func respondValidationSuccess[T any](w http.ResponseWriter, result T) {
	respondJSON(w, http.StatusOK, validationSuccessResponse[T]{
		Valid:      true,
		Normalised: result,
	})
}

// validateApplication runs server-side validation without persisting.
func (h Handler) validateApplication(w http.ResponseWriter, r *http.Request) {
	var body createApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	result, fieldErrs, existing, err := h.validateApplicationInput(r.Context(), applicationValidationInput(body), nil)
	if err != nil {
		h.Logger.Error("validate application payload", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to validate application")
		return
	}
	if len(fieldErrs) > 0 {
		code := errCodeValidationFailed
		message := "Application validation failed"
		if existing != nil {
			code = errCodeDuplicateIdentifier
			message = fmt.Sprintf("The identifier \"%s\" already belongs to \"%s\"", result.Identifier, existing.Name)
		}
		respondValidationError(w, http.StatusUnprocessableEntity, code, message, fieldErrs, existing)
		return
	}
	respondValidationSuccess(w, result)
}

// validateScopeForApplication ensures the target/action combination is valid.
func (h Handler) validateScopeForApplication(w http.ResponseWriter, r *http.Request) {
	appID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid application id")
		return
	}
	var body createScopeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	h.runScopeValidation(r.Context(), w, appID, body)
}

// runScopeValidation executes the shared validation logic after loading the rule.
func (h Handler) runScopeValidation(ctx context.Context, w http.ResponseWriter, appID uuid.UUID, body createScopeRequest) {
	rule, err := h.Store.GetRule(ctx, appID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondValidationError(w, http.StatusUnprocessableEntity, errCodeValidationFailed, "Scope validation failed", fieldErrors{
				"application_id": "application not found",
			}, nil)
			return
		}
		h.Logger.Error("validate scope application lookup", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to validate scope")
		return
	}
	result, fieldErrs, duplicate, err := h.validateScopeInput(ctx, rule, body)
	if err != nil {
		h.Logger.Error("validate scope payload", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to validate scope")
		return
	}
	if len(fieldErrs) > 0 {
		code := errCodeValidationFailed
		message := "Scope validation failed"
		if duplicate {
			code = errCodeDuplicateScope
			message = "The selected user or group already has an assignment for this application"
		}
		respondValidationError(w, http.StatusUnprocessableEntity, code, message, fieldErrs, nil)
		return
	}
	respondValidationSuccess(w, result)
}

// validateApplicationInput normalises + validates create/update payloads.
func (h Handler) validateApplicationInput(ctx context.Context, input applicationValidationInput, excludeID *uuid.UUID) (applicationValidationResult, fieldErrors, *applicationDTO, error) {
	errs := fieldErrors{}
	result := applicationValidationResult{}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		errs["name"] = "Application name is required"
	} else if utf8.RuneCountInString(name) > 100 {
		errs["name"] = "Name must be 100 characters or fewer"
	}
	result.Name = name

	ruleType := strings.ToUpper(strings.TrimSpace(input.RuleType))
	if ruleType == "" {
		errs["rule_type"] = "Rule type is required"
	} else if _, ok := applicationIdentifierValidators[ruleType]; !ok {
		errs["rule_type"] = "Unsupported rule type"
	}
	result.RuleType = ruleType

	identifier := strings.TrimSpace(input.Identifier)
	if identifier == "" {
		errs["identifier"] = "Identifier is required"
	} else if validator, ok := applicationIdentifierValidators[ruleType]; ok && !validator.regex.MatchString(identifier) {
		errs["identifier"] = validator.message
	}
	result.Identifier = identifier

	result.Description = strings.TrimSpace(input.Description)
	result.BlockMessage = strings.TrimSpace(input.BlockMessage)
	if utf8.RuneCountInString(result.BlockMessage) > 500 {
		errs["block_message"] = "Block message must be 500 characters or fewer"
	}

	result.CelEnabled = input.CelEnabled
	rawCel := strings.TrimSpace(input.CelExpression)
	if result.CelEnabled {
		if rawCel == "" {
			errs["cel_expression"] = "CEL expression is required when CEL mode is enabled"
		} else if err := validateCELExpression(rawCel); err != nil {
			errs["cel_expression"] = err.Error()
		}
		result.CelExpression = rawCel
	} else {
		result.CelExpression = ""
	}

	if len(errs) > 0 {
		return result, errs, nil, nil
	}

	rule, err := h.Store.GetRuleByTarget(ctx, identifier)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, nil, nil, nil
		}
		return result, nil, nil, err
	}
	if excludeID != nil && rule.ID == *excludeID {
		return result, nil, nil, nil
	}
	dto := mapApplication(rule)
	return result, fieldErrors{
		"identifier": "Identifier must be unique",
	}, &dto, nil
}

// validateScopeInput checks uniqueness and CEL compatibility for a target/action pair.
func (h Handler) validateScopeInput(ctx context.Context, rule sqlc.Rule, body createScopeRequest) (scopeValidationResult, fieldErrors, bool, error) {
	errs := fieldErrors{}
	result := scopeValidationResult{ApplicationID: rule.ID}

	meta, err := rules.ParseMetadata(rule.Metadata)
	if err != nil {
		h.Logger.Warn("parse rule metadata for scope validation", "rule", rule.ID, "err", err)
	}
	celEnabled := meta.CelEnabled

	targetType := strings.ToLower(strings.TrimSpace(body.TargetType))
	if targetType != "group" && targetType != "user" {
		errs["target_type"] = "target_type must be \"group\" or \"user\""
	} else {
		result.TargetType = targetType
	}

	targetIDStr := strings.TrimSpace(body.TargetID)
	if targetIDStr == "" {
		errs["target_id"] = "target_id is required"
	} else if parsed, err := uuid.Parse(targetIDStr); err == nil {
		result.TargetID = parsed
	} else {
		errs["target_id"] = "target_id must be a valid UUID"
	}

	action := strings.ToLower(strings.TrimSpace(body.Action))
	switch action {
	case string(rules.RuleActionAllow), string(rules.RuleActionBlock):
		if celEnabled {
			errs["action"] = "CEL-enabled applications must use the CEL action"
		} else {
			result.Action = action
		}
	case string(rules.RuleActionCel):
		if !celEnabled {
			errs["action"] = "action \"cel\" requires CEL mode to be enabled on the application"
		} else {
			result.Action = action
		}
	default:
		errs["action"] = "action must be allow, block, or cel"
	}

	if len(errs) > 0 {
		return result, errs, false, nil
	}

	if _, err := h.Store.GetRuleScopeByTarget(ctx, rule.ID, targetType, result.TargetID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, nil, false, nil
		}
		return result, nil, false, err
	}
	return result, fieldErrors{
		"target_id": "Selected target already has an assignment",
	}, true, nil
}
