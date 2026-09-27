package catalog

import (
	"encoding/json"
	"errors"
	"github.com/ChristianDenniss/api-engine/internal/httpx"
	"github.com/ChristianDenniss/go-data-model/catalog"
	"io"
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
	mux.HandleFunc("POST /v1/catalog/cart-comparison", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var req catalog.CartRequest
		if err := decoder.Decode(&req); err != nil {
			httpx.WriteJSON(w, 400, map[string]string{"error": "Invalid cart request"})
			return
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			httpx.WriteJSON(w, 400, map[string]string{"error": "Expected one cart request"})
			return
		}
		result, err := svc.CompareCart(r.Context(), req)
		if errors.Is(err, catalog.ErrInvalidCart) {
			httpx.WriteJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		if err != nil {
			log.Printf("cart comparison: %v", err)
			httpx.WriteError(w, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		httpx.WriteJSON(w, 200, result)
	})

}
