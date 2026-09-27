package httpx

import (
	"net/http"

	"github.com/ChristianDenniss/api-engine/internal/httpx/middleware"
)

// Wrap applies global HTTP middleware. Module routes are mounted on mux by
// the composition root so this package never imports a domain module.
func Wrap(mux http.Handler) http.Handler {
	return middleware.RequestID(
		middleware.SecurityHeaders(
			middleware.CORS(
				middleware.RequestLimits(
					middleware.Timeout(
						middleware.RateLimit(middleware.RequestLog(
							middleware.Recover(middleware.Pagination(mux)),
						)),
					),
				),
			),
		),
	)
}
