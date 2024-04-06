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

type ProductsParams struct {
	Ids             uuid.UUIDs
	WithDetribution bool
	WithUnavailable bool
	Limit           uint64
	Offset          uint64
}

type StorageProductsParams struct {
	Ids             uuid.UUIDs
	StorageId       uuid.UUID
	WithUnavailable bool
}

type Repository interface {
	// List products info
	Products(ctx context.Context, params ProductsParams) ([]*models.ProductInfo, error)
	// Add a new product info
	Add(ctx context.Context, productInfo *models.ProductInfo) error
	// Update the product info
	Update(ctx context.Context, productInfo *models.ProductInfo) error
	// Delete the product info. Warning! This action will remove this product from every storage!
	Delete(ctx context.Context, id uuid.UUID) error
	// List products in a spicific storage
	StorageProducts(ctx context.Context, params StorageProductsParams) ([]*models.StorageProduct, error)
	// Add a new product in a spicific storage
	AddStorageProduct(ctx context.Context, storageProduct *models.StorageProduct) error
	// Update a product in a spicific storage
	UpdateStorageProduct(ctx context.Context, storageProduct *models.StorageProduct) error
	// Delete product from a spicific storage
	DeleteStorageProduct(ctx context.Context, id uuid.UUID) error
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

// List products info
func (r *repositorySql) Products(ctx context.Context, params ProductsParams) ([]*models.ProductInfo, error) {
	batchSize := sqltools.DefaultLimit
	if len(params.Ids) > 0 {
		batchSize = uint64(len(params.Ids))
	}

	products := make([]*models.ProductInfo, 0, batchSize)
	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		var err error

		selectQuery := sq.Select("id", "name", "size", "created_at", "updated_at").
			From("products").PlaceholderFormat(sq.Dollar)

		if len(params.Ids) > 0 {
			selectQuery = selectQuery.Where(sq.Eq{
				"id": params.Ids,
			})
		}

		if params.Limit > 0 && params.Limit < sqltools.DefaultLimit {
			selectQuery = selectQuery.Limit(params.Limit)
		} else {
			selectQuery = selectQuery.Limit(sqltools.DefaultLimit)
		}

		if params.Offset > 0 {
			selectQuery = selectQuery.Offset(params.Offset)
		}

		rows, err := selectQuery.RunWith(r.Conn(ctx)).QueryContext(ctx)
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

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("error execute transactional operation. %w", err)
	}

	return products, nil
}

// Add a new product info
func (r *repositorySql) Add(ctx context.Context, productInfo *models.ProductInfo) error {
	return nil
}

// Update the product info
func (r *repositorySql) Update(ctx context.Context, productInfo *models.ProductInfo) error {
	return nil
}

// Delete the product info. Warning! This action will remove this product from every storage!
func (r *repositorySql) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// List products in a spicific storage
func (r *repositorySql) StorageProducts(
	ctx context.Context,
	params StorageProductsParams,
) ([]*models.StorageProduct, error) {
	return nil, nil
}

// Add a new product in a spicific storage
func (r *repositorySql) AddStorageProduct(ctx context.Context, storageProduct *models.StorageProduct) error {
	return nil
}

// Update a product in a spicific storage
func (r *repositorySql) UpdateStorageProduct(
	ctx context.Context,
	storageProduct *models.StorageProduct,
) error {
	return nil
}

// Delete product from a spicific storage
func (r *repositorySql) DeleteStorageProduct(ctx context.Context, id uuid.UUID) error {
	return nil
}
