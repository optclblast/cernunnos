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
	Limit      uint64
	Offset     uint64
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
		InnerJoin("products as p on p.id = r.product_id").
		InnerJoin("storages as s on s.id = r.storage_id").
		InnerJoin(
			"products_distribution as pd on pd.storage_id = r.storage_id AND pd.product_id = r.product_id",
		).PlaceholderFormat(sq.Dollar)

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

	if params.Limit > 0 && params.Limit < 500 {
		query = query.Limit(params.Limit)
	} else {
		query = query.Limit(uint64(sqltools.DefaultLimit))
	}

	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	fmt.Println(query.ToSql())

	return query
}

type ReserveParams struct {
	ProductIds uuid.UUIDs
	StorageId  uuid.UUID
	ShippingId uuid.UUID
	Amount     int64
}

func (r *repositorySql) Reserve(ctx context.Context, params ReserveParams) error {
	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		for _, productId := range params.ProductIds {
			storagesToReserve, err := r.storagesToReserveIn(ctx, storagesToReserveInParams{
				productId: productId,
				storageId: params.StorageId,
				amount:    params.Amount,
			})
			if err != nil {
				return fmt.Errorf("error fetch storages to reserve product in. %w", err)
			}

			for storageId, availableSlots := range storagesToReserve {
				fmt.Println(storageId.String())
				err = r.reserve(ctx, reserveParams{
					productId:  productId,
					storageId:  storageId,
					shippingId: params.ShippingId,
					amount:     availableSlots,
				})
				if err != nil {
					return fmt.Errorf("error reserve slots in %s. %w", storageId.String(), err)
				}
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
				"reserved":  sq.Expr("reserved + ?", params.amount),
				"available": sq.Expr("available - ?", params.amount),
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

		insertQuery := sq.Insert("products_reservations").Columns(
			"storage_id",
			"product_id",
			"shipping_id",
			"reserved",
			"created_at",
			"updated_at",
		).Values(
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

type storagesToReserveInParams struct {
	productId uuid.UUID
	storageId uuid.UUID
	amount    int64
}

func (r *repositorySql) storagesToReserveIn(
	ctx context.Context,
	params storagesToReserveInParams,
) (map[uuid.UUID]int64, error) {
	uuids := make(map[uuid.UUID]int64)
	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		isSpaceAvailableQuery := sq.Select(
			"storage_id",
			"available",
		).
			From("products_distribution").
			Where(sq.And{
				sq.Eq{
					"product_id": params.productId,
				},
				sq.Gt{
					"available": 0,
				},
			},
			).OrderBy("available DESC").
			Suffix("for update").
			PlaceholderFormat(sq.Dollar)

		if params.storageId != uuid.Nil {
			fmt.Println(params.storageId.String())
			isSpaceAvailableQuery = isSpaceAvailableQuery.Where(sq.Eq{
				"storage_id": params.storageId,
			})
		}

		rows, err := isSpaceAvailableQuery.RunWith(r.Conn(ctx)).QueryContext(ctx)
		if err != nil {
			return fmt.Errorf("error fetch available storages from database. %w", err)
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				err = errors.Join(fmt.Errorf("error close rows. %w", closeErr), err)
			}
		}()

		fmt.Println(isSpaceAvailableQuery.ToSql())
		fmt.Println(params.productId.String())

		var left int64 = params.amount

		for rows.Next() {
			var storageId uuid.UUID = uuid.Nil
			var available int64

			if err := rows.Scan(&storageId, &available); err != nil {
				return fmt.Errorf("error scan row. %w", err)
			}
			fmt.Println(storageId.String())

			if available >= left {
				uuids[storageId] = left
				return err
			} else {
				left = left - available
				uuids[storageId] = available
			}

		}

		if left > 0 {
			return fmt.Errorf("error all storages are busy. %w", ErrorNotEnoughSpace)
		}

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("error execure transactional operation. %w", err)
	}

	return uuids, nil
}

type CancelParams struct {
	ProductIds uuid.UUIDs
	StorageId  uuid.UUID
	ShippingId uuid.UUID
}

func (r *repositorySql) Cancel(ctx context.Context, params CancelParams) error {
	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		for _, productId := range params.ProductIds {
			reservations, err := r.reservedByStorage(ctx, reservedByStorageParams{
				storageId:  params.StorageId,
				shippingId: params.ShippingId,
				productId:  productId,
			})
			if err != nil {
				return fmt.Errorf("error fetch reservations in storages. %w", err)
			}

			for storage, reserved := range reservations {
				err = r.freeReservation(ctx, cancelReservationParams{
					storageId:  storage,
					shippingId: params.ShippingId,
					productId:  productId,
					amount:     reserved,
				})
				if err != nil {
					return fmt.Errorf(
						"error cancel product reservation at %s. %w", storage.String(), err,
					)
				}
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error execure transactional operation. %w", err)
	}

	return nil
}

type ReleaseParams struct {
	ProductIds uuid.UUIDs
	StorageId  uuid.UUID
	ShippingId uuid.UUID
}

func (r *repositorySql) Release(ctx context.Context, params ReleaseParams) error {
	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		for _, productId := range params.ProductIds {
			reservations, err := r.reservedByStorage(ctx, reservedByStorageParams{
				storageId:  params.StorageId,
				shippingId: params.ShippingId,
				productId:  productId,
			})
			if err != nil {
				return fmt.Errorf("error fetch reservations in storages. %w", err)
			}

			for storage, reserved := range reservations {
				err = r.freeReservation(ctx, cancelReservationParams{
					storageId:  storage,
					shippingId: params.ShippingId,
					productId:  productId,
					amount:     reserved,
					writeOff:   true,
				})
				if err != nil {
					return fmt.Errorf(
						"error release product reservation at %s. %w", storage.String(), err,
					)
				}
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error execure transactional operation. %w", err)
	}

	return nil
}

type reservedByStorageParams struct {
	productId  uuid.UUID
	storageId  uuid.UUID
	shippingId uuid.UUID
}

func (r *repositorySql) reservedByStorage(
	ctx context.Context,
	params reservedByStorageParams,
) (map[uuid.UUID]int64, error) {
	reservations := make(map[uuid.UUID]int64)
	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		selectReservations := sq.Select(
			"pr.storage_id",
			"pr.reserved",
		).
			From("products_reservations as pr").
			Where(sq.Eq{
				"product_id":  params.productId,
				"shipping_id": params.shippingId,
			}).PlaceholderFormat(sq.Dollar)

		if params.storageId != uuid.Nil {
			selectReservations = selectReservations.Where(sq.Eq{
				"storage_id": params.storageId,
			})
		}

		fmt.Println(selectReservations.ToSql())

		rows, err := selectReservations.RunWith(r.Conn(ctx)).QueryContext(ctx)
		if err != nil {
			return fmt.Errorf("error fetch reservations from database. %w", err)
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				err = errors.Join(fmt.Errorf("error close rows. %w", closeErr), err)
			}
		}()

		for rows.Next() {
			var storageId uuid.UUID
			var reserved int64

			if err = rows.Scan(&storageId, &reserved); err != nil {
				return fmt.Errorf("error scan row. %w", err)
			}

			reservations[storageId] = reserved
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error execure transactional operation. %w", err)
	}

	return reservations, nil
}

type cancelReservationParams struct {
	productId  uuid.UUID
	storageId  uuid.UUID
	shippingId uuid.UUID
	amount     int64
	writeOff   bool
}

func (r *repositorySql) freeReservation(ctx context.Context, params cancelReservationParams) error {
	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		query := sq.Update("products_distribution").SetMap(sq.Eq{
			"reserved": sq.Expr("reserved - ?", params.amount),
		}).Where(sq.Eq{
			"product_id": params.productId,
			"storage_id": params.storageId,
		}).PlaceholderFormat(sq.Dollar)

		if !params.writeOff {
			query = query.Set("available", sq.Expr("available + ?", params.amount))
		} else {
			query = query.Set("amount", sq.Expr("amount - ?", params.amount))
		}

		if _, err := query.RunWith(r.Conn(ctx)).ExecContext(ctx); err != nil {
			return fmt.Errorf("error update amount of an available items. %w", err)
		}

		delete := sq.Delete("products_reservations").
			Where(sq.Eq{
				"product_id":  params.productId,
				"shipping_id": params.shippingId,
				"storage_id":  params.storageId,
			}).PlaceholderFormat(sq.Dollar)
		if _, err := delete.RunWith(r.Conn(ctx)).ExecContext(ctx); err != nil {
			return fmt.Errorf("error delete reservation. %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error execure transactional operation. %w", err)
	}

	return nil
}
