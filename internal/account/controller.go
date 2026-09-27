package account

import (
	"errors"
	"net/http"

	"github.com/ChristianDenniss/api-engine/internal/httpx"
	"github.com/ChristianDenniss/api-engine/internal/session"
	accountentity "github.com/ChristianDenniss/go-data-model/account/entity"
	accountsvc "github.com/ChristianDenniss/go-data-model/account/service"
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
		httpx.WriteFailure(w, http.StatusBadRequest, "account and address id required")
		return
	}
	if !authorizeAccount(w, r, accountID) {
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
		httpx.WriteFailure(w, http.StatusBadRequest, "account and address id required")
		return
	}
	if !authorizeAccount(w, r, accountID) {
		return
	}
	err := c.svc.SelectSavedAddress(r.Context(), accountID, addressID)
	if err != nil {
		writeAccountError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// authorizeAccount only lets the signed-in account change its own data.
func authorizeAccount(w http.ResponseWriter, r *http.Request, accountID string) bool {
	signedIn := session.AccountID(r.Context())
	if signedIn == "" {
		httpx.WriteFailure(w, http.StatusUnauthorized, "sign in required")
		return false
	}
	if signedIn != accountID {
		httpx.WriteFailure(w, http.StatusForbidden, "not your account")
		return false
	}
	return true
}

func writeAccountError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, accountentity.ErrIDRequired),
		errors.Is(err, accountentity.ErrAddressIDRequired):
		httpx.WriteFailure(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, accountentity.ErrNotFound),
		errors.Is(err, accountentity.ErrAddressNotFound):
		httpx.WriteFailure(w, http.StatusNotFound, err.Error())
	default:
		httpx.WriteFailure(w, http.StatusInternalServerError, "account request failed")
	}
}
