package sourcemenu

import (
	"errors"
	"net/http"

	"github.com/ChristianDenniss/api-engine/internal/httpx"
	sourceentity "github.com/ChristianDenniss/go-data-model/source/entity"
	sourcesvc "github.com/ChristianDenniss/go-data-model/source/service"
)

type Controller struct {
	svc *sourcesvc.Service
}

func NewController(svc *sourcesvc.Service) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) GetMenu(w http.ResponseWriter, r *http.Request) {
	storeID := r.PathValue("storeId")
	if storeID == "" {
		httpx.WriteFailure(w, http.StatusBadRequest, "storeId required")
		return
	}
	fulfillment := r.URL.Query().Get("fulfillment_mode")
	if fulfillment == "" {
		httpx.WriteFailure(w, http.StatusBadRequest, "fulfillment_mode required")
		return
	}
	deliveryExecutor := r.URL.Query().Get("delivery_executor")

	browse, err := c.svc.LoadMenuBrowse(r.Context(), storeID, fulfillment, deliveryExecutor)
	if err != nil {
		if errors.Is(err, sourceentity.ErrIDRequired) {
			httpx.WriteFailure(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, sourceentity.ErrNotFound) {
			httpx.WriteFailure(w, http.StatusNotFound, "menu not found")
			return
		}
		httpx.WriteFailure(w, http.StatusInternalServerError, "failed to load menu")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toMenuBrowseResponse(browse))
}
