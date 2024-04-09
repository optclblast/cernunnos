package utils

import (
	"cernunnos/internal/pkg/models"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type FillDatabaseCommand struct {
	db  *sql.DB
	log *slog.Logger
}

func NewFillDatabaseCommand(db *sql.DB, log *slog.Logger) *FillDatabaseCommand {
	return &FillDatabaseCommand{
		db:  db,
		log: log,
	}
}

func (c *FillDatabaseCommand) Run(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	p := make([]*models.ProductInfo, 0, 250)

	g.Go(func() error {
		var err error

		p, err = c.fillProducts(gCtx)
		if err != nil {
			return fmt.Errorf("error fill database with products. %w", err)
		}

		return nil
	})

	s := make([]*models.Storage, 85)

	g.Go(func() error {
		var err error

		s, err = c.fillStorages(gCtx)
		if err != nil {
			return fmt.Errorf("error fill database with storages. %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error fill database with data. %w", err)
	}

	return c.destributeAndReserveProducts(ctx, p, s)
}

func (c *FillDatabaseCommand) fillProducts(ctx context.Context) ([]*models.ProductInfo, error) {
	products := make([]*models.ProductInfo, 0, 250)

	for i := 0; i < 250; i++ {
		product := &models.ProductInfo{
			Id:        uuid.New(),
			Name:      gofakeit.ProductName(),
			Size:      int64(rand.Int31n(250)),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		query := sq.Insert("products").Columns(
			"id", "name", "size", "created_at", "updated_at",
		).Values(
			product.Id,
			product.Name,
			product.Size,
			product.CreatedAt,
			product.UpdatedAt,
		).PlaceholderFormat(sq.Dollar)

		if _, err := query.RunWith(c.db).ExecContext(ctx); err != nil {
			return nil, fmt.Errorf("error insert product into database. %w", err)
		}

		products = append(products, product)
	}

	return products, nil
}

func (c *FillDatabaseCommand) fillStorages(ctx context.Context) ([]*models.Storage, error) {
	storages := make([]*models.Storage, 0, 85)

	for i := 0; i < 85; i++ {
		storage := &models.Storage{
			Id:        uuid.New(),
			Name:      gofakeit.BeerName(),
			Available: rand.Int63n(100000) + 1000,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		storage.Reserved = rand.Int63n(storage.Available - 1)

		query := sq.Insert("storages").Columns(
			"id", "name", "available", "reserved", "created_at", "updated_at",
		).Values(
			storage.Id,
			storage.Name,
			storage.Available,
			storage.Reserved,
			storage.CreatedAt,
			storage.UpdatedAt,
		).PlaceholderFormat(sq.Dollar)

		if _, err := query.RunWith(c.db).ExecContext(ctx); err != nil {
			return nil, fmt.Errorf("error insert product into database. %w", err)
		}

		storages = append(storages, storage)
	}

	return storages, nil
}

func (c *FillDatabaseCommand) destributeAndReserveProducts(
	ctx context.Context,
	products []*models.ProductInfo,
	storages []*models.Storage,
) error {
	storageFreeSpace := make(map[uuid.UUID]int64)

	for _, s := range storages {
		storageFreeSpace[s.Id] = s.Reserved
	}

	for _, p := range products {
		for _, s := range storages {
			if rand.Intn(100) > 50 {
				continue
			}

			amount := rand.Int63n(storageFreeSpace[s.Id] / 3)

			var reserved int64
			if amount > 0 {
				reserved = rand.Int63n(amount)
			}

			storageFreeSpace[s.Id] -= amount

			insert := sq.Insert("products_distribution").
				Columns(
					"storage_id",
					"product_id",
					"amount",
					"reserved",
					"available",
				).Values(
				s.Id,
				p.Id,
				amount,
				reserved,
				amount-reserved,
			).PlaceholderFormat(sq.Dollar)

			if _, err := insert.RunWith(c.db).ExecContext(ctx); err != nil {
				return fmt.Errorf("error insert product distribution into database. %w", err)
			}

			resQuery := sq.Insert("products_reservations").Columns(
				"storage_id",
				"product_id",
				"shipping_id",
				"reserved",
			).Values(
				s.Id,
				p.Id,
				uuid.New(),
				reserved,
			).PlaceholderFormat(sq.Dollar)
			if _, err := resQuery.RunWith(c.db).ExecContext(ctx); err != nil {
				return fmt.Errorf("error insert product reservation into database. %w", err)
			}
		}
	}

	return nil
}
