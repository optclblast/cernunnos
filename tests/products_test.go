package tests

import (
	"cernunnos/internal/usecase/interactors"
	"cernunnos/internal/usecase/repository"
	"cernunnos/internal/usecase/repository/products"
	"context"
	"log/slog"
	"math/rand"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

func TestProductsCreation(t *testing.T) {
	db, cleanup, err := repository.ProvideDatabaseConnection(&cfg)
	if err != nil {
		t.Fatal("error connect to database", err)
	}

	defer cleanup()

	t.Log("Test: products creation\n")

	var cases map[string]Testcase = map[string]Testcase{
		"Normal case": func(t *testing.T) {
			productId := uuid.New()
			storageId := uuid.New()
			productName := gofakeit.ProductName()
			storageName := gofakeit.StreetName()
			size := rand.Int63n(250)
			amount := rand.Int63n(10000)

			var reserved int64
			if amount > 0 {
				reserved = rand.Int63n(amount)
			}

			available := amount - reserved

			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			err = insertStorages(ctx, db, insertStoragesParams{
				storageId:    storageId,
				storageName:  storageName,
				availability: randAvailability(),
			})
			if err != nil {
				t.Fatal("error add storage", err)
			}

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
		},
		"Invalid case. Storage does not exists": func(t *testing.T) {
			productId := uuid.New()
			storageId := uuid.New()
			productName := gofakeit.ProductName()
			size := rand.Int63n(250)
			amount := rand.Int63n(10000)

			var reserved int64
			if amount > 0 {
				reserved = rand.Int63n(amount)
			}

			available := amount - reserved

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
			if err == nil {
				t.Fatal("error add product", err)
			}
		},
		"Invalid case. Product does not exists": func(t *testing.T) {
			productId := uuid.New()
			storageId := uuid.New()
			productName := gofakeit.ProductName()
			storageName := gofakeit.StreetName()
			size := rand.Int63n(250)
			amount := rand.Int63n(10000)

			var reserved int64
			if amount > 0 {
				reserved = rand.Int63n(amount)
			}

			available := amount - reserved

			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			err = insertStorages(ctx, db, insertStoragesParams{
				storageId:    storageId,
				storageName:  storageName,
				availability: randAvailability(),
			})
			if err != nil {
				t.Fatal("error add storage", err)
			}

			err = insertProducts(ctx, db, insertProductsParams{
				storageId:   storageId,
				productId:   productId,
				productName: productName,
				size:        size,
				amount:      amount,
				reserved:    reserved,
				available:   available,
				skipProduct: true,
			})
			if err == nil {
				t.Fatal("error add product", err)
			}
		},
	}

	for desc, test := range cases {
		go func() {
			t.Log(desc + "\n")

			testCase := test

			t.Parallel()
			testCase(t)
		}()
	}
}

func TestProductFetching(t *testing.T) {
	db, cleanup, err := repository.ProvideDatabaseConnection(&cfg)
	if err != nil {
		t.Fatal("error connect to database", err)
	}

	productsInteractor := interactors.NewProductInteractor(slog.Default(), products.NewRepository(db))

	defer cleanup()

	t.Log("Test: products fetching\n")

	var cases map[string]Testcase = map[string]Testcase{
		"Normal case": func(t *testing.T) {
			productId := uuid.New()
			storageId := uuid.New()
			productName := gofakeit.ProductName()
			storageName := gofakeit.StreetName()
			size := rand.Int63n(250)
			amount := rand.Int63n(10000) + 1

			var reserved int64
			if amount > 1 {
				reserved = rand.Int63n(amount - 1)
			}

			available := amount - reserved

			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			err = insertStorages(ctx, db, insertStoragesParams{
				storageId:    storageId,
				storageName:  storageName,
				availability: randAvailability(),
			})
			if err != nil {
				t.Fatal("error add storage", err)
			}

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
			storageId := uuid.New()
			storageName := gofakeit.StreetName()

			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			err = insertStorages(ctx, db, insertStoragesParams{
				storageId:    storageId,
				storageName:  storageName,
				availability: randAvailability(),
			})
			if err != nil {
				t.Fatal("error add storage", err)
			}

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
			storageId := uuid.New()
			productName := gofakeit.ProductName()
			storageName := gofakeit.StreetName()
			size := rand.Int63n(250)
			amount := rand.Int63n(10000) + 1

			var reserved int64
			if amount > 1 {
				reserved = rand.Int63n(amount - 1)
			}

			available := amount - reserved

			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()

			err = insertStorages(ctx, db, insertStoragesParams{
				storageId:    storageId,
				storageName:  storageName,
				availability: randAvailability(),
			})
			if err != nil {
				t.Fatal("error add storage", err)
			}

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
		go func() {
			t.Log(desc + "\n")

			testCase := test

			t.Parallel()
			testCase(t)
		}()
	}
}
