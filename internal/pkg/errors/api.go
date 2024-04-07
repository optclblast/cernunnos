package errors

import (
	"cernunnos/internal/usecase/repository/reservations"
	"errors"
	"fmt"
)

type APIError struct {
	Code    uint16 `json:"code"`
	Details string `json:"details"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("code: %v details: %s", e.Code, e.Details)
}

type ErrorHandler interface {
	Handle(err error) APIError
}

type errorHandler struct {
	errorBuilder ErrorBuilder
}

func NewErrorHandler() ErrorHandler {
	return &errorHandler{
		errorBuilder: NewErrorBuilder(),
	}
}

func (e *errorHandler) Handle(err error) APIError {
	switch {
	case errors.Is(err, ErrorBadRequest):
		return e.errorBuilder.Build(400, "Bad Request!")
	case errors.Is(err, ErrorInternalServerError):
		return e.errorBuilder.Build(500, "Internal Server Error!")
	case errors.Is(err, ErrorInvalidRequestPath):
		return e.errorBuilder.Build(400, "Invalid Path!")
	case errors.Is(err, ErrorUnexpectedData):
		return e.errorBuilder.Build(400, "Invalid Or Unexpected Request Data!")
	case errors.Is(err, reservations.ErrorNotEnoughSpace):
		return e.errorBuilder.Build(507, "Not Enough Space In Storage(s)!")
	default:
		return e.errorBuilder.Build(500, "Oops! Something went wrong!")
	}
}

type ErrorBuilder interface {
	Build(code uint16, details string) APIError
}

func NewErrorBuilder() ErrorBuilder {
	return new(errorBuilder)
}

type errorBuilder struct{}

func (e *errorBuilder) Build(code uint16, details string) APIError {
	return APIError{
		Code:    code,
		Details: details,
	}
}
