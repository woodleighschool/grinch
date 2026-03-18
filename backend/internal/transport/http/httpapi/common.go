package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	appgroupmemberships "github.com/woodleighschool/grinch/internal/app/groupmemberships"
	apprules "github.com/woodleighschool/grinch/internal/app/rules"
	"github.com/woodleighschool/grinch/internal/domain"
	pgutil "github.com/woodleighschool/grinch/internal/store/postgres/shared"
)

const (
	maxJSONBodyBytes = 1 << 20
)

type Server struct {
	admin            AdminService
	rules            RuleService
	groupMemberships GroupMembershipService
}

type AdminService interface {
	ListUsers(context.Context, domain.UserListOptions) ([]domain.User, int32, error)
	GetUser(context.Context, uuid.UUID) (domain.User, error)
	ListGroups(context.Context, domain.GroupListOptions) ([]domain.Group, int32, error)
	CreateLocalGroup(context.Context, string, string) (domain.Group, error)
	GetGroup(context.Context, uuid.UUID) (domain.Group, error)
	UpdateGroup(context.Context, uuid.UUID, string, string) (domain.Group, error)
	DeleteGroup(context.Context, uuid.UUID) error
	ListMachines(context.Context, domain.MachineListOptions) ([]domain.MachineSummary, int32, error)
	GetMachine(context.Context, uuid.UUID) (domain.Machine, error)
	DeleteMachine(context.Context, uuid.UUID) error
	ListExecutables(context.Context, domain.ExecutableListOptions) ([]domain.ExecutableSummary, int32, error)
	GetExecutable(context.Context, uuid.UUID) (domain.Executable, error)
	ListExecutionEvents(
		context.Context,
		domain.ExecutionEventListOptions,
	) ([]domain.ExecutionEventSummary, int32, error)
	GetExecutionEvent(context.Context, uuid.UUID) (domain.ExecutionEvent, error)
	DeleteExecutionEvent(context.Context, uuid.UUID) error
	ListFileAccessEvents(
		context.Context,
		domain.FileAccessEventListOptions,
	) ([]domain.FileAccessEventSummary, int32, error)
	GetFileAccessEvent(context.Context, uuid.UUID) (domain.FileAccessEvent, error)
	DeleteFileAccessEvent(context.Context, uuid.UUID) error
}

type RuleService interface {
	ListRules(context.Context, domain.RuleListOptions) ([]domain.RuleSummary, int32, error)
	GetRule(context.Context, uuid.UUID) (domain.Rule, error)
	CreateRule(context.Context, apprules.WriteInput) (domain.Rule, error)
	UpdateRule(context.Context, uuid.UUID, apprules.WriteInput) (domain.Rule, error)
	DeleteRule(context.Context, uuid.UUID) error
}

type GroupMembershipService interface {
	ListGroupMemberships(
		context.Context,
		domain.GroupMembershipListOptions,
	) ([]domain.GroupMembership, int32, error)
	GetGroupMembership(context.Context, uuid.UUID) (domain.GroupMembership, error)
	CreateGroupMembership(context.Context, appgroupmemberships.CreateInput) (domain.GroupMembership, error)
	DeleteGroupMembership(context.Context, uuid.UUID) error
}

func New(
	adminService AdminService,
	ruleService RuleService,
	groupMembershipService GroupMembershipService,
) *Server {
	return &Server{
		admin:            adminService,
		rules:            ruleService,
		groupMemberships: groupMembershipService,
	}
}

func (handler *Server) RegisterRoutes(router chi.Router) {
	_ = HandlerWithOptions(handler, ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(writer http.ResponseWriter, _ *http.Request, _ error) {
			writeProblem(writer, http.StatusBadRequest, problemSpec{
				Type:   "urn:grinch:problem:invalid-request",
				Title:  "Invalid request",
				Code:   "invalid_request",
				Detail: "Request parameters are invalid.",
			})
		},
	})
}

func parsePagination(limit *int32, offset *int32) (int32, int32, error) {
	resolvedLimit := int32(0)
	resolvedOffset := int32(0)

	if limit != nil {
		resolvedLimit = *limit
	}
	if offset != nil {
		resolvedOffset = *offset
	}

	switch {
	case limit != nil && resolvedLimit < 1:
		return 0, 0, badRequestError("limit must be >= 1")
	case resolvedOffset < 0:
		return 0, 0, badRequestError("offset must be >= 0")
	default:
		return resolvedLimit, resolvedOffset, nil
	}
}

func optionalString[T ~string](value *T) string {
	if value == nil {
		return ""
	}
	return string(*value)
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func mapSlice[T any, U any](items []T, mapper func(T) (U, error)) ([]U, error) {
	mapped := make([]U, 0, len(items))
	for _, item := range items {
		output, err := mapper(item)
		if err != nil {
			return nil, err
		}
		mapped = append(mapped, output)
	}
	return mapped, nil
}

func mapSliceValue[T any, U any](items []T, mapper func(T) U) []U {
	mapped := make([]U, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, mapper(item))
	}
	return mapped
}

