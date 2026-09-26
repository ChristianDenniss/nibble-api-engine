package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/ChristianDenniss/go-data-model/user/entity"
	"github.com/ChristianDenniss/go-data-model/user/repository"
)

var (
	_ repository.UserRepository          = (*UserRepository)(nil)
	_ repository.SettingsRepository      = (*UserSettingsRepository)(nil)
	_ repository.SessionRepository       = (*CompareSessionRepository)(nil)
	_ repository.OutboundClickRepository = (*OutboundClickRepository)(nil)
	_ repository.MembershipRepository    = (*UserMembershipRepository)(nil)
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (entity.User, error) {
	var out entity.User
	err := r.db.sql.QueryRowContext(ctx, `SELECT id, name, email, phone FROM users WHERE id = $1`, id).
		Scan(&out.ID, &out.Name, &out.Email, &out.Phone)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.User{}, entity.ErrNotFound
	}
	return out, err
}

func (r *UserRepository) Upsert(ctx context.Context, u entity.User) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO users (id, name, email, phone) VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, phone = EXCLUDED.phone, updated_at = now()`,
		u.ID, u.Name, u.Email, u.Phone)
	return err
}

type UserSettingsRepository struct {
	db *DB
}

func NewUserSettingsRepository(db *DB) *UserSettingsRepository {
	return &UserSettingsRepository{db: db}
}

func (r *UserSettingsRepository) Get(ctx context.Context, userID string) (entity.Settings, error) {
	var out entity.Settings
	var prefsJSON []byte
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT user_id, preferred_currency, notifications_enabled, compare_prefs
		FROM user_settings WHERE user_id = $1`, userID).
		Scan(&out.UserID, &out.PreferredCurrency, &out.NotificationsEnabled, &prefsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Settings{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Settings{}, err
	}
	if len(prefsJSON) > 0 {
		_ = json.Unmarshal(prefsJSON, &out.ComparePrefs)
	}
	return out, nil
}

func (r *UserSettingsRepository) Upsert(ctx context.Context, s entity.Settings) error {
	prefsJSON, err := json.Marshal(s.ComparePrefs)
	if err != nil {
		return err
	}
	_, err = r.db.sql.ExecContext(ctx, `
		INSERT INTO user_settings (user_id, preferred_currency, notifications_enabled, compare_prefs)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			preferred_currency = EXCLUDED.preferred_currency,
			notifications_enabled = EXCLUDED.notifications_enabled,
			compare_prefs = EXCLUDED.compare_prefs`,
		s.UserID, s.PreferredCurrency, s.NotificationsEnabled, prefsJSON)
	return err
}

type CompareSessionRepository struct {
	db *DB
}

func NewCompareSessionRepository(db *DB) *CompareSessionRepository {
	return &CompareSessionRepository{db: db}
}

func (r *CompareSessionRepository) GetByID(ctx context.Context, id string) (entity.Session, error) {
	var out entity.Session
	var userID sql.NullString
	var result []byte
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, user_id, place_id, query_snapshot, basket_snapshot, result_snapshot
		FROM compare_sessions WHERE id = $1`, id).
		Scan(&out.ID, &userID, &out.PlaceID, &out.QuerySnapshot, &out.BasketSnapshot, &result)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Session{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Session{}, err
	}
	if userID.Valid {
		out.UserID = userID.String
	}
	if len(result) > 0 {
		out.ResultSnapshot = result
	}
	return out, nil
}

func (r *CompareSessionRepository) Insert(ctx context.Context, s entity.Session) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO compare_sessions (id, user_id, place_id, query_snapshot, basket_snapshot, result_snapshot)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		s.ID, nullString(s.UserID), s.PlaceID, s.QuerySnapshot, s.BasketSnapshot, nullableBytes(s.ResultSnapshot))
	return err
}

func (r *CompareSessionRepository) UpdateResult(ctx context.Context, id string, result []byte) error {
	_, err := r.db.sql.ExecContext(ctx, `UPDATE compare_sessions SET result_snapshot = $2 WHERE id = $1`, id, result)
	return err
}

type OutboundClickRepository struct {
	db *DB
}

func NewOutboundClickRepository(db *DB) *OutboundClickRepository {
	return &OutboundClickRepository{db: db}
}

func (r *OutboundClickRepository) Insert(ctx context.Context, c entity.OutboundClick) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO outbound_clicks (id, user_id, compare_session_id, purchase_option_id, action_kind, target_url, clicked_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())`,
		c.ID, nullString(c.UserID), nullString(c.CompareSessionID), c.PurchaseOptionID, c.ActionKind, c.TargetURL)
	return err
}

func nullableBytes(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return b
}

type UserMembershipRepository struct {
	db *DB
}

func NewUserMembershipRepository(db *DB) *UserMembershipRepository {
	return &UserMembershipRepository{db: db}
}

func (r *UserMembershipRepository) ListProductSlugsByUser(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT mp.slug
		FROM user_memberships um
		JOIN membership_products mp ON mp.id = um.membership_product_id
		WHERE um.user_id = $1
		ORDER BY mp.slug`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var slugs []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, err
		}
		slugs = append(slugs, slug)
	}
	return slugs, rows.Err()
}
