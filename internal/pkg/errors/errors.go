package errors

import (
	"errors"
	"fmt"
)

var (
	ErrorInternalServerError = errors.New("internal server error")
	ErrorUnexpectedData      = errors.New("unexpected data")
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
	// TODO map errors
	switch {
	case errors.Is(err, nil):
		return e.errorBuilder.Build(0, "")
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
