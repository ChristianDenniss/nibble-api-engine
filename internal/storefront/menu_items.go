package storefront

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ChristianDenniss/api-engine/internal/httpx"
	mediaentity "github.com/ChristianDenniss/go-data-model/media/entity"
	menuentity "github.com/ChristianDenniss/go-data-model/menu/entity"
	menusvc "github.com/ChristianDenniss/go-data-model/menu/service"
)

type MenuItemController struct {
	svc *menusvc.Service
}

func NewMenuItemController(svc *menusvc.Service) *MenuItemController {
	return &MenuItemController{svc: svc}
}

type setImageRequest struct {
	ImageURL *string `json:"imageURL"`
}

func (c *MenuItemController) SetImage(w http.ResponseWriter, r *http.Request) {
	var body setImageRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&body); err != nil || body.ImageURL == nil {
		httpx.WriteFailure(w, http.StatusBadRequest, `body must be {"imageURL": "<url or empty>"}`)
		return
	}
	item, err := c.svc.SetImageURL(r.Context(), r.PathValue("itemId"), *body.ImageURL)
	switch {
	case err == nil:
		httpx.WriteJSON(w, http.StatusOK, toItemWire(item))
	case errors.Is(err, menuentity.ErrIDRequired), errors.Is(err, mediaentity.ErrImageURLInvalid):
		httpx.WriteFailure(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, menuentity.ErrNotFound):
		httpx.WriteFailure(w, http.StatusNotFound, err.Error())
	default:
		httpx.WriteFailure(w, http.StatusInternalServerError, "failed to update menu item image")
	}
}
