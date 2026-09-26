package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/ChristianDenniss/api-engine/internal/apperror"
)

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(body)
}

func WritePlain(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func WriteError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*apperror.Error); ok {
		WriteJSON(w, appErr.StatusCode, map[string]string{"error": appErr.Message})
		return
	}
	WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
