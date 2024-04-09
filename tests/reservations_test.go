package tests

import (
	"cernunnos/internal/usecase/interactors"
	"cernunnos/internal/usecase/repository"
	"cernunnos/internal/usecase/repository/reservations"
	"context"
	"log/slog"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

func TestReservationsList(t *testing.T) {
	db, cleanup, err := repository.ProvideDatabaseConnection(&cfg)
	if err != nil {
		t.Fatal("error connect to database", err)
	}

	defer cleanup()

	reservationsInteractor := interactors.NewReservationInteractor(slog.Default(), reservations.NewRepository(db))

	t.Log("Test: reservations fetching\n")

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
				t.FailNow()
			}

			shippingId := uuid.New()

			reserve := available / 2

			err = reservationsInteractor.Reserve(ctx, interactors.ReserveParams{
				ProductIds: []string{productId.String()},
				StorageId:  storageId.String(),
				ShippingId: shippingId.String(),
				Amount:     available / 2,
			})
			if err != nil {
				t.Fatal("error reserve product", err)
			}

			reservations, err := reservationsInteractor.Reservations(
				ctx,
				interactors.ReservationsParams{
					ProductId:  productId.String(),
					StorageId:  storageId.String(),
					ShippingId: shippingId.String(),
					Limit:      1,
				},
			)
			if err != nil {
				t.Fatal("error fetch product reservation", err)
			}

			if len(reservations) == 0 {
				t.Fatal("error empty product reservation", err)
			}

			if reservations[0].Reserved != reserve {
				t.Fatal("error invalid product reservation reserved value", err)
			}
		},
		"Not enough products case": func(t *testing.T) {
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

			shippingId := uuid.New()

			err = reservationsInteractor.Reserve(ctx, interactors.ReserveParams{
				ProductIds: []string{productId.String()},
				StorageId:  storageId.String(),
				ShippingId: shippingId.String(),
				Amount:     amount * 10,
			})
			if err == nil {
				t.Fatal("error reserve product", err)
			}
		},
		"Reservin zero products case": func(t *testing.T) {
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

			shippingId := uuid.New()

			err = reservationsInteractor.Reserve(ctx, interactors.ReserveParams{
				ProductIds: []string{productId.String()},
				StorageId:  storageId.String(),
				ShippingId: shippingId.String(),
				Amount:     0,
			})
			if err == nil {
				t.Fatal("error reserve product", err)
			}
		},
		"Reservin negative amount of products case": func(t *testing.T) {
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

			shippingId := uuid.New()

			err = reservationsInteractor.Reserve(ctx, interactors.ReserveParams{
				ProductIds: []string{productId.String()},
				StorageId:  storageId.String(),
				ShippingId: shippingId.String(),
				Amount:     -1,
			})
			if err == nil {
				t.Fatal("error reserve product", err)
			}
		},
	}

	for desc, test := range cases {
		t.Log(desc + "\n")

		test(t)
	}
}
