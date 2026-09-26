package apperror

import "net/http"

func Unavailable(message string) *Error {
	if message == "" {
		message = "service unavailable"
	}
	return New(http.StatusServiceUnavailable, message)
}
