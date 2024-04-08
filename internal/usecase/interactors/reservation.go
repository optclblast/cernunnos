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
	// Reserves a product. If StorageId is passed, then reservation will be performed in a
	// storage specified WITHOUT reservation distributing
	Reserve(ctx context.Context, params ReserveParams) error
	// Cancels product reservation. If StorageId is passed, then cancellation will be performed in a
	// storage specified only. Reserved products will be available for reservation again.
	Cancel(ctx context.Context, params CancelParams) error
	// Releases the reservation. If StorageId is passed, then reservation relese will be performed in
	// a storage specified only. Reserved products will be written off from stock.
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
	Limit      uint64
	Offset     uint64
}

func (c *reservationInteractor) Reservations(
	ctx context.Context,
	params ReservationsParams,
) ([]*models.Reservation, error) {
	var ids *reservationsIds = new(reservationsIds)

	var err error

	if params.ProductId != "" {
		ids, err = processIds([]string{params.ProductId}, params.StorageId, params.ShippingId)
		if err != nil {
			return nil, fmt.Errorf("error parse ids. %w", err)
		}
	} else {
		ids, err = processIds([]string{}, params.StorageId, params.ShippingId)
		if err != nil {
			return nil, fmt.Errorf("error parse ids. %w", err)
		}
	}

	reservationsParams := reservationsRepo.ReservationsParams{
		StorageId:  ids.storageId,
		ShippingId: ids.shippingId,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}

	if len(ids.productIds) > 0 {
		reservationsParams.ProductId = ids.productIds[0]
	}

	reservations, err := c.reservationsRepository.Reservations(ctx, reservationsParams)
	if err != nil {
		return nil, fmt.Errorf("error fetch reservations from repository. %w", err)
	}

	return reservations, nil
}

type ReserveParams struct {
	ProductIds []string
	StorageId  string
	ShippingId string
	Amount     int64
}

func (c *reservationInteractor) Reserve(ctx context.Context, params ReserveParams) error {
	if params.Amount == 0 || len(params.ProductIds) == 0 ||
		params.ShippingId == "" {
		return fmt.Errorf("error some of required fields are not provided. %w", ErrorFieldRequired)
	}

	ids, err := processIds(params.ProductIds, params.StorageId, params.ShippingId)
	if err != nil {
		return fmt.Errorf("error parse ids. %w", err)
	}

	err = c.reservationsRepository.Reserve(ctx, reservationsRepo.ReserveParams{
		ProductIds: ids.productIds,
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
	ProductIds []string
	StorageId  string // If StorageId is passed, then cancellation will be performed in a storage specified only.
	ShippingId string
}

func (c *reservationInteractor) Cancel(ctx context.Context, params CancelParams) error {
	if len(params.ProductIds) == 0 || params.ShippingId == "" {
		return fmt.Errorf("error some of required fields are not provided. %w", ErrorFieldRequired)
	}

	ids, err := processIds(params.ProductIds, params.StorageId, params.ShippingId)
	if err != nil {
		return fmt.Errorf("error parse ids. %w", err)
	}

	err = c.reservationsRepository.Cancel(ctx, reservationsRepo.CancelParams{
		ProductIds: ids.productIds,
		ShippingId: ids.shippingId,
		StorageId:  ids.storageId,
	})
	if err != nil {
		return fmt.Errorf("error cancel product reservation. %w", err)
	}

	return nil
}

type ReleaseParams struct {
	ProductIds []string
	StorageId  string // If StorageId is passed, then reservation relese will be performed in a storage specified only.
	ShippingId string
}

func (c *reservationInteractor) Release(ctx context.Context, params ReleaseParams) error {
	if len(params.ProductIds) == 0 || params.ShippingId == "" {
		return fmt.Errorf("error some of required fields are not provided. %w", ErrorFieldRequired)
	}

	ids, err := processIds(params.ProductIds, params.StorageId, params.ShippingId)
	if err != nil {
		return fmt.Errorf("error parse ids. %w", err)
	}

	err = c.reservationsRepository.Release(ctx, reservationsRepo.ReleaseParams{
		ProductIds: ids.productIds,
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
	productIds uuid.UUIDs
	shippingId uuid.UUID
}

func processIds(
	productIdsString []string,
	storageIdString string,
	shippingIdString string,
) (*reservationsIds, error) {
	var (
		err        error
		storageId  uuid.UUID
		productIds uuid.UUIDs = make(uuid.UUIDs, 0, len(productIdsString))
		shippingId uuid.UUID
	)

	if storageIdString != "" {
		storageId, err = uuid.Parse(storageIdString)
		if err != nil {
			return nil, fmt.Errorf("error parse storage id. %w", err)
		}
	}

	if len(productIdsString) > 0 {
		for _, id := range productIdsString {
			productId, err := uuid.Parse(id)
			if err != nil {
				return nil, fmt.Errorf("error parse product id. %w", err)
			}

			productIds = append(productIds, productId)
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
		productIds: productIds,
		shippingId: shippingId,
	}, nil
}
