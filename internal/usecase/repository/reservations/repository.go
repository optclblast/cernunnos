package reservations

import (
	"cernunnos/internal/pkg/models"
	"cernunnos/internal/pkg/sqltools"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type Repository interface {
	Reservations(ctx context.Context, params ReservationsParams) ([]*models.Reservation, error)
	Reserve(ctx context.Context, params ReserveParams) error
	// Cancels product reservation. If StorageId is passed, then cancellation will be performed in a storage specified only.
	// Reserved products will be available for reservation again.
	Cancel(ctx context.Context, params CancelParams) error
	// Releases the reservation. If StorageId is passed, then reservation relese will be performed in a storage specified only.
	// Reserved products will be written off from stock.
	Release(ctx context.Context, params ReleaseParams) error
}

func NewRepository(db *sql.DB) Repository {
	return &repositorySql{db}
}

type repositorySql struct {
	db *sql.DB
}

type txCtxKey struct{}

func (s *repositorySql) Conn(ctx context.Context) sqltools.DBTX {
	if tx, ok := ctx.Value(txCtxKey{}).(*sql.Tx); ok {
		return tx
	}

	return s.db
}

type ReservationsParams struct {
	ProductId  uuid.UUID
	StorageId  uuid.UUID
	ShippingId uuid.UUID
}

func (r *repositorySql) Reservations(
	ctx context.Context,
	params ReservationsParams,
) ([]*models.Reservation, error) {
	reservations := make([]*models.Reservation, 0)
	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		var err error

		rows, err := buildSelectReservationsQuery(params).RunWith(r.Conn(ctx)).QueryContext(ctx)
		if err != nil {
			return fmt.Errorf("error fetch reservations data from database. %w", err)
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				err = errors.Join(fmt.Errorf("error close rows. %w", closeErr), err)
			}
		}()

		for rows.Next() {
			var (
				shippingId           uuid.UUID
				reserved             int64
				reservationCreatedAt time.Time
				reservationUpdatedAt time.Time

				storageId           uuid.UUID
				storageName         string
				storageAvailability models.StorageAvailability
				storageCreatedAt    time.Time
				storageUpdatedAt    time.Time

				productId        uuid.UUID
				productName      string
				productSize      int64
				productCreatedAt time.Time
				productUpdatedAt time.Time

				productsDistributionReserved int64
				productsDistributionAmount   int64
			)

			if err = rows.Scan(
				&shippingId,
				&reserved,
				&reservationCreatedAt,
				&reservationUpdatedAt,
				&storageId,
				&storageName,
				&storageAvailability,
				&storageCreatedAt,
				&storageUpdatedAt,
				&productId,
				&productName,
				&productSize,
				&productCreatedAt,
				&productUpdatedAt,
				&productsDistributionReserved,
				&productsDistributionAmount,
			); err != nil {
				return fmt.Errorf("error scan row. %w", err)
			}

			reservations = append(reservations, &models.Reservation{
				StorageId: storageId,
				ReservedProduct: &models.StorageProduct{
					ProductInfo: models.ProductInfo{
						Id:        productId,
						Name:      productName,
						Size:      productSize,
						CreatedAt: productCreatedAt,
						UpdatedAt: productUpdatedAt,
					},
					ProductDestribution: models.ProductDestribution{
						Storage: &models.Storage{
							Id:           storageId,
							Name:         storageName,
							Availability: storageAvailability,
							CreatedAt:    storageCreatedAt,
							UpdatedAt:    storageUpdatedAt,
						},
						Amount:   productsDistributionAmount,
						Reserved: productsDistributionReserved,
					},
				},
				ShippingId: shippingId,
				Reserved:   reserved,
				CreatedAt:  reservationCreatedAt,
				UpdatedAt:  reservationUpdatedAt,
			})
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error execure transactional operation. %w", err)
	}

	return reservations, nil
}

func buildSelectReservationsQuery(params ReservationsParams) sq.SelectBuilder {
	query := sq.Select(
		// reservations
		"r.shipping_id",
		"r.reserved",
		"r.created_at",
		"r.updated_at",
		// storages
		"s.id",
		"s.name",
		"s.availability",
		"s.created_at",
		"s.updated_at",
		// products
		"p.id",
		"p.name",
		"p.size",
		"p.created_at",
		"p.updated_at",
		// products_distribution
		"pd.amount",
		"pd.reserved",
	).
		From("products_reservations as r").
		InnerJoin("products as p", "p.id = r.product_id").
		InnerJoin("storages as s", "s.id = r.storage_id").
		InnerJoin(
			"products_distribution as pd",
			"pd.storage_id = r.storage_id",
			"pd.product_id = r.product_id",
		)

	if params.ProductId != uuid.Nil {
		query = query.Where(sq.Eq{
			"r.product_id": params.ProductId,
		})
	}

	if params.StorageId != uuid.Nil {
		query = query.Where(sq.Eq{
			"r.storage_id": params.StorageId,
		})
	}

	if params.ShippingId != uuid.Nil {
		query = query.Where(sq.Eq{
			"r.shipping_id": params.ShippingId,
		})
	}

	return query
}

