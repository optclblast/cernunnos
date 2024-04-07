package repository

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

// Filter
type StoragesParams struct {
	Ids             []uuid.UUID // If passed, only listed storages will be fetched
	WithBusy        bool        // Fetch busy storages
	WithUnavailable bool        // Fetch unavailable storages
	Limit           uint64      // Limit. Max: 500
	Offset          uint64      // Offset
}

// Storages repository
type Repository interface {
	// Fetch storages by filter
	Storages(ctx context.Context, params StoragesParams) ([]*models.Storage, error)
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

func (r *repositorySql) Storages(ctx context.Context, params StoragesParams) ([]*models.Storage, error) {
	storages := make([]*models.Storage, 0, len(params.Ids))

	err := sqltools.Transaction(ctx, r.db, func(ctx context.Context) error {
		var err error

		query := buildStoragesQuery(params)

		rows, err := query.RunWith(r.Conn(ctx)).QueryContext(ctx)
		if err != nil {
			return fmt.Errorf("error fetch rows from database. %w", err)
		}

		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				err = errors.Join(fmt.Errorf("error close rows. %w", closeErr), err)
			}
		}()

		for rows.Next() {
			var (
				id           uuid.UUID
				name         string
				availability models.StorageAvailability
				createdAt    time.Time
				updatedAt    time.Time
			)

			if err := rows.Scan(&id, &name, &availability, &createdAt, &updatedAt); err != nil {
				return fmt.Errorf("error scan rows. %w", err)
			}

			storages = append(storages, &models.Storage{
				Id:           id,
				Name:         name,
				Availability: availability,
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
			})
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error execute transactional operation. %w", err)
	}

	return storages, nil
}

func buildStoragesQuery(params StoragesParams) sq.SelectBuilder {
	selectQuery := sq.Select("id", "name", "availability", "created_at", "updated_at").
		From("storages").
		Limit(uint64(sqltools.DefaultLimit)).
		Suffix("for update").
		PlaceholderFormat(sq.Dollar)

	if params.Limit != 0 && params.Limit < uint64(sqltools.DefaultLimit) {
		selectQuery = selectQuery.Limit(params.Limit)
	}

	if len(params.Ids) > 0 {
		selectQuery = selectQuery.Where(
			sq.Eq{
				"id": params.Ids,
			},
		)
	}

	if !params.WithBusy {
		selectQuery = selectQuery.Where(
			sq.NotEq{
				"availability": models.StorageAvailabilityBusy,
			},
		)
	}

	if !params.WithUnavailable {
		selectQuery = selectQuery.Where(
			sq.NotEq{
				"availability": models.StorageAvailabilityUnavailable,
			},
		)
	}

	return selectQuery
}
