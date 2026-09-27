package sourcestores

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

func (c *Controller) ListByChannel(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channelId")
	if channelID == "" {
		httpx.WriteFailure(w, http.StatusBadRequest, "channelId required")
		return
	}
	stores, err := c.svc.ListStoresByChannel(r.Context(), channelID)
	if err != nil {
		if errors.Is(err, sourceentity.ErrIDRequired) {
			httpx.WriteFailure(w, http.StatusBadRequest, err.Error())
			return
		}
		httpx.WriteFailure(w, http.StatusInternalServerError, "failed to list stores")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]interface{}{"stores": toStoreList(stores)})
}
