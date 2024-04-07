package controllers

import (
	"cernunnos/internal/pkg/dto"
	"cernunnos/internal/server/interface/presenters"
	"cernunnos/internal/usecase/interactors"
	"context"
	"fmt"
	"log/slog"
)

type ProductController interface {
	Products(ctx context.Context, req *dto.ProductsRequest) ([]byte, error)
	StorageProducts(ctx context.Context, req *dto.StorageProductsRequest) ([]byte, error)
}

type productController struct {
	log        *slog.Logger
	presenter  presenters.ProductPresenter
	interactor interactors.ProductInteractor
}

func NewProductController(
	log *slog.Logger,
	presenter presenters.ProductPresenter,
	interactor interactors.ProductInteractor,
) ProductController {
	return &productController{
		log:        log.WithGroup("product_controller"),
		presenter:  presenter,
		interactor: interactor,
	}
}

func (c *productController) Products(
	ctx context.Context,
	req *dto.ProductsRequest,
) ([]byte, error) {
	products, err := c.interactor.Products(ctx, interactors.ProductsParams{
		Ids:              req.Ids,
		StorageId:        req.StorageId,
		WithDestribution: req.WithDistribution,
		WithUnavailable:  req.WithUnavailable,
		Limit:            req.Limit,
		Offset:           req.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("error fetch products. %w", err)
	}

	response, err := c.presenter.ResponseProducts(products)
	if err != nil {
		return nil, fmt.Errorf("error build products response. %w", err)
	}

	return response, nil
}

func (c *productController) StorageProducts(
	ctx context.Context,
	req *dto.StorageProductsRequest,
) ([]byte, error) {
	products, err := c.interactor.StorageProducts(ctx, interactors.StorageProductsParams{
		Ids:             req.ProductsIds,
		StorageId:       req.StorageId,
		WithUnavailable: req.WithUnavailable,
		Limit:           uint64(req.Limit),
		Offset:          uint64(req.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("error fetch products. %w", err)
	}

	response, err := c.presenter.ResponseStorageProducts(products)
	if err != nil {
		return nil, fmt.Errorf("error build products response. %w", err)
	}

	return response, nil
}
