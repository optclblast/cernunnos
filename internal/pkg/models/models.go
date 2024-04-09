package models

import (
	"time"

	"github.com/google/uuid"
)

type Storage struct {
	Id        uuid.UUID
	Name      string
	Reserved  int64
	Available int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ProductInfo struct {
	Id               uuid.UUID
	Name             string
	Size             int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DestributionInfo []*ProductDestribution
}

type StorageProduct struct {
	ProductInfo
	ProductDestribution
}

type Reservation struct {
	StorageId       uuid.UUID
	ReservedProduct *StorageProduct
	ShippingId      uuid.UUID
	Reserved        int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ProductDestribution struct {
	Storage   *Storage
	Amount    int64
	Reserved  int64
	Available int64
}
