package health

import (
	"net/http"

	"github.com/ChristianDenniss/api-engine/internal/httpx"
)

type Controller struct {
	svc *Service
}

func NewController(svc *Service) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	if err := c.svc.Check(r.Context()); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WritePlain(w, http.StatusOK, "ok")
}
