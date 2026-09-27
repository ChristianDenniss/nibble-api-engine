package home

import (
	"errors"
	"net/http"
	"os"

	"github.com/ChristianDenniss/api-engine/internal/httpx"
	"github.com/ChristianDenniss/api-engine/internal/session"
	homesvc "github.com/ChristianDenniss/go-data-model/home/service"
	storefrontentity "github.com/ChristianDenniss/go-data-model/storefront/entity"
)

type Controller struct {
	svc            *homesvc.Service
	defaultAccount string
}

func NewController(svc *homesvc.Service) *Controller {
	account := os.Getenv("DEFAULT_ACCOUNT_ID")
	if account == "" {
		account = "acct_dev"
	}
	return &Controller{svc: svc, defaultAccount: account}
}

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	accountID := session.AccountID(r.Context())
	if accountID == "" {
		accountID = r.URL.Query().Get("account_id")
	}
	if accountID == "" {
		accountID = c.defaultAccount
	}
	feed, err := c.svc.BuildFeed(r.Context(), accountID)
	if err != nil {
		switch {
		case errors.Is(err, storefrontentity.ErrAccountIDRequired):
			httpx.WriteFailure(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, storefrontentity.ErrNotFound):
			httpx.WriteFailure(w, http.StatusNotFound, "account not found")
		default:
			httpx.WriteFailure(w, http.StatusInternalServerError, "failed to build home feed")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toFeedResponse(feed))
}
