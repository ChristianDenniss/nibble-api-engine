package storefront

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	menuentity "github.com/ChristianDenniss/go-data-model/menu/entity"
	restaurantentity "github.com/ChristianDenniss/go-data-model/restaurant/entity"
	storefrontentity "github.com/ChristianDenniss/go-data-model/storefront/entity"
	storefrontrepo "github.com/ChristianDenniss/go-data-model/storefront/repository"
	storefrontsvc "github.com/ChristianDenniss/go-data-model/storefront/service"
)

type searchRepository struct {
	restaurantQuery storefrontrepo.RestaurantQuery
	itemQuery       storefrontrepo.MenuItemQuery
}

func (r *searchRepository) LoadCatalog(context.Context, string) (storefrontentity.Catalog, error) {
	return storefrontentity.Catalog{}, nil
}

func (r *searchRepository) SearchRestaurants(_ context.Context, query storefrontrepo.RestaurantQuery) (storefrontrepo.RestaurantPage, error) {
	r.restaurantQuery = query
	return storefrontrepo.RestaurantPage{
		Restaurants: []restaurantentity.Restaurant{{ID: "rest_2", Name: "Second"}},
		Total:       3,
	}, nil
}

func (r *searchRepository) SearchMenuItems(_ context.Context, query storefrontrepo.MenuItemQuery) (storefrontrepo.MenuItemPage, error) {
	r.itemQuery = query
	return storefrontrepo.MenuItemPage{
		Items: []menuentity.Item{{ID: "item_2", RestaurantID: "rest_2", Name: "Second item"}},
		Total: 3,
	}, nil
}

func TestListRestaurantsForwardsFiltersAndPagination(t *testing.T) {
	repo := &searchRepository{}
	controller := NewController(storefrontsvc.New(repo), nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/restaurants?q=pizza&category=food&cuisine=italian&page=2&pageSize=1&sort=rating", nil)
	res := httptest.NewRecorder()

	controller.ListRestaurants(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	if repo.restaurantQuery.Search != "pizza" || repo.restaurantQuery.CategorySlug != "food" || repo.restaurantQuery.CuisineSlug != "italian" || repo.restaurantQuery.Page != 2 || repo.restaurantQuery.PageSize != 1 || repo.restaurantQuery.Sort != "rating" {
		t.Fatalf("query was not forwarded: %+v", repo.restaurantQuery)
	}
	var body paginatedRestaurantsResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Total != 3 || body.Page != 2 || body.PageSize != 1 || body.TotalPages != 3 || len(body.Data) != 1 {
		t.Fatalf("unexpected page response: %+v", body)
	}
}

func TestListMenuItemsReturnsPaginationMetadata(t *testing.T) {
	repo := &searchRepository{}
	controller := NewController(storefrontsvc.New(repo), nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/menu-items?restaurantId=rest_2&page=1&pageSize=2", nil)
	res := httptest.NewRecorder()

	controller.ListMenuItems(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	var body paginatedItemsResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if repo.itemQuery.RestaurantID != "rest_2" || body.TotalPages != 2 || body.PageSize != 2 || len(body.Data) != 1 {
		t.Fatalf("unexpected page response: %+v query=%+v", body, repo.itemQuery)
	}
}
