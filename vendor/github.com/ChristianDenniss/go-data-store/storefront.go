package postgres

import (
	"context"
	"database/sql"
	"errors"

	categoryentity "github.com/ChristianDenniss/go-data-model/category/entity"
	cartentity "github.com/ChristianDenniss/go-data-model/cart/entity"
	cuisineentity "github.com/ChristianDenniss/go-data-model/cuisine/entity"
	menuentity "github.com/ChristianDenniss/go-data-model/menu/entity"
	offerentity "github.com/ChristianDenniss/go-data-model/offer/entity"
	orderentity "github.com/ChristianDenniss/go-data-model/order/entity"
	providerentity "github.com/ChristianDenniss/go-data-model/provider/entity"
	restaurantentity "github.com/ChristianDenniss/go-data-model/restaurant/entity"
	"github.com/ChristianDenniss/go-data-model/storefront/entity"
	storefrontrepo "github.com/ChristianDenniss/go-data-model/storefront/repository"
)

var _ storefrontrepo.Repository = (*StorefrontRepository)(nil)

type StorefrontRepository struct {
	db *DB
}

func NewStorefrontRepository(db *DB) *StorefrontRepository {
	return &StorefrontRepository{db: db}
}

func (r *StorefrontRepository) LoadCatalog(ctx context.Context, accountID string) (entity.Catalog, error) {
	accountRepo := &AccountRepository{db: r.db}
	account, err := accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return entity.Catalog{}, err
	}

	out := entity.Catalog{Account: account}

	out.Providers, err = r.listProviders(ctx)
	if err != nil {
		return entity.Catalog{}, err
	}
	out.Categories, err = r.listCategories(ctx)
	if err != nil {
		return entity.Catalog{}, err
	}
	out.Cuisines, err = r.listCuisines(ctx)
	if err != nil {
		return entity.Catalog{}, err
	}
	out.Restaurants, err = r.listRestaurants(ctx)
	if err != nil {
		return entity.Catalog{}, err
	}
	out.Items, err = r.listMenuItems(ctx)
	if err != nil {
		return entity.Catalog{}, err
	}
	out.Offers, err = r.listOffers(ctx)
	if err != nil {
		return entity.Catalog{}, err
	}
	out.Cart, err = r.loadCartForAccount(ctx, accountID)
	if err != nil && !errors.Is(err, cartentity.ErrNotFound) {
		return entity.Catalog{}, err
	}
	out.Orders, err = r.listOrdersForAccount(ctx, accountID)
	if err != nil {
		return entity.Catalog{}, err
	}
	return out, nil
}

func (r *StorefrontRepository) listProviders(ctx context.Context) ([]providerentity.Provider, error) {
	rows, err := r.db.sql.QueryContext(ctx, `SELECT id, name FROM providers ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []providerentity.Provider
	for rows.Next() {
		var p providerentity.Provider
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *StorefrontRepository) listCategories(ctx context.Context) ([]categoryentity.Category, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, slug, name, description FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []categoryentity.Category
	for rows.Next() {
		var c categoryentity.Category
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.Description); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *StorefrontRepository) listCuisines(ctx context.Context) ([]cuisineentity.Cuisine, error) {
	rows, err := r.db.sql.QueryContext(ctx, `SELECT id, slug, name FROM cuisines ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []cuisineentity.Cuisine
	for rows.Next() {
		var c cuisineentity.Cuisine
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *StorefrontRepository) listRestaurants(ctx context.Context) ([]restaurantentity.Restaurant, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, name, latitude, longitude, address, city, region, postal_code, rating_average, rating_count
		FROM restaurants ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []restaurantentity.Restaurant
	for rows.Next() {
		var rest restaurantentity.Restaurant
		if err := rows.Scan(
			&rest.ID, &rest.Name,
			&rest.Location.Latitude, &rest.Location.Longitude, &rest.Location.Address,
			&rest.Location.City, &rest.Location.Region, &rest.Location.PostalCode,
			&rest.Rating.Average, &rest.Rating.Count,
		); err != nil {
			return nil, err
		}
		rest.CuisineIDs, err = listIDs(ctx, r.db.sql, `SELECT cuisine_id FROM restaurant_cuisines WHERE restaurant_id = $1`, rest.ID)
		if err != nil {
			return nil, err
		}
		rest.CategoryIDs, err = listIDs(ctx, r.db.sql, `SELECT category_id FROM restaurant_categories WHERE restaurant_id = $1`, rest.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, rest)
	}
	return out, rows.Err()
}

func (r *StorefrontRepository) listMenuItems(ctx context.Context) ([]menuentity.Item, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, restaurant_id, name, description, section FROM menu_items ORDER BY restaurant_id, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []menuentity.Item
	for rows.Next() {
		var item menuentity.Item
		if err := rows.Scan(&item.ID, &item.RestaurantID, &item.Name, &item.Description, &item.Section); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *StorefrontRepository) listOffers(ctx context.Context) ([]offerentity.Offer, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, restaurant_id, provider_id, menu_item_id, amount_cents, currency, estimated_minutes
		FROM offers ORDER BY restaurant_id, provider_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []offerentity.Offer
	for rows.Next() {
		var o offerentity.Offer
		if err := rows.Scan(
			&o.ID, &o.RestaurantID, &o.ProviderID, &o.MenuItemID,
			&o.Price.AmountCents, &o.Price.Currency, &o.EstimatedMinutes,
		); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *StorefrontRepository) loadCartForAccount(ctx context.Context, accountID string) (cartentity.Cart, error) {
	var cartID string
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id FROM carts WHERE account_id = $1 ORDER BY updated_at DESC LIMIT 1`, accountID).Scan(&cartID)
	if errors.Is(err, sql.ErrNoRows) {
		return cartentity.Cart{}, cartentity.ErrNotFound
	}
	if err != nil {
		return cartentity.Cart{}, err
	}
	return (&CartRepository{db: r.db}).GetByID(ctx, cartID)
}

func (r *StorefrontRepository) listOrdersForAccount(ctx context.Context, accountID string) ([]orderentity.Order, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id FROM orders WHERE account_id = $1 ORDER BY placed_at DESC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	orderRepo := &OrderRepository{db: r.db}
	out := make([]orderentity.Order, 0, len(ids))
	for _, id := range ids {
		o, err := orderRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}
