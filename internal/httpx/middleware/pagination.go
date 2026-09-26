package middleware

import (
	"context"
	"net/http"

	"github.com/ChristianDenniss/api-engine/internal/pagination"
)

type ctxKey int

const paginationKey ctxKey = 1

func Pagination(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), paginationKey, pagination.FromRequest(r))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Query(r *http.Request) pagination.Query {
	q, _ := r.Context().Value(paginationKey).(pagination.Query)
	if q.Page == 0 {
		return pagination.FromRequest(r)
	}
	return q
}
