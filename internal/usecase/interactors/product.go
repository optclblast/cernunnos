package interactors

import (
	"cernunnos/internal/pkg/dto"
	"cernunnos/internal/pkg/models"
	productsRepo "cernunnos/internal/usecase/repository/products"
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type ProductInteractor interface {
	Products(ctx context.Context, params ProductsParams) ([]*models.ProductInfo, error)
	StorageProducts(ctx context.Context, params StorageProductsParams) ([]*models.StorageProduct, error)
}

type productInteractor struct {
	log                *slog.Logger
	productsRepository productsRepo.Repository
}

func NewProductInteractor(
	log *slog.Logger,
	productsRepository productsRepo.Repository,
) ProductInteractor {
	return &productInteractor{
		log:                log.WithGroup("product_interactor"),
		productsRepository: productsRepository,
	}
}

type ProductsParams struct {
	Ids             []string
	StorageId       string
	WithUnavailable bool
	Limit           uint32
	Offset          uint32
}

func (c *productInteractor) Products(
	ctx context.Context,
	params ProductsParams,
) ([]*models.ProductInfo, error) {
	var err error

	ids := make(uuid.UUIDs, 0, len(params.Ids))
	if len(params.Ids) > 0 {
		ids, err = dto.MapIdsToUUIDs(params.Ids)
		if err != nil {
			return nil, fmt.Errorf("error map product ids to uuids. %w", err)
		}
	}

	if params.StorageId != "" {
		storageUUID, err := uuid.Parse(params.StorageId)
		if err != nil {
			return nil, fmt.Errorf("error parse storage id. %w", err)
		}

		storageProducts, err := c.productsRepository.StorageProducts(
			ctx,
			productsRepo.StorageProductsParams{
				Ids:             ids,
				StorageId:       storageUUID,
				WithUnavailable: params.WithUnavailable,
				Limit:           uint64(params.Limit),
				Offset:          uint64(params.Offset),
			},
		)
		if err != nil {
			return nil, fmt.Errorf("error fetch storage products. %w", err)
		}

		infos, err := models.MapStorageProductsToProductInfos(storageProducts)
		if err != nil {
			return nil, fmt.Errorf("error map storage products to product infos. %w", err)
		}

		return infos, nil
	}

	products, err := c.productsRepository.Products(ctx, productsRepo.ProductsParams{
		Ids:    ids,
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("error fetch products. %w", err)
	}

	return products, nil
}

type StorageProductsParams struct {
	Ids             []string
	StorageId       string
	WithUnavailable bool
	Limit           uint64
	Offset          uint64
}

func (c *productInteractor) StorageProducts(
	ctx context.Context,
	params StorageProductsParams,
) ([]*models.StorageProduct, error) {
	var err error

	ids := make(uuid.UUIDs, 0, len(params.Ids))
	if len(params.Ids) > 0 {
		ids, err = dto.MapIdsToUUIDs(params.Ids)
		if err != nil {
			return nil, fmt.Errorf("error map product ids to uuids. %w", err)
		}
	}

	storageUUID, err := uuid.Parse(params.StorageId)
	if err != nil {
		return nil, fmt.Errorf("error parse storage id. %w", err)
	}

	storageProducts, err := c.productsRepository.StorageProducts(
		ctx,
		productsRepo.StorageProductsParams{
			Ids:             ids,
			WithUnavailable: params.WithUnavailable,
			StorageId:       storageUUID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error fetch storage products. %w", err)
	}

	return storageProducts, nil
}
