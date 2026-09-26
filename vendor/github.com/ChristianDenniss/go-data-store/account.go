package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/account/entity"
	"github.com/ChristianDenniss/go-data-model/account/repository"
)

var _ repository.Repository = (*AccountRepository)(nil)

type AccountRepository struct {
	db *DB
}

func NewAccountRepository(db *DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (entity.Account, error) {
	var out entity.Account
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, name, email, phone
		FROM accounts
		WHERE id = $1`, id).Scan(&out.ID, &out.Name, &out.Email, &out.Phone)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Account{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Account{}, err
	}

	addrRows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, label, latitude, longitude, address, city, region, postal_code, current
		FROM saved_addresses
		WHERE account_id = $1`, id)
	if err != nil {
		return entity.Account{}, err
	}
	defer addrRows.Close()
	for addrRows.Next() {
		var address entity.SavedAddress
		if err := addrRows.Scan(
			&address.ID, &address.Label,
			&address.Location.Latitude, &address.Location.Longitude, &address.Location.Address,
			&address.Location.City, &address.Location.Region, &address.Location.PostalCode,
			&address.Current,
		); err != nil {
			return entity.Account{}, err
		}
		out.Addresses = append(out.Addresses, address)
	}
	if err := addrRows.Err(); err != nil {
		return entity.Account{}, err
	}

	payRows, err := r.db.sql.QueryContext(ctx, `
		SELECT id, brand, last4, exp_month, exp_year, is_default
		FROM payment_methods
		WHERE account_id = $1`, id)
	if err != nil {
		return entity.Account{}, err
	}
	defer payRows.Close()
	for payRows.Next() {
		var method entity.PaymentMethod
		if err := payRows.Scan(&method.ID, &method.Brand, &method.Last4, &method.ExpMonth, &method.ExpYear, &method.Default); err != nil {
			return entity.Account{}, err
		}
		out.PaymentMethods = append(out.PaymentMethods, method)
	}
	return out, payRows.Err()
}

func (r *AccountRepository) Upsert(ctx context.Context, in entity.Account) error {
	tx, err := r.db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO accounts (id, name, email, phone)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			email = EXCLUDED.email,
			phone = EXCLUDED.phone,
			updated_at = now()`,
		in.ID, in.Name, in.Email, in.Phone); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM saved_addresses WHERE account_id = $1`, in.ID); err != nil {
		return err
	}
	for _, address := range in.Addresses {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO saved_addresses (id, account_id, label, latitude, longitude, address, city, region, postal_code, current)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			address.ID, in.ID, address.Label,
			address.Location.Latitude, address.Location.Longitude, address.Location.Address,
			address.Location.City, address.Location.Region, address.Location.PostalCode,
			address.Current); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM payment_methods WHERE account_id = $1`, in.ID); err != nil {
		return err
	}
	for _, method := range in.PaymentMethods {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO payment_methods (id, account_id, brand, last4, exp_month, exp_year, is_default)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			method.ID, in.ID, method.Brand, method.Last4, method.ExpMonth, method.ExpYear, method.Default); err != nil {
			return err
		}
	}
	return tx.Commit()
}
