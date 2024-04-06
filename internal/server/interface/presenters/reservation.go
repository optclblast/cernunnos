package presenters

import "cernunnos/internal/pkg/models"

type ReservationPresenter interface {
	ResponseReservations(reservations []*models.Reservation) ([]byte, error)
}

type reservationPresenter struct{}

func (p *reservationPresenter) ResponseReservations(reservations []*models.Reservation) ([]byte, error) {
	return nil, nil
}
