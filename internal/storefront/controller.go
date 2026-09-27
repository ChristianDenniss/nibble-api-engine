package storefront

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ChristianDenniss/api-engine/internal/httpx"
	"github.com/ChristianDenniss/api-engine/internal/httpx/middleware"
	"github.com/ChristianDenniss/api-engine/internal/pagination"
	"github.com/ChristianDenniss/api-engine/internal/session"
	cartentity "github.com/ChristianDenniss/go-data-model/cart/entity"
	cartsvc "github.com/ChristianDenniss/go-data-model/cart/service"
	catalogdomain "github.com/ChristianDenniss/go-data-model/catalog"
	storefrontentity "github.com/ChristianDenniss/go-data-model/storefront/entity"
	storefrontrepo "github.com/ChristianDenniss/go-data-model/storefront/repository"
	storefrontsvc "github.com/ChristianDenniss/go-data-model/storefront/service"
	userentity "github.com/ChristianDenniss/go-data-model/user/entity"
	userrepo "github.com/ChristianDenniss/go-data-model/user/repository"
)

type Controller struct {
	svc             *storefrontsvc.Service
	cart            *cartsvc.Service
	analytics       userrepo.OutboundClickRepository
	defaultAccount  string
	capturedCatalog *catalogdomain.Service
}

func NewController(svc *storefrontsvc.Service, carts *cartsvc.Service, analytics userrepo.OutboundClickRepository, captured ...*catalogdomain.Service) *Controller {
	var capturedCatalog *catalogdomain.Service
	if len(captured) > 0 {
		capturedCatalog = captured[0]
	}
	return &Controller{
		svc:             svc,
		cart:            carts,
		analytics:       analytics,
		defaultAccount:  getenv("DEFAULT_ACCOUNT_ID", "acct_dev"),
		capturedCatalog: capturedCatalog,
	}
}

type cartEventRequest struct {
	Kind       string `json:"kind"`
	ItemID     string `json:"itemId"`
	ProviderID string `json:"providerId"`
	TargetURL  string `json:"targetURL"`
}

