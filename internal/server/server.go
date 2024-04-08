package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"cernunnos/internal/middleware"
	"cernunnos/internal/pkg/config"
	errs "cernunnos/internal/pkg/errors"
	"cernunnos/internal/pkg/logger"
	"cernunnos/internal/server/interface/controllers"

	"github.com/go-chi/render"

	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	*chi.Mux

	address       string
	log           *slog.Logger
	errorsHandler errs.ErrorHandler
	controllers   *controllers.RootController
}

func newServer(
	cfg *config.Config,
	log *slog.Logger,
	rootController *controllers.RootController,
) *Server {
	s := &Server{
		address:       cfg.Address,
		log:           log,
		errorsHandler: errs.NewErrorHandler(),
		controllers:   rootController,
	}

	s.initializeRouter()

	return s
}

func (s *Server) Start() error {
	s.log.Info("starting cernunnos server", slog.String("address", s.address))

	if err := http.ListenAndServe(s.address, s); err != nil {
		return fmt.Errorf("error listen to %s. %w", s.address, err)
	}

	return nil
}

func (s *Server) initializeRouter() {
	router := chi.NewRouter()
	middlewareBuilder := middleware.NewMiddlewareBuilder(s.log.WithGroup("middleware"))

	router.Use(middlewareBuilder.Recovery)
	router.Use(chimw.RequestID)
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Route("/storages", func(r chi.Router) {
		r.Get("/", s.handle(s.storages, "storages"))
		r.Route("/{storage_id}/products", func(r chi.Router) {
			r.Get("/", s.handle(s.storageProducts, "storage_products"))
		})
	})

	router.Route("/products", func(r chi.Router) {
		r.Get("/", s.handle(s.products, "products"))
	})

	router.Route("/reservations", func(r chi.Router) {
		r.Get("/", s.handle(s.reservations, "reservations"))
		r.Post("/new", s.handle(s.reserveProduct, "reserve_product"))
		r.Delete("/cancel", s.handle(s.cancelProductReservation, "cancel_reservation"))
		r.Delete("/release", s.handle(s.releaseProductReservation, "release_reservation"))
	})

	s.Mux = router
}

const requestTimeout time.Duration = 1000000 * time.Second

type handlerFunc func(ctx context.Context, r *http.Request) ([]byte, error)

func (s *Server) handle(h handlerFunc, mathodName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
		defer cancel()

		resp, err := h(ctx, r)
		if err != nil {
			s.responseError(w, err, mathodName)

			return
		}

		s.response(w, http.StatusOK, resp)
	}
}

func (s *Server) responseError(
	w http.ResponseWriter,
	e error,
	methodName string,
) {
	log := s.log.WithGroup("api_error").With(slog.String("method_name", methodName))
	apiErr := s.errorsHandler.Handle(e)

	log.Error("error", logger.Err(e))

	out, err := json.Marshal(&apiErr)
	if err != nil {
		log.Error("error marshal api error", logger.Err(err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(int(apiErr.Code))

	if _, err = w.Write(out); err != nil {
		log.Error("error write api error", logger.Err(err))
	}
}

func (s *Server) response(
	w http.ResponseWriter,
	code int,
	response []byte,
) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	if _, err := w.Write(response); err != nil {
		s.log.Error("error write response", logger.Err(err))
	}
}

func buildRequest[R any](r *http.Request) (*R, error) {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error read request body. %w", err), errs.ErrorBadRequest)
	}

	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			err = errors.Join(fmt.Errorf("error close request body. %w", closeErr), err)
		}
	}()

	var request R

	if err = json.Unmarshal(rawBody, &request); err != nil {
		return nil, errors.Join(
			fmt.Errorf("error unmarshal request body. %w", err),
			errs.ErrorBadRequest,
		)
	}

	return &request, nil
}
