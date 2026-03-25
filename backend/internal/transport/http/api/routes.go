package apihttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) RegisterRoutes(r chi.Router) {
	_ = HandlerWithOptions(s, ChiServerOptions{
		BaseRouter: r,
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) {
			w.WriteHeader(http.StatusBadRequest)
		},
	})
}
