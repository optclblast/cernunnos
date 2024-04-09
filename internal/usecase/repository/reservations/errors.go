package reservations

import "errors"

var (
	ErrorNotEnoughSpace    = errors.New("not enough space")
	ErrorNotEnoughProducts = errors.New("not enough products")
)