type ReserveParams struct {
	ProductId  uuid.UUID
	StorageId  uuid.UUID
	ShippingId uuid.UUID
	Amount     int64
}

func (r *repositorySql) Reserve(ctx context.Context, params ReserveParams) error {
	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		storagesToReserve, err := r.storagesToReserveIn(ctx, params.ProductId, params.Amount)
		if err != nil {
			return fmt.Errorf("error fetch storages to reserve product in. %w", err)
		}

		for storageId, availableSlots := range storagesToReserve {
			err = r.reserve(ctx, reserveParams{
				productId:  params.ProductId,
				storageId:  storageId,
				shippingId: params.ShippingId,
				amount:     availableSlots,
			})
			if err != nil {
				return fmt.Errorf("error reserve slots in %s. %w", storageId.String(), err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error execure transactional operation. %w", err)
	}

	return nil
}

type reserveParams struct {
	productId  uuid.UUID
	storageId  uuid.UUID
	shippingId uuid.UUID
	amount     int64
}

func (r *repositorySql) reserve(ctx context.Context, params reserveParams) error {
	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		updateQuery := sq.Update("products_distribution").
			SetMap(sq.Eq{
				"reserved":  params.amount,
				"available": sq.Expr("available - $1", params.amount),
			}).
			Where(sq.Eq{
				"storage_id": params.storageId,
				"product_id": params.productId,
			}).
			PlaceholderFormat(sq.Dollar)

		if _, err := updateQuery.RunWith(r.Conn(ctx)).ExecContext(ctx); err != nil {
			return fmt.Errorf(
				"error update product %s distribution data at storage %s. %w",
				params.productId.String(),
				params.storageId.String(),
				err,
			)
		}

		now := time.Now()
		id := uuid.New()

		insertQuery := sq.Insert("products_reservations").Columns(
			"id",
			"storage_id",
			"product_id",
			"shipping_id",
			"reserved",
			"created_at",
			"updated_at",
		).Values(
			id,
			params.storageId,
			params.productId,
			params.shippingId,
			params.amount,
			now,
			now,
		).PlaceholderFormat(sq.Dollar)

		if _, err := insertQuery.RunWith(r.Conn(ctx)).ExecContext(ctx); err != nil {
			return fmt.Errorf(
				"error add product %s reservations data for storage %s. %w",
				params.productId.String(),
				params.storageId.String(),
				err,
			)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error execure transactional operation. %w", err)
	}

	return nil
}

func (r *repositorySql) storagesToReserveIn(
	ctx context.Context,
	productId uuid.UUID,
	amount int64,
) (map[uuid.UUID]int64, error) {
	uuids := make(map[uuid.UUID]int64, 1)
	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		isSpaceAvailableQuery := sq.Select(
			"storage_id",
			"available",
		).
			From("products_distribution").
			Where(
				sq.Eq{
					"product_id": productId,
				},
				sq.Gt{
					"available": 0,
				},
			).OrderBy("available DESC").
			Suffix("for updated").
			PlaceholderFormat(sq.Dollar)

		rows, err := isSpaceAvailableQuery.RunWith(r.Conn(ctx)).QueryContext(ctx)
		if err != nil {
			return fmt.Errorf("error fetch available storages from database. %w", err)
		}

		var availableTotal int64

		for rows.Next() {
			var storageId uuid.UUID
			var available int64

			if err := rows.Scan(&storageId, &available); err != nil {
				return fmt.Errorf("error scan row. %w", err)
			}

			availableTotal += available

			if availableTotal >= amount {
				uuids[storageId] = amount
				return nil
			} else {
				uuids[storageId] = amount - availableTotal
			}

		}

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("error execure transactional operation. %w", err)
	}

	return uuids, nil
}

func (r *repositorySql) lockProductsDistribution(
	ctx context.Context,
	storageId uuid.UUID,
	productId uuid.UUID,
) error {
	return sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		lockQuery := sq.Select("*").
			From("products_distribution").
			Where(sq.Eq{
				"storage_id": storageId,
				"product_id": productId,
			}).Suffix("for update").PlaceholderFormat(sq.Dollar)

		if _, err := lockQuery.RunWith(r.Conn(ctx)).ExecContext(ctx); err != nil {
			return fmt.Errorf("error lock products_distribution rows for update. %w", err)
		}

		return nil
	})
}

type CancelParams struct {
	ProductId  uuid.UUID
	StorageId  uuid.UUID // If StorageId is passed, then cancellation will be performed in a storage specified only.
	ShippingId uuid.UUID
}

func (r *repositorySql) Cancel(ctx context.Context, params CancelParams) error {
	return nil
}

type ReleaseParams struct {
	ProductId  uuid.UUID
	StorageId  uuid.UUID // If StorageId is passed, then reservation relese will be performed in a storage specified only.
	ShippingId uuid.UUID
}

func (r *repositorySql) Release(ctx context.Context, params ReleaseParams) error {
	return nil
}
