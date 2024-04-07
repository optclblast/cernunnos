package errors

import "errors"

var (
	ErrorInternalServerError = errors.New("INTERNAL_SERVER_ERROR")
	ErrorUnexpectedData      = errors.New("UNEXPECTED_DATA")
	ErrorInvalidRequestPath  = errors.New("INVALID_REQUEST_PATH")
	ErrorBadRequest          = errors.New("BAD_REQUEST")
)
