package dto

type StoragesRequest struct {
	Ids             []string `json:"ids,omitempty"`
	WithBusy        bool     `json:"with_busy,omitempty"`
	WithUnavailable bool     `json:"with_unavailable,omitempty"`
	Limit           uint32   `json:"limit,omitempty"`
	Offset          uint32   `json:"offset,omitempty"`
}

type StoragesResponse struct {
	Storages []*Storage `json:"storages"`
	Offset   uint32     `json:"offset"`
}

// Storage DTO object
type Storage struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Availability string `json:"availability"` // possible values: "available", "unavailable", "busy"
	CreatedAt    uint64 `json:"created_at"`   // unix milli
	UpdatedAt    uint64 `json:"updated_at"`   // unix milli
}

type StorageProductsRequest struct {
	StorageId       string   // Fetched from URL params
	ProductsIds     []string `json:"products_ids,omitempty"`
	WithUnavailable bool     `json:"with_unavailable,omitempty"`
	Limit           uint32   `json:"limit,omitempty"`
	Offset          uint32   `json:"offset,omitempty"`
}

type StorageProductsResponse struct {
	Products []*StorageProduct `json:"products"`
	Offset   uint32            `json:"offset"`
}

type StorageProduct struct {
	ProductInfo
	ProductDestribution
}

type ProductInfo struct {
	Id               string                 `json:"id"`
	Name             string                 `json:"name"`
	Size             int64                  `json:"size"`
	CreatedAt        uint64                 `json:"created_at"` // unix milli
	UpdatedAt        uint64                 `json:"updated_at"` // unix milli
	DestributionInfo []*ProductDestribution `json:"destribution_info,omitempty"`
}

type ProductsRequest struct {
	Ids              []string `json:"ids,omitempty"`
	StorageId        string   `json:"storage_id,omitempty"`
	WithDistribution bool     `json:"with_distribution,omitempty"` // Include product destribution info into response
	WithUnavailable  bool     `json:"with_unavailable,omitempty"`  // Fetch with unavailable products
	Limit            uint32   `json:"limit"`                       // Amount of items to fetch. Default and max 500
	Offset           uint32   `json:"offset"`                      // Pagination
}

type ProductsResponse struct {
	Products []*ProductInfo `json:"products"`
	Offset   uint32         `json:"offset"`
}

type ProductDestribution struct {
	StorageId string `json:"storage_id,omitempty"`
	Amount    int64  `json:"amount"`
	Reserved  int64  `json:"reserved"`
}

type Reservation struct {
	StorageId  string `json:"storage_id"`
	ProductId  string `json:"product_id"`
	ShippingId string `json:"shipping_id"`
	Reserved   int64  `json:"reserved"`
	CreatedAt  uint64 `json:"created_at"` // unix milli
	UpdatedAt  uint64 `json:"updated_at"` // unix milli
}

type ReservationsRequest struct {
	StorageId  string `json:"storage_id,omitempty"`
	ProductId  string `json:"product_id,omitempty"`
	ShippingId string `json:"shipping_id,omitempty"`
	Limit      uint32 `json:"limit,omitempty"`
	Offset     uint32 `json:"offset,omitempty"`
}

type ReservationsResponse struct {
	Reservations []*Reservation `json:"reservations"`
	Offset       uint32         `json:"offset"`
}

type ReserveRequest struct {
	StorageId  string   `json:"storage_id,omitempty"`
	Products   []string `json:"products,omitempty"`
	ShippingId string   `json:"shipping_id,omitempty"`
	Amount     int64    `json:"amount"`
}

type ReserveResponse struct{}

type ReleaseRequest struct {
	StorageId  string   `json:"storage_id,omitempty"`
	Products   []string `json:"products,omitempty"`
	ShippingId string   `json:"shipping_id,omitempty"`
}

type ReleaseResponse struct{}

type CancelRequest struct {
	StorageId  string   `json:"storage_id,omitempty"`
	Products   []string `json:"products,omitempty"`
	ShippingId string   `json:"shipping_id,omitempty"`
}

type CancelResponse struct{}
