package storefront

import (
	"errors"
	"net/http"
	"os"

	storefrontentity "github.com/ChristianDenniss/go-data-model/storefront/entity"
	storefrontsvc "github.com/ChristianDenniss/go-data-model/storefront/service"
	"github.com/ChristianDenniss/api-engine/internal/httpx"
)

type Controller struct {
	svc            *storefrontsvc.Service
	defaultAccount string
}

func NewController(svc *storefrontsvc.Service) *Controller {
	return &Controller{
		svc:            svc,
		defaultAccount: getenv("DEFAULT_ACCOUNT_ID", "acct_dev"),
	}
}

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("account_id")
	if accountID == "" {
		accountID = c.defaultAccount
	}
	catalog, err := c.svc.LoadCatalog(r.Context(), accountID)
	if err != nil {
		if errors.Is(err, storefrontentity.ErrAccountIDRequired) {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, storefrontentity.ErrNotFound) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "catalog not found"})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load catalog"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toCatalogResponse(catalog))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
