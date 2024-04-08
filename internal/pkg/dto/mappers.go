package dto

import (
	"cernunnos/internal/pkg/models"
	"fmt"

	"github.com/google/uuid"
)

func MapStoragesFromModels(models []*models.Storage) ([]*Storage, error) {
	mapped := make([]*Storage, len(models))

	for i, model := range models {
		if models == nil {
			continue
		}

		dto, err := MapStorageFromModel(model)
		if err != nil {
			return nil, fmt.Errorf("error map storage to dto. %w", err)
		}

		mapped[i] = dto
	}

	return mapped, nil
}

func MapStorageFromModel(model *models.Storage) (*Storage, error) {
	if model == nil {
		return nil, fmt.Errorf("error nil storage model")
	}

	availability, err := mapStorageAvailability(model.Availability)
	if err != nil {
		return nil, fmt.Errorf("error map storage availability to dto. %w", err)
	}

	return &Storage{
		Id:           model.Id.String(),
		Name:         model.Name,
		Availability: availability,
		CreatedAt:    uint64(model.CreatedAt.UnixMilli()),
		UpdatedAt:    uint64(model.UpdatedAt.UnixMilli()),
	}, nil
}

func MapIdsToUUIDs(ids []string) (uuid.UUIDs, error) {
	uuids := make(uuid.UUIDs, len(ids))

	for i, id := range ids {
		uuid, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("error parse uuid. %w", err)
		}

		uuids[i] = uuid
	}

	return uuids, nil
}

func MapStorageProductsFromModels(storageProducts []*models.StorageProduct) ([]*StorageProduct, error) {
	mapped := make([]*StorageProduct, len(storageProducts))

	for i, storageProduct := range storageProducts {
		mappedProduct, err := MapStorageProductFromModel(storageProduct)
		if err != nil {
			return nil, fmt.Errorf("error map product info to dto. %w", err)
		}

		mapped[i] = mappedProduct
	}

	return mapped, nil
}

func MapStorageProductFromModel(storageProduct *models.StorageProduct) (*StorageProduct, error) {
	if storageProduct == nil {
		return nil, fmt.Errorf("error nil storage product model")
	}

	return &StorageProduct{
		ProductInfo: ProductInfo{
			Id:        storageProduct.Id.String(),
			Name:      storageProduct.Name,
			Size:      storageProduct.Size,
			CreatedAt: uint64(storageProduct.CreatedAt.UnixMilli()),
			UpdatedAt: uint64(storageProduct.UpdatedAt.UnixMilli()),
		},
		ProductDestribution: ProductDestribution{
			StorageId: storageProduct.Storage.Id.String(),
			Amount:    storageProduct.Amount,
			Reserved:  storageProduct.Reserved,
			Available: storageProduct.Available,
		},
	}, nil
}

func MapProductsInfoFromModels(products []*models.ProductInfo) ([]*ProductInfo, error) {
	mapped := make([]*ProductInfo, len(products))

	for i, product := range products {
		mappedProduct, err := MapProductInfoFromModel(product)
		if err != nil {
			return nil, fmt.Errorf("error map product info to dto. %w", err)
		}

		mapped[i] = mappedProduct
	}

	return mapped, nil
}

func MapProductInfoFromModel(product *models.ProductInfo) (*ProductInfo, error) {
	if product == nil {
		return nil, fmt.Errorf("error nil product info model")
	}

	destributions, err := MapProductDestributionFromModel(product.DestributionInfo)
	if err != nil {
		return nil, fmt.Errorf("error map product destribution info to dto. %w", err)
	}

	return &ProductInfo{
		Id:               product.Id.String(),
		Name:             product.Name,
		Size:             product.Size,
		CreatedAt:        uint64(product.CreatedAt.UnixMilli()),
		UpdatedAt:        uint64(product.UpdatedAt.UnixMilli()),
		DestributionInfo: destributions,
	}, nil
}

func MapProductDestributionFromModel(destridutions []*models.ProductDestribution) ([]*ProductDestribution, error) {
	mapped := make([]*ProductDestribution, len(destridutions))

	for i, destribution := range destridutions {
		mapped[i] = &ProductDestribution{
			StorageId: destribution.Storage.Id.String(),
			Amount:    destribution.Amount,
			Reserved:  destribution.Reserved,
		}
	}

	return mapped, nil
}

func MapStorageProductsToProductInfos(storageProducts []*StorageProduct) ([]*ProductInfo, error) {
	infos := make([]*ProductInfo, len(storageProducts))

	for i, storageProduct := range storageProducts {
		info, err := MapStorageProductToProductInfo(storageProduct)
		if err != nil {
			return nil, fmt.Errorf("error map storage product to product info. %w", err)
		}

		infos[i] = info
	}

	return infos, nil
}

func MapStorageProductToProductInfo(storageProduct *StorageProduct) (*ProductInfo, error) {
	if storageProduct == nil {
		return nil, fmt.Errorf("error nil storage product")
	}

	return &ProductInfo{
		Id:        storageProduct.Id,
		Name:      storageProduct.Name,
		Size:      storageProduct.Size,
		CreatedAt: storageProduct.CreatedAt,
		UpdatedAt: storageProduct.UpdatedAt,
		DestributionInfo: []*ProductDestribution{
			{
				StorageId: storageProduct.StorageId,
				Amount:    storageProduct.Amount,
				Reserved:  storageProduct.Reserved,
			},
		},
	}, nil
}

func MapReservationsFromModels(models []*models.Reservation) ([]*Reservation, error) {
	reservations := make([]*Reservation, len(models))

	for i, model := range models {
		reservation, err := MapReservationFromModel(model)
		if err != nil {
			return nil, fmt.Errorf("error map reservation to dto. %w", err)
		}

		reservations[i] = reservation
	}

	return reservations, nil
}

func MapReservationFromModel(model *models.Reservation) (*Reservation, error) {
	if model == nil {
		return nil, fmt.Errorf("error nil reservation model")
	}

	return &Reservation{
		StorageId:  model.StorageId.String(),
		ProductId:  model.ReservedProduct.Id.String(),
		ShippingId: model.ShippingId.String(),
		Reserved:   model.Reserved,
		CreatedAt:  uint64(model.CreatedAt.UnixMilli()),
		UpdatedAt:  uint64(model.UpdatedAt.UnixMilli()),
	}, nil
}

func mapStorageAvailability(availability models.StorageAvailability) (string, error) {
	switch availability {
	case models.StorageAvailabilityAvailable:
		return "available", nil
	case models.StorageAvailabilityUnavailable:
		return "unavailable", nil
	case models.StorageAvailabilityBusy:
		return "busy", nil
	default:
		return "", fmt.Errorf("error unexpected availability status")
	}
}
