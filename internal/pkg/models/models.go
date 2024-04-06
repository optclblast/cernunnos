package models

import (
	"time"

	"github.com/google/uuid"
)

type StorageAvailability string

const (
	StorageAvailabilityAvailable   StorageAvailability = "available"
	StorageAvailabilityBusy        StorageAvailability = "busy"
	StorageAvailabilityUnavailable StorageAvailability = "unavailable"
)

type Storage struct {
	Id           uuid.UUID
	Name         string
	Availability StorageAvailability
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
	Storage  *Storage
	Amount   int64
	Reserved int64
}
