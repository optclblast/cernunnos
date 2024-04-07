package controllers

import (
	"cernunnos/internal/pkg/dto"
	"cernunnos/internal/server/interface/presenters"
	"cernunnos/internal/usecase/interactors"
	"context"
	"fmt"
	"log/slog"
)

type ReservationController interface {
	// List product reservations
	Reservations(ctx context.Context, req *dto.ReservationsRequest) ([]byte, error)
	// Reserves a product. If StorageId is passed, then reservation will be performed in a storage specified WITHOUT
	// reservation distributing
	Reserve(ctx context.Context, req *dto.ReserveRequest) error
	// Cancels product reservation. If StorageId is passed, then cancellation will be performed in a storage specified only.
	// Reserved products will be available for reservation again.
	Cancel(ctx context.Context, req *dto.CancelRequest) error
	// Releases the reservation. If StorageId is passed, then reservation relese will be performed in a storage specified only.
	// Reserved products will be written off from stock.
	Release(ctx context.Context, req *dto.ReleaseRequest) error
}

func NewReservationController(
	log *slog.Logger,
	interactor interactors.ReservationInteractor,
	presenter presenters.ReservationPresenter,
) ReservationController {
	return &reservationController{
		log:        log.WithGroup("reservation_controller"),
		interactor: interactor,
		presenter:  presenter,
	}
}

type reservationController struct {
	log        *slog.Logger
	interactor interactors.ReservationInteractor
	presenter  presenters.ReservationPresenter
}

func (c *reservationController) Reservations(
	ctx context.Context,
	req *dto.ReservationsRequest,
) ([]byte, error) {
	reservations, err := c.interactor.Reservations(ctx, interactors.ReservationsParams{
		StorageId:  req.StorageId,
		ProductId:  req.ProductId,
		ShippingId: req.ShippingId,
		Limit:      uint64(req.Limit),
		Offset:     uint64(req.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("error fetch reservations. %w", err)
	}

	response, err := c.presenter.ResponseReservations(reservations)
	if err != nil {
		return nil, fmt.Errorf("error build reservations response. %w", err)
	}

	return response, nil
}

func (c *reservationController) Reserve(ctx context.Context, req *dto.ReserveRequest) error {
	err := c.interactor.Reserve(ctx, interactors.ReserveParams{
		ProductIds: req.Products,
		StorageId:  req.StorageId,
		ShippingId: req.ShippingId,
		Amount:     req.Amount,
	})
	if err != nil {
		return fmt.Errorf("error reserve product for shipping. %w", err)
	}

	return nil
}

func (c *reservationController) Cancel(ctx context.Context, req *dto.CancelRequest) error {
	err := c.interactor.Cancel(ctx, interactors.CancelParams{
		ProductIds: req.Products,
		ShippingId: req.ShippingId,
		StorageId:  req.StorageId,
	})
	if err != nil {
		return fmt.Errorf("error cancel product reservation. %w", err)
	}

	return nil
}

func (c *reservationController) Release(ctx context.Context, req *dto.ReleaseRequest) error {
	err := c.interactor.Release(ctx, interactors.ReleaseParams{
		ProductIds: req.Products,
		ShippingId: req.ShippingId,
		StorageId:  req.StorageId,
	})
	if err != nil {
		return fmt.Errorf("error release product reservation. %w", err)
	}

	return nil
}
