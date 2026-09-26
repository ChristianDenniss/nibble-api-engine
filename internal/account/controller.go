package account

import (
	"errors"
	"net/http"

	accountentity "github.com/ChristianDenniss/go-data-model/account/entity"
	accountsvc "github.com/ChristianDenniss/go-data-model/account/service"
	"github.com/ChristianDenniss/api-engine/internal/httpx"
)

type Controller struct {
	svc *accountsvc.Service
}

func NewController(svc *accountsvc.Service) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("accountId")
	addressID := r.PathValue("addressId")
	if accountID == "" || addressID == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "account and address id required"})
		return
	}
	err := c.svc.RemoveSavedAddress(r.Context(), accountID, addressID)
	if err != nil {
		writeAccountError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) SetCurrentAddress(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("accountId")
	addressID := r.PathValue("addressId")
	if accountID == "" || addressID == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "account and address id required"})
		return
	}
	err := c.svc.SelectSavedAddress(r.Context(), accountID, addressID)
	if err != nil {
		writeAccountError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeAccountError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, accountentity.ErrIDRequired),
		errors.Is(err, accountentity.ErrAddressIDRequired):
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, accountentity.ErrNotFound),
		errors.Is(err, accountentity.ErrAddressNotFound):
		httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	default:
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "account request failed"})
	}
}
