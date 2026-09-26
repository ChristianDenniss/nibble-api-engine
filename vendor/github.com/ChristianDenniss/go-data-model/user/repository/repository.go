package repository

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/user/entity"
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (entity.User, error)
	Upsert(ctx context.Context, u entity.User) error
}

type SettingsRepository interface {
	Get(ctx context.Context, userID string) (entity.Settings, error)
	Upsert(ctx context.Context, s entity.Settings) error
}

type SessionRepository interface {
	GetByID(ctx context.Context, id string) (entity.Session, error)
	Insert(ctx context.Context, s entity.Session) error
	UpdateResult(ctx context.Context, id string, result []byte) error
}

type OutboundClickRepository interface {
	Insert(ctx context.Context, c entity.OutboundClick) error
}

type MembershipRepository interface {
	ListProductSlugsByUser(ctx context.Context, userID string) ([]string, error)
}
