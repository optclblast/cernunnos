package presenters

import (
	"cernunnos/internal/pkg/dto"
	"cernunnos/internal/pkg/models"
	"encoding/json"
	"fmt"
)

type ReservationPresenter interface {
	ResponseReservations(reservations []*models.Reservation) ([]byte, error)
	ResponseReserve() ([]byte, error)
	ResponseCancel() ([]byte, error)
	ResponseRelease() ([]byte, error)
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
		Offset:       uint32(len(mappedReservations)),
	}

	rawResponse, err := json.Marshal(&response)
	if err != nil {
		return nil, fmt.Errorf("error marshal response. %w", err)
	}

	return rawResponse, nil
}

func (p *reservationPresenter) ResponseReserve() ([]byte, error) {
	response := &dto.ReserveResponse{
		Ok: true,
	}

	rawResponse, err := json.Marshal(&response)
	if err != nil {
		return nil, fmt.Errorf("error marshal response. %w", err)
	}

	return rawResponse, nil
}

func (p *reservationPresenter) ResponseCancel() ([]byte, error) {
	response := &dto.CancelResponse{
		Ok: true,
	}

	rawResponse, err := json.Marshal(&response)
	if err != nil {
		return nil, fmt.Errorf("error marshal response. %w", err)
	}

	return rawResponse, nil
}

func (p *reservationPresenter) ResponseRelease() ([]byte, error) {
	response := &dto.ReleaseResponse{
		Ok: true,
	}

	rawResponse, err := json.Marshal(&response)
	if err != nil {
		return nil, fmt.Errorf("error marshal response. %w", err)
	}

	return rawResponse, nil
}
