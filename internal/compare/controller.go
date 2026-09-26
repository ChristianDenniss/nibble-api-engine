package compare

import (
	"encoding/json"
	"net/http"

	compareentity "github.com/ChristianDenniss/go-data-model/compare/entity"
	comparesvc "github.com/ChristianDenniss/go-data-model/compare/service"
	userentity "github.com/ChristianDenniss/go-data-model/user/entity"
	usersvc "github.com/ChristianDenniss/go-data-model/user/service"
	"github.com/ChristianDenniss/api-engine/internal/httpx"
)

type Controller struct {
	compare *comparesvc.Service
	users   *usersvc.Service
}

func NewController(compare *comparesvc.Service, users *usersvc.Service) *Controller {
	return &Controller{compare: compare, users: users}
}

type compareRequestBody struct {
	PlaceID            string                         `json:"place_id"`
	FulfillmentContext compareentity.FulfillmentContext `json:"fulfillment_context"`
	Basket             compareentity.Basket           `json:"basket"`
	Filters            userentity.ComparePrefs        `json:"filters"`
	Memberships        []string                       `json:"memberships"`
	QuotePreference    string                         `json:"quote_preference"`
	UserID             string                         `json:"user_id"`
}

func (c *Controller) PostCompare(w http.ResponseWriter, r *http.Request) {
	var body compareRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	req := compareentity.Request{
		PlaceID:            body.PlaceID,
		FulfillmentContext: body.FulfillmentContext,
		Basket:             body.Basket,
		Filters:            body.Filters,
		Memberships:        body.Memberships,
		QuotePreference:    body.QuotePreference,
	}
	result, session, err := c.compare.Compare(r.Context(), body.UserID, req)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	var payload map[string]interface{}
	if len(session.ResultSnapshot) > 0 {
		_ = json.Unmarshal(session.ResultSnapshot, &payload)
	} else {
		payload = map[string]interface{}{
			"compare_session_id": result.CompareSessionID,
			"recommendation":     result.Recommendation,
			"runners_up":         result.RunnersUp,
		}
	}
	httpx.WriteJSON(w, http.StatusOK, payload)
}

func (c *Controller) GetSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	session, err := c.users.GetSession(r.Context(), id)
	if err != nil {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		return
	}
	var result interface{}
	if len(session.ResultSnapshot) > 0 {
		_ = json.Unmarshal(session.ResultSnapshot, &result)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":              session.ID,
		"user_id":         session.UserID,
		"place_id":        session.PlaceID,
		"query_snapshot":  session.QuerySnapshot,
		"basket_snapshot": session.BasketSnapshot,
		"result":          result,
	})
}
