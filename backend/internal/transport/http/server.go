package httprouter

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const (
	readTimeout     = 15 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 10 * time.Second
)

// Server wraps http.Server and manages graceful shutdown.
type Server struct {
	server *http.Server
	log    *slog.Logger
}

// NewServer constructs a Server with standard timeouts.
func NewServer(handler http.Handler, port int, log *slog.Logger) *Server {
	return &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		log: log,
	}
}

// Run starts the HTTP server and blocks until shutdown or failure.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.log.InfoContext(ctx, "http server starting", "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		s.log.InfoContext(ctx, "http server shutting down")
		return s.server.Shutdown(shutdownCtx)

	case err := <-errCh:
		s.log.ErrorContext(ctx, "http server failed", "error", err)
		return err
	}
}
