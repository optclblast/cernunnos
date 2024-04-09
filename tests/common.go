package tests

import (
	"cernunnos/internal/pkg/config"
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var cfg config.Config = config.Config{
	DatabaseHost:     "cernunnos-db:5432",
	DatabaseUser:     "cernunnos",
	DatabasePassword: "cernunnos",
}

type insertProductsParams struct {
	productId   uuid.UUID
	storageId   uuid.UUID
	productName string
	size        int64
	amount      int64
	available   int64
	reserved    int64
	skipProduct bool
}

func insertProducts(ctx context.Context, db *sql.DB, params insertProductsParams) error {
	if !params.skipProduct {
		productsQuery := squirrel.Insert("products").
			Columns(
				"id", "name", "size",
			).
			Values(
				params.productId,
				params.productName,
				params.size,
			).
			PlaceholderFormat(squirrel.Dollar)

		if _, err := productsQuery.RunWith(db).ExecContext(ctx); err != nil {
			return fmt.Errorf("error insert into products. %w", err)
		}
	}

	storageProductsQuery := squirrel.Insert("products_distribution").
		Columns(
			"storage_id", "product_id", "amount", "reserved", "available",
		).
		Values(
			params.storageId,
			params.productId,
			params.amount,
			params.reserved,
			params.available,
		).
		PlaceholderFormat(squirrel.Dollar)
	if _, err := storageProductsQuery.RunWith(db).ExecContext(ctx); err != nil {
		return fmt.Errorf("error insert into products_distribution. %w", err)
	}

	return nil
}

type insertStoragesParams struct {
	storageId   uuid.UUID
	storageName string
	available   int64
	reserved    int64
}

func insertStorages(ctx context.Context, db *sql.DB, params insertStoragesParams) error {
	productsQuery := squirrel.Insert("storages").
		Columns(
			"id", "name", "available", "reserved",
		).
		Values(
			params.storageId,
			params.storageName,
			params.available,
			params.reserved,
		).
		PlaceholderFormat(squirrel.Dollar)

	if _, err := productsQuery.RunWith(db).ExecContext(ctx); err != nil {
		return fmt.Errorf("error insert into storages. %w", err)
	}

	return nil
}
