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

type problemSpec struct {
	Type        string
	Title       string
	Code        string
	Detail      string
	FieldErrors []domain.FieldError
}

type apiErrorOptions struct {
	NotFoundMessage string
}

type badRequestError string

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

func (s *Server) RegisterRoutes(r chi.Router) {
	_ = HandlerWithOptions(s, ChiServerOptions{
		BaseRouter: r,
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) {
			writeProblem(w, http.StatusBadRequest, problemSpec{
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
	var resolvedLimit int32
	if limit != nil {
		resolvedLimit = *limit
	}

	var resolvedOffset int32
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
		IDs:    cloneUUIDs(ids),
		Limit:  resolvedLimit,
		Offset: resolvedOffset,
		Search: optionalString(search),
		Sort:   optionalString(sort),
		Order:  optionalString(order),
	}, nil
}

func optionalString[T ~string](v *T) string {
	if v == nil {
		return ""
	}
	return string(*v)
}

func cloneUUIDs(v *IdsFilter) []uuid.UUID {
	if v == nil {
		return nil
	}

	out := make([]uuid.UUID, len(*v))
	copy(out, *v)

	return out
}

func cloneBools(v *EnabledFilter) []bool {
	if v == nil {
		return nil
	}

	out := make([]bool, len(*v))
	copy(out, *v)

	return out
}

func parseOptionalValues[T any, V ~string](values *[]V, parse func(string) (T, error)) ([]T, error) {
	if values == nil {
		return nil, nil
	}

	out := make([]T, 0, len(*values))
	for _, v := range *values {
		parsed, err := parse(string(v))
		if err != nil {
			return nil, badRequestError(err.Error())
		}
		out = append(out, parsed)
	}

	return out, nil
}

func decodeJSONBody(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, maxJSONBodyBytes))

	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return badRequestError("request body is required")
		}
		return badRequestError("request body is invalid")
	}

	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return badRequestError("request body must contain a single JSON object")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}

func writeProblem(w http.ResponseWriter, statusCode int, spec problemSpec) {
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

	writeJSON(w, statusCode, problem)
}

func writeClassifiedError(w http.ResponseWriter, err error, opts apiErrorOptions) {
	var validationErr *domain.ValidationError

	switch {
	case opts.NotFoundMessage != "" && isNotFound(err):
		writeProblem(w, http.StatusNotFound, problemSpec{
			Type:   "urn:grinch:problem:not-found",
			Title:  "Not found",
			Code:   "not_found",
			Detail: opts.NotFoundMessage,
		})
		return

	case errors.Is(err, domain.ErrGroupReadOnly):
		writeProblem(w, http.StatusForbidden, problemSpec{
			Type:   "urn:grinch:problem:forbidden",
			Title:  "Forbidden",
			Code:   "forbidden",
			Detail: "Entra groups are read-only.",
		})
		return

	case errors.Is(err, domain.ErrInvalidSort), isBadRequestError(err):
		writeProblem(w, http.StatusBadRequest, problemSpec{
			Type:   "urn:grinch:problem:invalid-request",
			Title:  "Invalid request",
			Code:   "invalid_request",
			Detail: err.Error(),
		})
		return

	case errors.As(err, &validationErr):
		writeProblem(w, http.StatusUnprocessableEntity, problemSpec{
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
			writeProblem(w, http.StatusConflict, problemSpec{
				Type:   "urn:grinch:problem:conflict",
				Title:  "Conflict",
				Code:   "conflict",
				Detail: "Resource already exists.",
			})
			return

		case pgerrcode.ForeignKeyViolation:
			writeProblem(w, http.StatusUnprocessableEntity, problemSpec{
				Type:   "urn:grinch:problem:validation-error",
				Title:  "Validation failed",
				Code:   "validation_error",
				Detail: "Referenced resource does not exist.",
			})
			return
		}
	}

	writeProblem(w, http.StatusInternalServerError, problemSpec{
		Type:   "urn:grinch:problem:internal-error",
		Title:  "Internal server error",
		Code:   "internal_error",
		Detail: "An internal server error occurred.",
	})
}

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func isBadRequestError(err error) bool {
	var reqErr badRequestError
	return errors.As(err, &reqErr)
}

func (e badRequestError) Error() string {
	return string(e)
}
