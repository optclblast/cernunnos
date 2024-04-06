package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"cernunnos/internal/middleware"
	"cernunnos/internal/pkg/config"
	errs "cernunnos/internal/pkg/errors"
	"cernunnos/internal/pkg/logger"
	"cernunnos/internal/server/interface/controllers"
	productsRepo "cernunnos/internal/usecase/repository/products"
	storagesRepo "cernunnos/internal/usecase/repository/storages"

	"github.com/go-chi/chi"
)

type Server struct {
	*chi.Mux

	address            string
	log                *slog.Logger
	errorsHandler      errs.ErrorHandler
	storagesRepository storagesRepo.Repository
	productsRepository productsRepo.Repository
	controllers        *controllers.RootController
}

func newServer(
	cfg *config.Config,
	log *slog.Logger,
	storagesRepository storagesRepo.Repository,
	productsRepository productsRepo.Repository,
	rootController *controllers.RootController,
) *Server {
	s := &Server{
		address:            cfg.Address,
		log:                log,
		errorsHandler:      errs.NewErrorHandler(),
		storagesRepository: storagesRepository,
		productsRepository: productsRepository,
		controllers:        rootController,
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

	router.Route("/storages", func(r chi.Router) {
		r.Get("/", s.storages())
		r.Route("/{storage_id}/products", func(r chi.Router) {
			r.Get("/", s.storageProducts())
			r.Post("/", nil) // todo add a new product (-s)
			//r.Put("/{product_id}", nil)                // todo edit product info on a storage
			r.Delete("/{product_id}", nil)             // todo remove product from storage
			r.Post("/{product_id}/reservation", nil)   // todo reserve product
			r.Put("/{product_id}/reservation", nil)    // todo update reservation
			r.Delete("/{product_id}/reservation", nil) // todo delete reservation
		})
	})

	router.Route("/products", func(r chi.Router) {
		r.Get("/", s.products())             // todo list products available on a spicific storage
		r.Post("/new", nil)                  // todo add product
		r.Put("/{product_id}/edit", nil)     // todo update product info
		r.Post("/{product_id}/reserve", nil) // todo add reservation
	})

	router.Route("/reservations", func(r chi.Router) {
		r.Get("/", s.reservations())
		r.Post("/reserve", s.reserve())
		r.Delete("/cancel", s.cancelReservation())
		r.Delete("/release", s.releaseReservation())
	})

	router.Route("/shippings", func(r chi.Router) {
		r.Post("/{shipping_id}/confirm", nil)
		r.Post("/{shipping_id}/cancel", nil)
	})

	s.Mux = router
}

func (s *Server) responseError(
	ctx context.Context,
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

	w.WriteHeader(int(apiErr.Code))
	if _, err = w.Write(out); err != nil {
		log.Error("error write api error", logger.Err(err))
	}
}

func (s *Server) response(
	ctx context.Context,
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
