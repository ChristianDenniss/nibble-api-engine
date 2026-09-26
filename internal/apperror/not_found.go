package apperror

import "net/http"

func NotFound(resource string) *Error {
	if resource == "" {
		return New(http.StatusNotFound, "not found")
	}
	return New(http.StatusNotFound, resource+" not found")
}
