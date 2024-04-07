package presenters

import (
	"cernunnos/internal/pkg/dto"
	"cernunnos/internal/pkg/models"
	"encoding/json"
	"fmt"
)

type ProductPresenter interface {
	ResponseStorageProducts(products []*models.StorageProduct) ([]byte, error)
	ResponseProducts(products []*models.ProductInfo) ([]byte, error)
}

func NewProductPresenter() ProductPresenter {
	return new(productPresenter)
}

type productPresenter struct{}

func (p *productPresenter) ResponseStorageProducts(products []*models.StorageProduct) ([]byte, error) {
	dtoProducts, err := dto.MapStorageProductsFromModels(products)
	if err != nil {
		return nil, fmt.Errorf("error map storage products to dto. %w", err)
	}

	response := &dto.StorageProductsResponse{
		Products: dtoProducts,
		Offset:   uint32(len(dtoProducts)),
	}

	rawResponse, err := json.Marshal(&response)
	if err != nil {
		return nil, fmt.Errorf("error marshal response. %w", err)
	}

	return rawResponse, nil
}

func (p *productPresenter) ResponseProducts(products []*models.ProductInfo) ([]byte, error) {
	pInfos, err := dto.MapProductsInfoFromModels(products)
	if err != nil {
		return nil, fmt.Errorf("error map products info to dto. %w", err)
	}

	response := &dto.ProductsResponse{
		Products: pInfos,
		Offset:   uint32(len(pInfos)),
	}

	rawResponse, err := json.Marshal(&response)
	if err != nil {
		return nil, fmt.Errorf("error marshal response. %w", err)
	}

	return rawResponse, nil
}
