package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/rules"
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
		regex:   regexp.MustCompile(`^[A-Z0-9]{10}:[a-zA-Z0-9.-]+$`),
		message: "Signing IDs must follow TEAMID:bundle.identifier format",
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

type applicationValidationResult struct {
	Name        string `json:"name"`
	RuleType    string `json:"rule_type"`
	Identifier  string `json:"identifier"`
	Description string `json:"description,omitempty"`
}

type scopeValidationResult struct {
	ApplicationID uuid.UUID `json:"application_id"`
	TargetType    string    `json:"target_type"`
	TargetID      uuid.UUID `json:"target_id"`
	Action        string    `json:"action"`
}

type apiErrorResponse struct {
	Error               string            `json:"error"`
	Message             string            `json:"message"`
	FieldErrors         map[string]string `json:"field_errors,omitempty"`
	ExistingApplication *applicationDTO   `json:"existing_application,omitempty"`
}

type validationSuccessResponse[T any] struct {
	Valid      bool `json:"valid"`
	Normalized T    `json:"normalized"`
}

type scopeValidationRequest struct {
	ApplicationID string `json:"application_id"`
	TargetType    string `json:"target_type"`
	TargetID      string `json:"target_id"`
	Action        string `json:"action"`
}

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

func respondValidationSuccess[T any](w http.ResponseWriter, result T) {
	respondJSON(w, http.StatusOK, validationSuccessResponse[T]{
		Valid:      true,
		Normalized: result,
	})
}

func (h Handler) validateApplication(w http.ResponseWriter, r *http.Request) {
	var body createApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	result, fieldErrs, existing, err := h.validateApplicationInput(r.Context(), body)
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

func (h Handler) validateScope(w http.ResponseWriter, r *http.Request) {
	var body scopeValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	appIDStr := strings.TrimSpace(body.ApplicationID)
	if appIDStr == "" {
		respondValidationError(w, http.StatusUnprocessableEntity, errCodeValidationFailed, "Scope validation failed", fieldErrors{
			"application_id": "application_id is required",
		}, nil)
		return
	}
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		respondValidationError(w, http.StatusUnprocessableEntity, errCodeValidationFailed, "Scope validation failed", fieldErrors{
			"application_id": "application_id must be a valid UUID",
		}, nil)
		return
	}
	if _, err := h.Store.GetRule(r.Context(), appID); err != nil {
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
	createBody := createScopeRequest{
		TargetType: body.TargetType,
		TargetID:   body.TargetID,
		Action:     body.Action,
	}
	result, fieldErrs, duplicate, err := h.validateScopeInput(r.Context(), appID, createBody)
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

func (h Handler) validateApplicationInput(ctx context.Context, body createApplicationRequest) (applicationValidationResult, fieldErrors, *applicationDTO, error) {
	errs := fieldErrors{}
	result := applicationValidationResult{}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		errs["name"] = "Application name is required"
	} else if utf8.RuneCountInString(name) > 100 {
		errs["name"] = "Name must be 100 characters or fewer"
	}
	result.Name = name

	ruleType := strings.ToUpper(strings.TrimSpace(body.RuleType))
	if ruleType == "" {
		errs["rule_type"] = "Rule type is required"
	} else if _, ok := applicationIdentifierValidators[ruleType]; !ok {
		errs["rule_type"] = "Unsupported rule type"
	}
	result.RuleType = ruleType

	identifier := strings.TrimSpace(body.Identifier)
	if identifier == "" {
		errs["identifier"] = "Identifier is required"
	} else if validator, ok := applicationIdentifierValidators[ruleType]; ok && !validator.regex.MatchString(identifier) {
		errs["identifier"] = validator.message
	}
	result.Identifier = identifier

	result.Description = strings.TrimSpace(body.Description)

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
	dto := mapApplication(rule)
	return result, fieldErrors{
		"identifier": "Identifier must be unique",
	}, &dto, nil
}

func (h Handler) validateScopeInput(ctx context.Context, appID uuid.UUID, body createScopeRequest) (scopeValidationResult, fieldErrors, bool, error) {
	errs := fieldErrors{}
	result := scopeValidationResult{ApplicationID: appID}

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
	if action != string(rules.RuleActionAllow) && action != string(rules.RuleActionBlock) {
		errs["action"] = "action must be allow or block"
	} else {
		result.Action = action
	}

	if len(errs) > 0 {
		return result, errs, false, nil
	}

	if _, err := h.Store.GetRuleScopeByTarget(ctx, appID, targetType, result.TargetID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, nil, false, nil
		}
		return result, nil, false, err
	}
	return result, fieldErrors{
		"target_id": "Selected target already has an assignment",
	}, true, nil
}