func parseListOptions[T ~string, O ~string](
	limit *int32,
	offset *int32,
	search *T,
	sort *Sort,
	order *O,
) (domain.ListOptions, error) {
	resolvedLimit, resolvedOffset, err := parsePagination(limit, offset)
	if err != nil {
		return domain.ListOptions{}, err
	}

	return domain.ListOptions{
		Limit:  resolvedLimit,
		Offset: resolvedOffset,
		Search: optionalString(search),
		Sort:   optionalString(sort),
		Order:  optionalString(order),
	}, nil
}

func decodeJSONBody(request *http.Request, dst any) error {
	decoder := json.NewDecoder(io.LimitReader(request.Body, maxJSONBodyBytes))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return badRequestError("request body is required")
		}
		return badRequestError("request body is invalid")
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return badRequestError("request body must contain a single JSON object")
	}

	return nil
}

func writeJSON(writer http.ResponseWriter, statusCode int, payload any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(payload)
}

type problemSpec struct {
	Type        string
	Title       string
	Code        string
	Detail      string
	FieldErrors []domain.FieldError
}

func writeProblem(writer http.ResponseWriter, statusCode int, spec problemSpec) {
	problem := Problem{
		Type:   spec.Type,
		Title:  spec.Title,
		Status: safeInt32(statusCode),
		Detail: spec.Detail,
		Code:   spec.Code,
	}
	if len(spec.FieldErrors) > 0 {
		fieldErrors := make([]FieldError, 0, len(spec.FieldErrors))
		for _, fieldErr := range spec.FieldErrors {
			mapped := FieldError{
				Field:   fieldErr.Field,
				Message: fieldErr.Message,
			}
			if fieldErr.Code != "" {
				code := fieldErr.Code
				mapped.Code = &code
			}
			fieldErrors = append(fieldErrors, mapped)
		}
		problem.FieldErrors = &fieldErrors
	}
	writeJSON(writer, statusCode, problem)
}

func safeInt32(value int) int32 {
	if value < math.MinInt32 {
		return math.MinInt32
	}
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(value)
}

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

type apiErrorOptions struct {
	NotFoundMessage string
}

func writeClassifiedError(writer http.ResponseWriter, err error, options apiErrorOptions) {
	var validationErr *domain.ValidationError
	switch {
	case options.NotFoundMessage != "" && isNotFound(err):
		writeProblem(writer, http.StatusNotFound, problemSpec{
			Type:   "urn:grinch:problem:not-found",
			Title:  "Not found",
			Code:   "not_found",
			Detail: options.NotFoundMessage,
		})
		return
	case errors.Is(err, domain.ErrGroupReadOnly):
		writeProblem(writer, http.StatusForbidden, problemSpec{
			Type:   "urn:grinch:problem:forbidden",
			Title:  "Forbidden",
			Code:   "forbidden",
			Detail: "Entra groups are read-only.",
		})
		return
	case errors.Is(err, pgutil.ErrInvalidSort), isBadRequestError(err):
		writeProblem(writer, http.StatusBadRequest, problemSpec{
			Type:   "urn:grinch:problem:invalid-request",
			Title:  "Invalid request",
			Code:   "invalid_request",
			Detail: err.Error(),
		})
		return
	case errors.As(err, &validationErr):
		writeProblem(writer, http.StatusUnprocessableEntity, problemSpec{
			Type:        "urn:grinch:problem:validation-error",
			Title:       "Validation failed",
			Code:        validationErr.Code,
			Detail:      validationErr.Detail,
			FieldErrors: validationErr.FieldErrors,
		})
		return
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			writeProblem(writer, http.StatusConflict, problemSpec{
				Type:   "urn:grinch:problem:conflict",
				Title:  "Conflict",
				Code:   "conflict",
				Detail: "Resource already exists.",
			})
			return
		case pgerrcode.ForeignKeyViolation:
			writeProblem(writer, http.StatusUnprocessableEntity, problemSpec{
				Type:   "urn:grinch:problem:validation-error",
				Title:  "Validation failed",
				Code:   "validation_error",
				Detail: "Referenced resource does not exist.",
			})
			return
		}
	}

	writeProblem(writer, http.StatusInternalServerError, problemSpec{
		Type:   "urn:grinch:problem:internal-error",
		Title:  "Internal server error",
		Code:   "internal_error",
		Detail: "An internal server error occurred.",
	})
}

func isBadRequestError(err error) bool {
	var requestErr badRequestError
	return errors.As(err, &requestErr)
}

type badRequestError string

func (err badRequestError) Error() string {
	return string(err)
}

func toStringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
