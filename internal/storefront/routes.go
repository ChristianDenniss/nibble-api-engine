package storefront

import "net/http"

func Mount(mux *http.ServeMux, c *Controller) {
	mux.HandleFunc("GET /v1/storefront", c.Get)
	mux.HandleFunc("GET /storefront", c.Get)
	mux.HandleFunc("GET /v1/restaurants", c.ListRestaurants)
	mux.HandleFunc("GET /v1/menu-items", c.ListMenuItems)
	mux.HandleFunc("PUT /v1/cart", c.ReplaceCart)
	mux.HandleFunc("PUT /cart", c.ReplaceCart)
	mux.HandleFunc("POST /v1/cart/events", c.RecordCartEvent)
	mux.HandleFunc("POST /cart/events", c.RecordCartEvent)
}

func MountMenuItems(mux *http.ServeMux, c *MenuItemController) {
	mux.HandleFunc("PUT /v1/menu-items/{itemId}/image", c.SetImage)
}
