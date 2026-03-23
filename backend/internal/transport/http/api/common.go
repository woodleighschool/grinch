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
			w.WriteHeader(http.StatusBadRequest)
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

// writeValidationErrors writes a React Admin compatible validation error body.
// Shape: {"errors": {"root": {"serverError": "..."}, "field": "message", ...}}.
func writeValidationErrors(w http.ResponseWriter, root string, fields []domain.FieldError) {
	errs := make(map[string]any, len(fields)+1)
	if root != "" {
		errs["root"] = map[string]string{"serverError": root}
	}
	for _, fe := range fields {
		errs[fe.Field] = fe.Message
	}
	writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
}

func writeError(w http.ResponseWriter, err error) {
	var validationErr *domain.ValidationError

	var badReqErr badRequestError

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		w.WriteHeader(http.StatusNotFound)
		return

	case errors.Is(err, domain.ErrGroupReadOnly):
		w.WriteHeader(http.StatusForbidden)
		return

	case errors.Is(err, domain.ErrInvalidSort), errors.As(err, &badReqErr):
		w.WriteHeader(http.StatusBadRequest)
		return

	case errors.As(err, &validationErr):
		writeValidationErrors(w, validationErr.Detail, validationErr.FieldErrors)
		return
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			writeValidationErrors(w, "Resource already exists.", nil)
			return

		case pgerrcode.ForeignKeyViolation:
			writeValidationErrors(w, "Referenced resource does not exist.", nil)
			return
		}
	}

	w.WriteHeader(http.StatusInternalServerError)
}

func (e badRequestError) Error() string {
	return string(e)
}
