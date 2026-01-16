package helpers

import (
	"context"
	"encoding/json"
	"net/http"

	coreerrors "github.com/woodleighschool/grinch/internal/core/errors"
	"github.com/woodleighschool/grinch/internal/logging"
	httpstatus "github.com/woodleighschool/grinch/internal/transport/http/status"
)

// WriteJSON writes v as a JSON response with the given status code.
func WriteJSON(ctx context.Context, w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if v == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(v); err != nil {
		logging.FromContext(ctx).WarnContext(ctx, "encode json", "error", err)
	}
}

// WriteError writes an error response in the format expected by react-admin.
func WriteError(ctx context.Context, w http.ResponseWriter, log logging.Logger, err error, msg string) {
	code := httpstatus.Status(err)

	if code >= http.StatusInternalServerError {
		log.ErrorContext(ctx, msg, "error", err)
	} else {
		log.WarnContext(ctx, msg, "error", err)
	}

	resp := struct {
		Errors map[string]string `json:"errors"`
	}{
		Errors: make(map[string]string),
	}

	if derr, ok := coreerrors.As(err); ok {
		switch {
		case len(derr.Fields) > 0:
			resp.Errors = derr.Fields
		case derr.Message != "":
			resp.Errors["root"] = derr.Message
		default:
			resp.Errors["root"] = http.StatusText(code)
		}
	} else {
		resp.Errors["root"] = http.StatusText(code)
	}

	WriteJSON(ctx, w, code, resp)
}
