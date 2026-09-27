package sponsored

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ChristianDenniss/api-engine/internal/httpx"
	merchentity "github.com/ChristianDenniss/go-data-model/merchandising/entity"
	merchsvc "github.com/ChristianDenniss/go-data-model/merchandising/service"
)

type Controller struct {
	svc *merchsvc.Service
}

func NewController(svc *merchsvc.Service) *Controller {
	return &Controller{svc: svc}
}

type eventRequest struct {
	PlacementID string `json:"placement_id"`
	Kind        string `json:"kind"`
	Surface     string `json:"surface"`
}

type eventResponse struct {
	ID         string `json:"id"`
	OccurredAt string `json:"occurred_at"`
}

func (c *Controller) RecordEvent(w http.ResponseWriter, r *http.Request) {
	var req eventRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		httpx.WriteFailure(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	event, err := c.svc.RecordEvent(r.Context(), merchentity.Event{
		PlacementID: req.PlacementID,
		Kind:        req.Kind,
		Surface:     req.Surface,
	})
	if err != nil {
		switch {
		case errors.Is(err, merchentity.ErrIDRequired), errors.Is(err, merchentity.ErrEventKindInvalid):
			httpx.WriteFailure(w, http.StatusBadRequest, err.Error())
		default:
			httpx.WriteFailure(w, http.StatusInternalServerError, "failed to record event")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, eventResponse{ID: event.ID, OccurredAt: event.OccurredAt.Format("2006-01-02T15:04:05Z07:00")})
}
