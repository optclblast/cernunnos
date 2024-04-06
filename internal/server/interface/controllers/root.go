package controllers

type RootController struct {
	ProductController     ProductController
	ReservationController ReservationController
	StorageController     StorageController
}

func NewRootController(
	productController ProductController,
	reservationController ReservationController,
	storageController StorageController,
) *RootController {
	return &RootController{
		ProductController:     productController,
		ReservationController: reservationController,
		StorageController:     storageController,
	}
}
