package apperror

import "net/http"

func BadRequest(message string) *Error {
	return New(http.StatusBadRequest, message)
}
