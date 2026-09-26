package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/order/entity"
	"github.com/ChristianDenniss/go-data-model/order/repository"
)

var _ repository.Repository = (*OrderRepository)(nil)

type OrderRepository struct {
	db *DB
}

func NewOrderRepository(db *DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (entity.Order, error) {
	var out entity.Order
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, account_id, restaurant_id, provider_id, placed_at, status, amount_cents, currency
		FROM orders
		WHERE id = $1`, id).Scan(
		&out.ID, &out.AccountID, &out.RestaurantID, &out.ProviderID,
		&out.PlacedAt, &out.Status, &out.Total.AmountCents, &out.Total.Currency)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Order{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Order{}, err
	}
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, menu_item_id, name, quantity
		FROM order_lines
		WHERE order_id = $1`, id)
	if err != nil {
		return entity.Order{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var line entity.OrderLine
		if err := rows.Scan(&line.ID, &line.MenuItemID, &line.Name, &line.Quantity); err != nil {
			return entity.Order{}, err
		}
		out.Lines = append(out.Lines, line)
	}
	return out, rows.Err()
}

func (r *OrderRepository) Upsert(ctx context.Context, in entity.Order) error {
	tx, err := r.db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO orders (id, account_id, restaurant_id, provider_id, placed_at, status, amount_cents, currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			account_id = EXCLUDED.account_id,
			restaurant_id = EXCLUDED.restaurant_id,
			provider_id = EXCLUDED.provider_id,
			placed_at = EXCLUDED.placed_at,
			status = EXCLUDED.status,
			amount_cents = EXCLUDED.amount_cents,
			currency = EXCLUDED.currency,
			updated_at = now()`,
		in.ID, in.AccountID, in.RestaurantID, in.ProviderID, in.PlacedAt, string(in.Status),
		in.Total.AmountCents, in.Total.Currency); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM order_lines WHERE order_id = $1`, in.ID); err != nil {
		return err
	}
	for _, line := range in.Lines {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO order_lines (id, order_id, menu_item_id, name, quantity)
			VALUES ($1, $2, $3, $4, $5)`,
			line.ID, in.ID, line.MenuItemID, line.Name, line.Quantity); err != nil {
			return err
		}
	}
	return tx.Commit()
}