func (c *Controller) RecordCartEvent(w http.ResponseWriter, r *http.Request) {
	var body cartEventRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Kind == "" {
		httpx.WriteFailure(w, http.StatusBadRequest, "invalid cart event")
		return
	}
	if c.analytics != nil {
		id := "cart_event_" + strconv.FormatInt(time.Now().UnixNano(), 10)
		_ = c.analytics.Insert(r.Context(), userentity.OutboundClick{ID: id, ActionKind: "cart_" + body.Kind, TargetURL: body.TargetURL})
	}
	httpx.WriteJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

type replaceCartRequest struct {
	ID    string         `json:"id"`
	Lines []cartLineWire `json:"lines"`
}

func (c *Controller) ReplaceCart(w http.ResponseWriter, r *http.Request) {
	accountID := session.AccountID(r.Context())
	if accountID == "" {
		accountID = r.URL.Query().Get("account_id")
	}
	if accountID == "" {
		accountID = c.defaultAccount
	}
	var body replaceCartRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteFailure(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(body.Lines) > 100 {
		httpx.WriteFailure(w, http.StatusBadRequest, "cart has too many lines")
		return
	}
	cartID := body.ID
	if cartID == "" {
		cartID = "cart_" + accountID
	}
	if existing, err := c.cart.GetByID(r.Context(), cartID); err == nil && existing.AccountID != accountID {
		httpx.WriteFailure(w, http.StatusForbidden, "cart does not belong to account")
		return
	}
	lines := make([]cartentity.CartLine, 0, len(body.Lines))
	for i, line := range body.Lines {
		if line.MenuItemID == "" || line.RestaurantID == "" || line.ProviderID == "" || line.Quantity < 1 || line.Quantity > 50 {
			httpx.WriteFailure(w, http.StatusBadRequest, "invalid cart line")
			return
		}
		id := line.ID
		if id == "" {
			id = cartID + "_line_" + strconv.Itoa(i)
		}
		lines = append(lines, cartentity.CartLine{ID: id, RestaurantID: line.RestaurantID, MenuItemID: line.MenuItemID, ProviderID: line.ProviderID, Quantity: line.Quantity})
	}
	cart := cartentity.Cart{ID: cartID, AccountID: accountID, Lines: lines}
	if err := c.cart.Replace(r.Context(), cart); err != nil {
		httpx.WriteFailure(w, http.StatusInternalServerError, "failed to save cart")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toCartWire(cart))
}

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	accountID := session.AccountID(r.Context())
	if accountID == "" {
		accountID = r.URL.Query().Get("account_id")
	}
	if accountID == "" {
		accountID = c.defaultAccount
	}
	var catalog storefrontentity.Catalog
	var err error
	if r.URL.Query().Get("lightweight") == "true" {
		catalog, err = c.svc.LoadBootstrap(r.Context(), accountID)
	} else {
		catalog, err = c.svc.LoadCatalog(r.Context(), accountID)
	}
	if err != nil {
		if errors.Is(err, storefrontentity.ErrAccountIDRequired) {
			httpx.WriteFailure(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, storefrontentity.ErrNotFound) {
			httpx.WriteFailure(w, http.StatusNotFound, "catalog not found")
			return
		}
		httpx.WriteFailure(w, http.StatusInternalServerError, "failed to load catalog")
		return
	}
	response := toCatalogResponse(catalog)
	if c.capturedCatalog != nil && r.URL.Query().Get("lightweight") != "true" {
		if captured, readErr := c.capturedCatalog.Read(r.Context()); readErr == nil {
			mergeCapturedCatalog(&response, captured)
		}
	}
	httpx.WriteJSON(w, http.StatusOK, response)
}

type paginatedRestaurantsResponse struct {
	Data       []restaurantWire `json:"data"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
	TotalPages int              `json:"totalPages"`
}

type paginatedItemsResponse struct {
	Data       []itemWire `json:"data"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
	TotalPages int        `json:"totalPages"`
}

func (c *Controller) ListRestaurants(w http.ResponseWriter, r *http.Request) {
	p := middleware.Query(r)
	q := storefrontrepo.RestaurantQuery{
		Search: r.URL.Query().Get("q"), CategorySlug: r.URL.Query().Get("category"),
		CuisineSlug: r.URL.Query().Get("cuisine"), Page: p.Page, PageSize: p.PageSize,
		Sort: r.URL.Query().Get("sort"),
	}
	result, err := c.svc.SearchRestaurants(r.Context(), q)
	if err != nil {
		httpx.WriteFailure(w, http.StatusInternalServerError, "failed to list restaurants")
		return
	}
	meta := pagination.ToPaginated(mapRestaurants(result.Restaurants), result.Total, p.Page, p.PageSize)
	httpx.WriteJSON(w, http.StatusOK, paginatedRestaurantsResponse{Data: meta.Data, Total: meta.Total, Page: meta.Page, PageSize: meta.PageSize, TotalPages: meta.TotalPages})
}

func (c *Controller) ListMenuItems(w http.ResponseWriter, r *http.Request) {
	p := middleware.Query(r)
	q := storefrontrepo.MenuItemQuery{
		Search: r.URL.Query().Get("q"), RestaurantID: r.URL.Query().Get("restaurantId"),
		CategorySlug: r.URL.Query().Get("category"), CuisineSlug: r.URL.Query().Get("cuisine"),
		Page: p.Page, PageSize: p.PageSize,
	}
	result, err := c.svc.SearchMenuItems(r.Context(), q)
	if err != nil {
		httpx.WriteFailure(w, http.StatusInternalServerError, "failed to list menu items")
		return
	}
	meta := pagination.ToPaginated(mapItems(result.Items), result.Total, p.Page, p.PageSize)
	httpx.WriteJSON(w, http.StatusOK, paginatedItemsResponse{Data: meta.Data, Total: meta.Total, Page: meta.Page, PageSize: meta.PageSize, TotalPages: meta.TotalPages})
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
