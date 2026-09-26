package sourcestores

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

func (c *Controller) ListByChannel(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channelId")
	if channelID == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "channelId required"})
		return
	}
	stores, err := c.svc.ListStoresByChannel(r.Context(), channelID)
	if err != nil {
		if errors.Is(err, sourceentity.ErrIDRequired) {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list stores"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]interface{}{"stores": toStoreList(stores)})
}
