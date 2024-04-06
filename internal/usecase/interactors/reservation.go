package interactors

import (
	"cernunnos/internal/pkg/models"
	reservationsRepo "cernunnos/internal/usecase/repository/reservations"
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type ReservationInteractor interface {
	// List product reservations
	Reservations(ctx context.Context, params ReservationsParams) ([]*models.Reservation, error)
	// Reserves a product. If StorageId is passed, then reservation will be performed in a storage specified WITHOUT
	// reservation distributing
	Reserve(ctx context.Context, params ReserveParams) error
	// Cancels product reservation. If StorageId is passed, then cancellation will be performed in a storage specified only.
	// Reserved products will be available for reservation again.
	Cancel(ctx context.Context, params CancelParams) error
	// Releases the reservation. If StorageId is passed, then reservation relese will be performed in a storage specified only.
	// Reserved products will be written off from stock.
	Release(ctx context.Context, params ReleaseParams) error
}

func NewReservationInteractor(
	log *slog.Logger,
	reservationsRepository reservationsRepo.Repository,
) ReservationInteractor {
	return &reservationInteractor{
		log:                    log.WithGroup("reservation_interactor"),
		reservationsRepository: reservationsRepository,
	}
}

type reservationInteractor struct {
	log                    *slog.Logger
	reservationsRepository reservationsRepo.Repository
}

type ReservationsParams struct {
	StorageId  string // If StorageId is passed, then reservation will be performed in a storage specified WITHOUT
	ProductId  string
	ShippingId string
}

func (c *reservationInteractor) Reservations(
	ctx context.Context,
	params ReservationsParams,
) ([]*models.Reservation, error) {
	ids, err := processIds(params.ProductId, params.StorageId, params.ShippingId)
	if err != nil {
		return nil, fmt.Errorf("error parse ids. %w", err)
	}

	reservations, err := c.reservationsRepository.Reservations(ctx, reservationsRepo.ReservationsParams{
		StorageId:  ids.storageId,
		ProductId:  ids.productId,
		ShippingId: ids.shippingId,
	})
	if err != nil {
		return nil, fmt.Errorf("error fetch reservations from repository. %w", err)
	}

	return reservations, nil
}

type ReserveParams struct {
	ProductId  string
	StorageId  string
	ShippingId string
	Amount     int64
}

func (c *reservationInteractor) Reserve(ctx context.Context, params ReserveParams) error {
	ids, err := processIds(params.ProductId, params.StorageId, params.ShippingId)
	if err != nil {
		return fmt.Errorf("error parse ids. %w", err)
	}

	err = c.reservationsRepository.Reserve(ctx, reservationsRepo.ReserveParams{
		ProductId:  ids.productId,
		StorageId:  ids.storageId,
		ShippingId: ids.shippingId,
		Amount:     params.Amount,
	})
	if err != nil {
		return fmt.Errorf("error reserve product for shipping. %w", err)
	}

	return nil
}

type CancelParams struct {
	ProductId  string
	StorageId  string // If StorageId is passed, then cancellation will be performed in a storage specified only.
	ShippingId string
}

func (c *reservationInteractor) Cancel(ctx context.Context, params CancelParams) error {
	ids, err := processIds(params.ProductId, params.StorageId, params.ShippingId)
	if err != nil {
		return fmt.Errorf("error parse ids. %w", err)
	}

	err = c.reservationsRepository.Cancel(ctx, reservationsRepo.CancelParams{
		ProductId:  ids.productId,
		ShippingId: ids.shippingId,
		StorageId:  ids.storageId,
	})
	if err != nil {
		return fmt.Errorf("error cancel product reservation. %w", err)
	}

	return nil
}

type ReleaseParams struct {
	ProductId  string
	StorageId  string // If StorageId is passed, then reservation relese will be performed in a storage specified only.
	ShippingId string
}

func (c *reservationInteractor) Release(ctx context.Context, params ReleaseParams) error {
	ids, err := processIds(params.ProductId, params.StorageId, params.ShippingId)
	if err != nil {
		return fmt.Errorf("error parse ids. %w", err)
	}

	err = c.reservationsRepository.Release(ctx, reservationsRepo.ReleaseParams{
		ProductId:  ids.productId,
		ShippingId: ids.shippingId,
		StorageId:  ids.storageId,
	})
	if err != nil {
		return fmt.Errorf("error release product reservation. %w", err)
	}

	return nil
}

type reservationsIds struct {
	storageId  uuid.UUID
	productId  uuid.UUID
	shippingId uuid.UUID
}

func processIds(
	productIdString string,
	storageIdString string,
	shippingIdString string,
) (*reservationsIds, error) {
	var (
		err        error
		storageId  uuid.UUID
		productId  uuid.UUID
		shippingId uuid.UUID
	)

	if storageIdString != "" {
		storageId, err = uuid.Parse(storageIdString)
		if err != nil {
			return nil, fmt.Errorf("error parse storage id. %w", err)
		}
	}

	if productIdString != "" {
		productId, err = uuid.Parse(productIdString)
		if err != nil {
			return nil, fmt.Errorf("error parse product id. %w", err)
		}
	}

	if shippingIdString != "" {
		shippingId, err = uuid.Parse(shippingIdString)
		if err != nil {
			return nil, fmt.Errorf("error parse shipping id. %w", err)
		}
	}

	return &reservationsIds{
		storageId:  storageId,
		productId:  productId,
		shippingId: shippingId,
	}, nil
}
