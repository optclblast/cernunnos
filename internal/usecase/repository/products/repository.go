package products

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
	// List products info
	Products(ctx context.Context, params ProductsParams) ([]*models.ProductInfo, error)
	// List products in a spicific storage
	StorageProducts(ctx context.Context, params StorageProductsParams) ([]*models.StorageProduct, error)
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

type ProductsParams struct {
	Ids    uuid.UUIDs
	Limit  uint32
	Offset uint32
}

// List products info
func (r *repositorySql) Products(ctx context.Context, params ProductsParams) ([]*models.ProductInfo, error) {
	batchSize := sqltools.DefaultLimit

	if len(params.Ids) > 0 {
		batchSize = uint32(len(params.Ids))
	}

	products := make([]*models.ProductInfo, 0, batchSize)

	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		var err error

		rows, err := buildSelectProductsQuery(params).RunWith(r.Conn(ctx)).QueryContext(ctx)
		if err != nil {
			return fmt.Errorf("error fetch products from database. %w", err)
		}

		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				err = errors.Join(fmt.Errorf("error close rows. %w", closeErr), err)
			}
		}()

		for rows.Next() {
			var (
				id        uuid.UUID
				name      string
				size      int64
				createdAt time.Time
				updatedAt time.Time
			)

			if err = rows.Scan(&id, &name, &size, &createdAt, &updatedAt); err != nil {
				return fmt.Errorf("error scan row. %w", err)
			}

			products = append(products, &models.ProductInfo{
				Id:        id,
				Name:      name,
				Size:      size,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			})
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error process rows. %w", err)
		}

		return err
	})

	if err != nil {
		return nil, fmt.Errorf("error execute transactional operation. %w", err)
	}

	return products, nil
}

func buildSelectProductsQuery(params ProductsParams) sq.SelectBuilder {
	selectQuery := sq.Select("id", "name", "size", "created_at", "updated_at").
		From("products").PlaceholderFormat(sq.Dollar)

	if len(params.Ids) > 0 {
		selectQuery = selectQuery.Where(sq.Eq{
			"id": params.Ids,
		})
	}

	if params.Limit > 0 && params.Limit < sqltools.DefaultLimit {
		selectQuery = selectQuery.Limit(uint64(params.Limit))
	} else {
		selectQuery = selectQuery.Limit(uint64(sqltools.DefaultLimit))
	}

	if params.Offset > 0 {
		selectQuery = selectQuery.Offset(uint64(params.Offset))
	}

	return selectQuery
}

type StorageProductsParams struct {
	Ids             uuid.UUIDs
	StorageId       uuid.UUID
	WithUnavailable bool
	Limit           uint64
	Offset          uint64
}

// List products in a spicific storage
func (r *repositorySql) StorageProducts(
	ctx context.Context,
	params StorageProductsParams,
) ([]*models.StorageProduct, error) {
	storageProducts := make([]*models.StorageProduct, 0, len(params.Ids))

	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		var err error
		if len(params.Ids) > 0 {
			storageProducts, err = r.storageProductsByIds(ctx, params)
			if err != nil {
				return fmt.Errorf("error fetch storage profucts by ids. %w", err)
			}

			return nil
		}

		storageProducts, err = r.allStorageProducts(ctx, params)
		if err != nil {
			return fmt.Errorf("error fetch storage profucts. %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error execute transactional operation. %w", err)
	}

	return storageProducts, nil
}

func (r *repositorySql) allStorageProducts(
	ctx context.Context,
	params StorageProductsParams,
) ([]*models.StorageProduct, error) {
	storageProducts := make([]*models.StorageProduct, 0, len(params.Ids))

	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		query := buildSelectStorageProductQuery(buildSelectStorageProductQueryParams{
			storageId:       params.StorageId,
			withUnavailable: params.WithUnavailable,
			limit:           params.Limit,
			offset:          params.Offset,
		})

		var err error

		storageProducts, err = r.fetchStorageProductsWithQuery(ctx, query)
		if err != nil {
			return fmt.Errorf("error fetch products. %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error execute transactional operation. %w", err)
	}

	return storageProducts, nil
}

func (r *repositorySql) storageProductsByIds(
	ctx context.Context,
	params StorageProductsParams,
) ([]*models.StorageProduct, error) {
	storageProducts := make([]*models.StorageProduct, 0, len(params.Ids))

	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		for _, productId := range params.Ids {
			query := buildSelectStorageProductQuery(buildSelectStorageProductQueryParams{
				productId:       productId,
				storageId:       params.StorageId,
				withUnavailable: params.WithUnavailable,
				limit:           params.Limit,
				offset:          params.Offset,
			})

			var err error

			products, err := r.fetchStorageProductsWithQuery(ctx, query)
			if err != nil {
				return fmt.Errorf("error fetch products. %w", err)
			}

			storageProducts = append(storageProducts, products...)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error execute transactional operation. %w", err)
	}

	return storageProducts, nil
}

func (r *repositorySql) fetchStorageProductsWithQuery(
	ctx context.Context,
	query sq.SelectBuilder,
) ([]*models.StorageProduct, error) {
	storageProducts := make([]*models.StorageProduct, 0)
	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		rows, err := query.RunWith(r.Conn(ctx)).QueryContext(ctx)
		if err != nil {
			return fmt.Errorf("error fetch data from database. %w", err)
		}

		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				err = errors.Join(fmt.Errorf("error close rows. %w", closeErr), err)
			}
		}()

		for rows.Next() {
			var (
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

				productsDistributionReserved  int64
				productsDistributionAmount    int64
				productsDistributionAvailable int64
			)

			if err = rows.Scan(
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
				&productsDistributionAmount,
				&productsDistributionReserved,
				&productsDistributionAvailable,
			); err != nil {
				return fmt.Errorf("error scan row. %w", err)
			}

			storageProducts = append(storageProducts, &models.StorageProduct{
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
					Amount:    productsDistributionAmount,
					Reserved:  productsDistributionReserved,
					Available: productsDistributionAvailable,
				},
			})
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error process rows. %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error execute transactional operation. %w", err)
	}

	return storageProducts, nil
}

type buildSelectStorageProductQueryParams struct {
	productId       uuid.UUID
	storageId       uuid.UUID
	withUnavailable bool
	limit           uint64
	offset          uint64
}

func buildSelectStorageProductQuery(params buildSelectStorageProductQueryParams) sq.SelectBuilder {
	query := sq.Select(
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
		"pd.available",
	).
		From("products as p").
		InnerJoin(
			"products_distribution as pd on pd.product_id = p.id",
		).
		InnerJoin("storages as s on pd.storage_id = s.id").
		PlaceholderFormat(sq.Dollar)

	if !params.withUnavailable {
		query = query.Where(sq.Gt{
			"pd.available": 0,
		})
	}

	if params.productId != uuid.Nil {
		query = query.Where(sq.Eq{
			"p.id": params.productId,
		})
	}

	if params.storageId != uuid.Nil {
		query = query.Where(sq.Eq{
			"s.id": params.storageId,
		})
	}

	if params.limit > 0 && params.limit < 500 {
		query = query.Limit(params.limit)
	} else {
		query = query.Limit(uint64(sqltools.DefaultLimit))
	}

	if params.offset > 0 {
		query = query.Offset(params.offset)
	}

	return query
}
