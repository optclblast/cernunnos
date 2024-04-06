//go:build wireinject
// +build wireinject

package server

import (
	"cernunnos/internal/pkg/config"
	"cernunnos/internal/pkg/logger"
	"cernunnos/internal/server/interface/controllers"
	prodController "cernunnos/internal/server/interface/controllers/product"
	resController "cernunnos/internal/server/interface/controllers/reservation"
	stController "cernunnos/internal/server/interface/controllers/storage"
	"cernunnos/internal/usecase/repository"
	productsRepo "cernunnos/internal/usecase/repository/products"
	reservationsRepo "cernunnos/internal/usecase/repository/reservations"
	storagesRepo "cernunnos/internal/usecase/repository/storages"
	"database/sql"
	"log/slog"

	"github.com/google/wire"
)

func ProvideServer(c *config.Config) (*Server, func(), error) {
	wire.Build(
		repository.ProvideDatabaseConnection,
		provideStoragesRepository,
		provideProductsRepository,
		provideReservationsRepository,
		provideLogger,
		prodController.NewProductController,
		stController.NewStorageController,
		resController.NewReservationController,
		controllers.NewRootController,
		newServer,
	)
	return &Server{}, func() {}, nil
}

func provideStoragesRepository(db *sql.DB) storagesRepo.Repository {
	return storagesRepo.NewRepository(db)
}

func provideProductsRepository(db *sql.DB) productsRepo.Repository {
	return productsRepo.NewRepository(db)
}

func provideReservationsRepository(db *sql.DB) reservationsRepo.Repository {
	return reservationsRepo.NewRepository(db)
}

func provideLogger(c *config.Config) *slog.Logger {
	return logger.NewLogger(logger.MapLevel(c.LogLevel))
}
