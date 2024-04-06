package middleware

import (
	"cernunnos/internal/pkg/logger"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	chi "github.com/go-chi/chi/v5/middleware"
)

type middlewareBuilder struct {
	log *slog.Logger
}

func NewMiddlewareBuilder(log *slog.Logger) *middlewareBuilder {
	return &middlewareBuilder{log: log}
}

func (m *middlewareBuilder) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				reqID, ctxErr := requestId(r.Context())
				if ctxErr != nil {
					m.log.Error("incomplete request context", logger.Err(ctxErr))
				}

				m.log.Error(
					"recovered from panic",
					slog.String("request id", reqID),
					slog.Any("error", slog.AnyValue(err)),
				)

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func requestId(ctx context.Context) (string, error) {
	if v := ctx.Value(chi.RequestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s, nil
		}
	}

	return "", fmt.Errorf("request id not passed through context")
}
