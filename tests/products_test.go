package tests

import (
	"cernunnos/internal/pkg/dto"
	"cernunnos/internal/server/interface/controllers"
	"cernunnos/internal/server/interface/presenters"
	"cernunnos/internal/usecase/interactors"
	"cernunnos/internal/usecase/repository"
	"cernunnos/internal/usecase/repository/products"
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

func TestProductFetching(t *testing.T) {
	db, cleanup, err := repository.ProvideDatabaseConnection(&cfg)
	if err != nil {
		t.Fatal("error connect to database", err)
	}

	productsInteractor := interactors.NewProductInteractor(slog.Default(), products.NewRepository(db))

	defer cleanup()

	t.Log("Test: products fetching\n")

	storageId := uuid.New()
	storageName := gofakeit.StreetName()
	storageSpace := int64(math.MaxInt64)

	var reserved int64
	var storageReserved int64 = storageSpace / 2

	amount := storageReserved
	if amount > 1 {
		reserved = rand.Int63n(amount / 100)
	}

	available := amount - reserved

	insertStoragesCtx, cancelCtx := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancelCtx()

	err = insertStorages(insertStoragesCtx, db, insertStoragesParams{
		storageId:   storageId,
		storageName: storageName,
		available:   storageSpace - storageReserved,
		reserved:    storageReserved,
	})
	if err != nil {
		t.Fatal("error add storage", err)
	}

	var cases map[string]Testcase = map[string]Testcase{
		"Normal case": func(t *testing.T) {
			productId := uuid.New()
			productName := gofakeit.ProductName()
			size := rand.Int63n(250)

			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			err = insertProducts(ctx, db, insertProductsParams{
				storageId:   storageId,
				productId:   productId,
				productName: productName,
				size:        size,
				amount:      amount,
				reserved:    reserved,
				available:   available,
			})
			if err != nil {
				t.Fatal("error add product", err)
			}

			products, err := productsInteractor.Products(ctx, interactors.ProductsParams{
				Ids:       []string{productId.String()},
				StorageId: storageId.String(),
				Limit:     1,
			})
			if err != nil {
				t.Fatal("error fetch product", err)
			}

			if len(products) == 0 {
				t.Fatal("error no products fetched", err)
			}

			if productId != products[0].Id {
				t.Fatal("error wrong product", err)
			}
		},
		"Invalid case. Product does not exists": func(t *testing.T) {
			productId := uuid.New()

			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			products, err := productsInteractor.Products(ctx, interactors.ProductsParams{
				Ids:             []string{productId.String()},
				WithUnavailable: true,
				StorageId:       storageId.String(),
				Limit:           1,
			})
			if err != nil {
				t.Fatal("error fetch product", err)
			}

			if len(products) != 0 {
				t.Fatal("error products fetched, but it shouldnt", err)
			}
		},
		"Invalid case. Wrong storage id": func(t *testing.T) {
			productId := uuid.New()
			productName := gofakeit.ProductName()
			size := rand.Int63n(250)

			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			err = insertProducts(ctx, db, insertProductsParams{
				storageId:   storageId,
				productId:   productId,
				productName: productName,
				size:        size,
				amount:      amount,
				reserved:    reserved,
				available:   available,
			})
			if err != nil {
				t.Fatal("error add product", err)
			}

			products, err := productsInteractor.Products(ctx, interactors.ProductsParams{
				Ids:       []string{productId.String()},
				StorageId: uuid.New().String(),
				Limit:     1,
			})
			if err != nil {
				t.Fatal("error fetch product", err)
			}

			if len(products) != 0 {
				t.Fatal("error products fetched, but it shouldnt", err)
			}
		},
	}

	for desc, test := range cases {
		t.Log(desc + "\n")

		test(t)
	}
}

func TestStorageProductRequest(t *testing.T) {
	db, cleanup, err := repository.ProvideDatabaseConnection(&cfg)
	if err != nil {
		t.Fatal("error connect to database", err)
	}

	productsController := controllers.NewProductController(
		slog.Default(),
		presenters.NewProductPresenter(),
		interactors.NewProductInteractor(slog.Default(), products.NewRepository(db)),
	)

	storageId := uuid.New()
	storageSpace := int64(math.MaxInt64)
	storageName := gofakeit.StreetName()

	var reserved int64
	var storageReserved int64 = storageSpace / 2

	amount := storageReserved
	if amount > 1 {
		reserved = rand.Int63n(amount / 100)
	}

	available := amount - reserved

	insertStoragesCtx, cancelCtx := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancelCtx()

	err = insertStorages(insertStoragesCtx, db, insertStoragesParams{
		storageId:   storageId,
		storageName: storageName,
		available:   storageSpace - storageReserved,
		reserved:    storageReserved,
	})
	if err != nil {
		t.Fatal("error add storage", err)
	}

	defer cleanup()

	t.Log("Test: products fetching\n")

	var cases map[string]Testcase = map[string]Testcase{
		"Normal case": func(t *testing.T) {
			productId := uuid.New()
			productName := gofakeit.ProductName()
			size := rand.Int63n(250)

			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			err = insertProducts(ctx, db, insertProductsParams{
				storageId:   storageId,
				productId:   productId,
				productName: productName,
				size:        size,
				amount:      amount,
				reserved:    reserved,
				available:   available,
			})
			if err != nil {
				t.Fatal("error add product", err)
			}

			data, err := productsController.StorageProducts(ctx, &dto.StorageProductsRequest{
				StorageId:   storageId.String(),
				ProductsIds: []string{productId.String()},
				Limit:       1,
			})
			if err != nil {
				t.Fatal("error fetch storage product", err)
			}

			response := new(dto.StorageProductsResponse)

			err = json.Unmarshal(data, response)
			if err != nil {
				t.Fatal("error unmarshal StorageProducts response", err)
			}

			if len(response.Products) == 0 {
				t.Fatal("error no products fetched", err)
			}

			if response.Products[0].Id != productId.String() {
				t.Fatal("error wrong product id")
			}

			if response.Products[0].Amount != amount {
				t.Fatal("error wrong product id")
			}
		},
		"Invalid product id": func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			data, err := productsController.StorageProducts(ctx, &dto.StorageProductsRequest{
				StorageId:   storageId.String(),
				ProductsIds: []string{uuid.NewString()},
				Limit:       1,
			})
			if err != nil {
				t.Fatal("error fetch storage product", err)
			}

			response := new(dto.StorageProductsResponse)

			err = json.Unmarshal(data, response)
			if err != nil {
				t.Fatal("error unmarshal StorageProducts response", err)
			}

			if len(response.Products) > 0 {
				t.Fatal("error invalid response", err)
			}
		},
		"Invalid storage id": func(t *testing.T) {
			productId := uuid.New()
			productName := gofakeit.ProductName()
			size := rand.Int63n(250)

			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			err = insertProducts(ctx, db, insertProductsParams{
				storageId:   storageId,
				productId:   productId,
				productName: productName,
				size:        size,
				amount:      amount,
				reserved:    reserved,
				available:   available,
			})
			if err != nil {
				t.Fatal("error add product", err)
			}

			data, err := productsController.StorageProducts(ctx, &dto.StorageProductsRequest{
				StorageId:   uuid.NewString(),
				ProductsIds: []string{productId.String()},
				Limit:       1,
			})
			if err != nil {
				t.Fatal("error fetch storage product", err)
			}

			response := new(dto.StorageProductsResponse)

			err = json.Unmarshal(data, response)
			if err != nil {
				t.Fatal("error unmarshal StorageProducts response", err)
			}

			if len(response.Products) > 0 {
				t.Fatal("error invalid response", err)
			}
		},
	}

	for desc, test := range cases {
		t.Log(desc + "\n")

		test(t)
	}
}
