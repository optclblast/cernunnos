package server

import (
	"cernunnos/internal/pkg/dto"
	"cernunnos/internal/server/interface/controllers/reservation"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const requestTimeout time.Duration = 10 * time.Second

func (s *Server) storages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const methodName = "storages"

		ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
		defer cancel()

		log := s.log.WithGroup(methodName)

		var err error

		defer func() {
			if err != nil {
				s.responseError(ctx, w, err, methodName)
			}
		}()

		request, err := buildRequest[dto.StoragesRequest](r)
		if err != nil {
			err = fmt.Errorf("error build storages request. %w", err)
			return
		}

		log.Debug("request", slog.Any("dto", request))

		response, err := s.controllers.StorageController.Storages(ctx, request)
		if err != nil {
			return
		}

		s.response(ctx, w, http.StatusOK, response)
	}
}

func (s *Server) storageProducts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const methodName = "storage_products"

		ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
		defer cancel()

		log := s.log.WithGroup(methodName)

		var err error

		defer func() {
			if err != nil {
				s.responseError(ctx, w, err, methodName)
			}
		}()

		request, err := buildRequest[dto.StorageProductsRequest](r)
		if err != nil {
			err = fmt.Errorf("error build storage_products request. %w", err)
			return
		}

		log.Debug("request", slog.Any("dto", request))

		response, err := s.controllers.ProductController.StorageProducts(ctx, request)
		if err != nil {
			return
		}

		s.response(ctx, w, http.StatusOK, response)
	}
}

func (s *Server) products() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const methodName = "products"

		ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
		defer cancel()

		log := s.log.WithGroup(methodName)

		var err error

		defer func() {
			if err != nil {
				s.responseError(ctx, w, err, methodName)
			}
		}()

		request, err := buildRequest[dto.ProductsRequest](r)
		if err != nil {
			err = fmt.Errorf("error build products request. %w", err)
			return
		}

		log.Debug("request", slog.Any("dto", request))

		response, err := s.controllers.ProductController.Products(ctx, request)
		if err != nil {
			return
		}

		s.response(ctx, w, http.StatusOK, response)
	}
}

func (s *Server) reservations() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const methodName = "reservations"

		ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
		defer cancel()

		log := s.log.WithGroup(methodName)

		var err error

		defer func() {
			if err != nil {
				s.responseError(ctx, w, err, methodName)
			}
		}()

		request, err := buildRequest[dto.ReservationsRequest](r)
		if err != nil {
			err = fmt.Errorf("error build products request. %w", err)
			return
		}

		log.Debug("request", slog.Any("dto", request))

		reservations, err := s.controllers.ReservationController.Reservations(
			ctx,
			reservation.ReservationsParams{
				StorageId:  request.StorageId,
				ShippingId: request.ShippingId,
				ProductId:  request.ProductId,
			},
		)
		if err != nil {
			err = fmt.Errorf("error fetch reservations. %w", err)
			return
		}

	}
}
