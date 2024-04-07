package presenters

import (
	"cernunnos/internal/pkg/dto"
	"cernunnos/internal/pkg/models"
	"encoding/json"
	"fmt"
)

type StoragePresenter interface {
	ResponseStorages(storages []*models.Storage) ([]byte, error)
}

type storagePresenter struct{}

func NewStoragePresenter() StoragePresenter {
	return new(storagePresenter)
}

func (p *storagePresenter) ResponseStorages(storages []*models.Storage) ([]byte, error) {
	dtoStorages, err := dto.MapStoragesFromModels(storages)
	if err != nil {
		return nil, fmt.Errorf("error map storages to dto. %w", err)

	}

	response := dto.StoragesResponse{
		Storages: dtoStorages,
		Offset:   uint32(len(dtoStorages)),
	}

	rawResponse, err := json.Marshal(&response)
	if err != nil {
		return nil, fmt.Errorf("error marshal response. %w", err)
	}

	return rawResponse, nil
}
