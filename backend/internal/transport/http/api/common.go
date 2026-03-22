package apihttp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	appfileaccessevents "github.com/woodleighschool/grinch/internal/app/fileaccessevents"
	appgroups "github.com/woodleighschool/grinch/internal/app/groups"
	appmemberships "github.com/woodleighschool/grinch/internal/app/memberships"
	apprules "github.com/woodleighschool/grinch/internal/app/rules"
	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/postgres"
)

const maxJSONBodyBytes = 1 << 20

type Server struct {
	store            *postgres.Store
	groups           *appgroups.Service
	fileAccessEvents *appfileaccessevents.Service
	memberships      *appmemberships.Service
	rules            *apprules.Service
}

func New(
	store *postgres.Store,
	groups *appgroups.Service,
	fileAccessEvents *appfileaccessevents.Service,
	rules *apprules.Service,
	memberships *appmemberships.Service,
) *Server {
	return &Server{
		store:            store,
		groups:           groups,
		fileAccessEvents: fileAccessEvents,
		memberships:      memberships,
		rules:            rules,
	}
}

func (handler *Server) RegisterRoutes(router chi.Router) {
	_ = HandlerWithOptions(handler, ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(writer http.ResponseWriter, _ *http.Request, _ error) {
			WriteProblem(writer, http.StatusBadRequest, ProblemSpec{
				Type:   "urn:grinch:problem:invalid-request",
				Title:  "Invalid request",
				Code:   "invalid_request",
				Detail: "Request parameters are invalid.",
			})
		},
	})
}

func parseListOptions[T ~string, O ~string](
	limit *int32,
	offset *int32,
	search *T,
	sort *Sort,
	order *O,
	ids *IdsFilter,
) (domain.ListOptions, error) {
	resolvedLimit := int32(0)
	if limit != nil {
		resolvedLimit = *limit
	}
	resolvedOffset := int32(0)
	if offset != nil {
		resolvedOffset = *offset
	}
	switch {
	case limit != nil && resolvedLimit < 1:
		return domain.ListOptions{}, badRequestError("limit must be >= 1")
	case resolvedOffset < 0:
		return domain.ListOptions{}, badRequestError("offset must be >= 0")
	}
	return domain.ListOptions{
		IDs:    optionalUUIDs(ids),
		Limit:  resolvedLimit,
		Offset: resolvedOffset,
		Search: optionalString(search),
		Sort:   optionalString(sort),
		Order:  optionalString(order),
	}, nil
}

func optionalString[T ~string](value *T) string {
	if value == nil {
		return ""
	}
	return string(*value)
}

func optionalUUIDs(values *IdsFilter) []uuid.UUID {
	if values == nil {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(*values))
	ids = append(ids, *values...)
	return ids
}

func optionalBools(values *EnabledFilter) []bool {
	if values == nil {
		return nil
	}
	result := make([]bool, 0, len(*values))
	result = append(result, (*values)...)
	return result
}

func parseOptionalValues[T any, V ~string](values *[]V, parse func(string) (T, error)) ([]T, error) {
	if values == nil {
		return nil, nil
	}
	result := make([]T, 0, len(*values))
	for _, value := range *values {
		parsed, err := parse(string(value))
		if err != nil {
			return nil, badRequestError(err.Error())
		}
		result = append(result, parsed)
	}
	return result, nil
}

func decodeJSONBody(request *http.Request, dst any) error {
	decoder := json.NewDecoder(io.LimitReader(request.Body, maxJSONBodyBytes))
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

type ProblemSpec struct {
	Type        string
	Title       string
	Code        string
	Detail      string
	FieldErrors []domain.FieldError
}

func WriteProblem(writer http.ResponseWriter, statusCode int, spec ProblemSpec) {
	problem := Problem{
		Type:   spec.Type,
		Title:  spec.Title,
		Status: statusCode,
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
		WriteProblem(writer, http.StatusNotFound, ProblemSpec{
			Type:   "urn:grinch:problem:not-found",
			Title:  "Not found",
			Code:   "not_found",
			Detail: options.NotFoundMessage,
		})
		return
	case errors.Is(err, domain.ErrGroupReadOnly):
		WriteProblem(writer, http.StatusForbidden, ProblemSpec{
			Type:   "urn:grinch:problem:forbidden",
			Title:  "Forbidden",
			Code:   "forbidden",
			Detail: "Entra groups are read-only.",
		})
		return
	case errors.Is(err, domain.ErrInvalidSort), isBadRequestError(err):
		WriteProblem(writer, http.StatusBadRequest, ProblemSpec{
			Type:   "urn:grinch:problem:invalid-request",
			Title:  "Invalid request",
			Code:   "invalid_request",
			Detail: err.Error(),
		})
		return
	case errors.As(err, &validationErr):
		WriteProblem(writer, http.StatusUnprocessableEntity, ProblemSpec{
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
			WriteProblem(writer, http.StatusConflict, ProblemSpec{
				Type:   "urn:grinch:problem:conflict",
				Title:  "Conflict",
				Code:   "conflict",
				Detail: "Resource already exists.",
			})
			return
		case pgerrcode.ForeignKeyViolation:
			WriteProblem(writer, http.StatusUnprocessableEntity, ProblemSpec{
				Type:   "urn:grinch:problem:validation-error",
				Title:  "Validation failed",
				Code:   "validation_error",
				Detail: "Referenced resource does not exist.",
			})
			return
		}
	}

	WriteProblem(writer, http.StatusInternalServerError, ProblemSpec{
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
