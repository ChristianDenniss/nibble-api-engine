package catalog

import (
	"github.com/ChristianDenniss/api-engine/internal/httpx"
	"github.com/ChristianDenniss/go-data-model/catalog"
	"log"
	"net/http"
)

func Mount(mux *http.ServeMux, svc *catalog.Service) {
	mux.HandleFunc("GET /v1/catalog", func(w http.ResponseWriter, r *http.Request) {
		result, err := svc.Read(r.Context())
		if err != nil {
			log.Printf("catalog read: %v", err)
			httpx.WriteError(w, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		httpx.WriteJSON(w, http.StatusOK, result)
	})
}
