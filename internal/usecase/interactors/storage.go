package interactors

import (
	"cernunnos/internal/pkg/dto"
	"cernunnos/internal/pkg/models"
	productsRepo "cernunnos/internal/usecase/repository/products"
	storagesRepo "cernunnos/internal/usecase/repository/storages"
	"context"
	"fmt"
	"log/slog"
)

type StorageInteractor interface {
	Storages(ctx context.Context, params StoragesParams) ([]*models.Storage, error)
}

type storageInteractor struct {
	log                *slog.Logger
	storagesRepository storagesRepo.Repository
	productsRepository productsRepo.Repository
}

func NewStorageInteractor(
	log *slog.Logger,
	storagesRepository storagesRepo.Repository,
	productsRepository productsRepo.Repository,
) StorageInteractor {
	return &storageInteractor{
		log:                log.WithGroup("storage_interactor"),
		storagesRepository: storagesRepository,
		productsRepository: productsRepository,
	}
}

type StoragesParams struct {
	Ids             []string
	WithBusy        bool
	WithUnavailable bool
}

func (c *storageInteractor) Storages(ctx context.Context, params StoragesParams) ([]*models.Storage, error) {
	uuids, err := dto.MapIdsToUUIDs(params.Ids)
	if err != nil {
		return nil, fmt.Errorf("error map storage ids to uuids. %w", err)
	}

	storages, err := c.storagesRepository.Storages(ctx, storagesRepo.StoragesParams{
		Ids:             uuids,
		WithBusy:        params.WithBusy,
		WithUnavailable: params.WithUnavailable,
	})
	if err != nil {
		return nil, fmt.Errorf("error fetch storages from database. %w", err)
	}

	return storages, nil
}
