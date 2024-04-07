package presenters

import (
	"cernunnos/internal/pkg/dto"
	"cernunnos/internal/pkg/models"
	"encoding/json"
	"fmt"
)

type ReservationPresenter interface {
	ResponseReservations(reservations []*models.Reservation) ([]byte, error)
}

func NewReservationPresenter() ReservationPresenter {
	return new(reservationPresenter)
}

type reservationPresenter struct{}

func (p *reservationPresenter) ResponseReservations(reservations []*models.Reservation) ([]byte, error) {
	mappedReservations, err := dto.MapReservationsFromModels(reservations)
	if err != nil {
		return nil, fmt.Errorf("error map reservations from models. %w", err)
	}

	response := &dto.ReservationsResponse{
		Reservations: mappedReservations,
	}

	rawResponse, err := json.Marshal(&response)
	if err != nil {
		return nil, fmt.Errorf("error marshal response. %w", err)
	}

	return rawResponse, nil
}
