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
		return err
	})

	s := make([]*models.Storage, 85)
	g.Go(func() error {
		var err error
		s, err = c.fillStorages(gCtx)
		return err
	})

	if err := g.Wait(); err != nil {
		return err
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
			Id:           uuid.New(),
			Name:         gofakeit.BeerName(),
			Availability: randAvailability(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		query := sq.Insert("storages").Columns(
			"id", "name", "availability", "created_at", "updated_at",
		).Values(
			storage.Id,
			storage.Name,
			storage.Availability,
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

func randAvailability() models.StorageAvailability {
	n := rand.Int31n(300)
	if n < 100 {
		return models.StorageAvailabilityAvailable
	}

	if n >= 100 && n < 200 {
		return models.StorageAvailabilityBusy
	}

	return models.StorageAvailabilityUnavailable
}

func (c *FillDatabaseCommand) destributeAndReserveProducts(
	ctx context.Context,
	products []*models.ProductInfo,
	storages []*models.Storage,
) error {
	for _, p := range products {
		for _, s := range storages {
			if rand.Intn(100) > 50 {
				continue
			}

			amount := rand.Int63n(10000)

			var reserved int64
			if amount > 0 {
				reserved = rand.Int63n(amount)
			}

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
