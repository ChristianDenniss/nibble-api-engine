package pagination

import (
	"net/http"
	"strconv"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// Query is the pagination contract passed from HTTP middleware through
// controllers into services, matching the dashboard page/pageSize/all shape.
type Query struct {
	Page     int
	PageSize int
	All      bool
}

type Result[T any] struct {
	Data       []T `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalPages int `json:"totalPages"`
}

func FromRequest(r *http.Request) Query {
	q := Query{Page: defaultPage, PageSize: defaultPageSize}
	if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && page >= 1 {
		q.Page = page
	}
	if size, err := strconv.Atoi(r.URL.Query().Get("pageSize")); err == nil && size >= 1 {
		if size > maxPageSize {
			size = maxPageSize
		}
		q.PageSize = size
	}
	q.All = r.URL.Query().Get("all") == "true"
	return q
}

func ToPaginated[T any](data []T, total, page, pageSize int) Result[T] {
	if pageSize < 1 {
		pageSize = 1
	}
	totalPages := (total + pageSize - 1) / pageSize
	return Result[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

func Offset(q Query) (offset, limit int) {
	if q.All {
		return 0, 0
	}
	return (q.Page - 1) * q.PageSize, q.PageSize
}
