package server

import (
	"cernunnos/internal/pkg/dto"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) storages(ctx context.Context, r *http.Request) ([]byte, error) {
	request, err := buildRequest[dto.StoragesRequest](r)
	if err != nil {
		return nil, fmt.Errorf("error build storages request. %w", err)
	}

	response, err := s.controllers.StorageController.Storages(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error fetch storages. %w", err)
	}

	return response, nil
}

func (s *Server) storageProducts(ctx context.Context, r *http.Request) ([]byte, error) {
	const methodName = "storage_products"
	log := s.log.WithGroup(methodName)

	request, err := buildRequest[dto.StorageProductsRequest](r)
	if err != nil {
		return nil, fmt.Errorf("error build storage_products request. %w", err)
	}

	storageId := chi.URLParam(r, "storage_id")

	request.StorageId = storageId

	log.Debug("request", slog.Any("dto", request))

	response, err := s.controllers.ProductController.StorageProducts(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error fetch storage products. %w", err)
	}

	return response, nil
}

func (s *Server) products(ctx context.Context, r *http.Request) ([]byte, error) {
	const methodName = "products"

	log := s.log.WithGroup(methodName)

	request, err := buildRequest[dto.ProductsRequest](r)
	if err != nil {
		return nil, fmt.Errorf("error build products request. %w", err)
	}

	log.Debug("request", slog.Any("dto", request))

	response, err := s.controllers.ProductController.Products(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error fetch products. %w", err)
	}

	return response, nil
}

func (s *Server) reservations(ctx context.Context, r *http.Request) ([]byte, error) {
	const methodName = "reservations"

	log := s.log.WithGroup(methodName)

	request, err := buildRequest[dto.ReservationsRequest](r)
	if err != nil {
		return nil, fmt.Errorf("error build products request. %w", err)
	}

	log.Debug("request", slog.Any("dto", request))

	response, err := s.controllers.ReservationController.Reservations(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error fetch reservations. %w", err)
	}

	return response, nil
}

func (s *Server) reserveProduct(ctx context.Context, r *http.Request) ([]byte, error) {
	request, err := buildRequest[dto.ReserveRequest](r)
	if err != nil {
		return nil, fmt.Errorf("error build reserve request. %w", err)
	}

	response, err := s.controllers.ReservationController.Reserve(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error reserve product. %w", err)
	}

	return response, nil
}

func (s *Server) cancelProductReservation(ctx context.Context, r *http.Request) ([]byte, error) {
	request, err := buildRequest[dto.CancelRequest](r)
	if err != nil {
		return nil, fmt.Errorf("error build reserve request. %w", err)
	}

	response, err := s.controllers.ReservationController.Cancel(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error cancel product reservation. %w", err)
	}

	return response, nil
}

func (s *Server) releaseProductReservation(ctx context.Context, r *http.Request) ([]byte, error) {
	request, err := buildRequest[dto.ReleaseRequest](r)
	if err != nil {
		return nil, fmt.Errorf("error build reserve request. %w", err)
	}

	response, err := s.controllers.ReservationController.Release(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error release product reservation. %w", err)
	}

	return response, nil
}
