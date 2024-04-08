package models

import "fmt"

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
				Storage:  storageProduct.Storage,
				Amount:   storageProduct.Amount,
				Reserved: storageProduct.Reserved,
			},
		},
	}, nil
}
