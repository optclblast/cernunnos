package controllers

import (
	"cernunnos/internal/pkg/dto"
	"cernunnos/internal/server/interface/presenters"
	"cernunnos/internal/usecase/interactors"
	"context"
	"fmt"
	"log/slog"
)

type StorageController interface {
	Storages(ctx context.Context, req *dto.StoragesRequest) ([]byte, error)
}

type storageController struct {
	log        *slog.Logger
	interactor interactors.StorageInteractor
	presenter  presenters.StoragePresenter
}

func NewStorageController(
	log *slog.Logger,
	interactor interactors.StorageInteractor,
	presenter presenters.StoragePresenter,
) StorageController {
	return &storageController{
		log:        log.WithGroup("storage_controller"),
		interactor: interactor,
		presenter:  presenter,
	}
}

func (c *storageController) Storages(ctx context.Context, req *dto.StoragesRequest) ([]byte, error) {
	storages, err := c.interactor.Storages(ctx, interactors.StoragesParams{
		Ids:             req.Ids,
		WithBusy:        req.WithBusy,
		WithUnavailable: req.WithUnavailable,
	})
	if err != nil {
		return nil, fmt.Errorf("error fetch storages. %w", err)
	}

	response, err := c.presenter.ResponseStorages(storages)
	if err != nil {
		return nil, fmt.Errorf("error build storages response. %w", err)
	}

	return response, nil
}
