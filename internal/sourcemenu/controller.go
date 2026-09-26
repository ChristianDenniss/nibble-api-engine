package sourcemenu

import (
	"errors"
	"net/http"

	sourceentity "github.com/ChristianDenniss/go-data-model/source/entity"
	sourcesvc "github.com/ChristianDenniss/go-data-model/source/service"
	"github.com/ChristianDenniss/api-engine/internal/httpx"
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
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "storeId required"})
		return
	}
	fulfillment := r.URL.Query().Get("fulfillment_mode")
	if fulfillment == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "fulfillment_mode required"})
		return
	}
	deliveryExecutor := r.URL.Query().Get("delivery_executor")

	browse, err := c.svc.LoadMenuBrowse(r.Context(), storeID, fulfillment, deliveryExecutor)
	if err != nil {
		if errors.Is(err, sourceentity.ErrIDRequired) {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, sourceentity.ErrNotFound) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "menu not found"})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load menu"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toMenuBrowseResponse(browse))
}
