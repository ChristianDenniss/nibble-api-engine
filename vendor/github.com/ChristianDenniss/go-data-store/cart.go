package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/cart/entity"
	"github.com/ChristianDenniss/go-data-model/cart/repository"
)

var _ repository.Repository = (*CartRepository)(nil)

type CartRepository struct {
	db *DB
}

func NewCartRepository(db *DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) GetByID(ctx context.Context, id string) (entity.Cart, error) {
	var out entity.Cart
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, account_id FROM carts WHERE id = $1`, id).Scan(&out.ID, &out.AccountID)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Cart{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Cart{}, err
	}
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, restaurant_id, menu_item_id, provider_id, quantity
		FROM cart_lines
		WHERE cart_id = $1`, id)
	if err != nil {
		return entity.Cart{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var line entity.CartLine
		if err := rows.Scan(&line.ID, &line.RestaurantID, &line.MenuItemID, &line.ProviderID, &line.Quantity); err != nil {
			return entity.Cart{}, err
		}
		out.Lines = append(out.Lines, line)
	}
	return out, rows.Err()
}

func (r *CartRepository) Upsert(ctx context.Context, in entity.Cart) error {
	tx, err := r.db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO carts (id, account_id)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET
			account_id = EXCLUDED.account_id,
			updated_at = now()`,
		in.ID, in.AccountID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_lines WHERE cart_id = $1`, in.ID); err != nil {
		return err
	}
	for _, line := range in.Lines {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO cart_lines (id, cart_id, restaurant_id, menu_item_id, provider_id, quantity)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			line.ID, in.ID, line.RestaurantID, line.MenuItemID, line.ProviderID, line.Quantity); err != nil {
			return err
		}
	}
	return tx.Commit()
}
