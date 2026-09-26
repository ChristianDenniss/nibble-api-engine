package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/user/entity"
	"github.com/ChristianDenniss/go-data-model/user/repository"
)

type Service struct {
	users       repository.UserRepository
	settings    repository.SettingsRepository
	sessions    repository.SessionRepository
	memberships repository.MembershipRepository
}

func New(
	users repository.UserRepository,
	settings repository.SettingsRepository,
	sessions repository.SessionRepository,
	memberships repository.MembershipRepository,
) *Service {
	return &Service{users: users, settings: settings, sessions: sessions, memberships: memberships}
}

func (s *Service) RecordUser(ctx context.Context, u entity.User) error {
	if u.ID == "" {
		return entity.ErrIDRequired
	}
	return s.users.Upsert(ctx, u)
}

func (s *Service) SaveSettings(ctx context.Context, settings entity.Settings) error {
	if settings.UserID == "" {
		return entity.ErrIDRequired
	}
	return s.settings.Upsert(ctx, settings)
}

func (s *Service) GetSettings(ctx context.Context, userID string) (entity.Settings, error) {
	return s.settings.Get(ctx, userID)
}

func (s *Service) SaveSession(ctx context.Context, sess entity.Session) error {
	if sess.ID == "" {
		return entity.ErrIDRequired
	}
	return s.sessions.Insert(ctx, sess)
}

func (s *Service) GetSession(ctx context.Context, id string) (entity.Session, error) {
	return s.sessions.GetByID(ctx, id)
}

func (s *Service) ListMembershipSlugs(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, entity.ErrIDRequired
	}
	if s.memberships == nil {
		return nil, nil
	}
	return s.memberships.ListProductSlugsByUser(ctx, userID)
}
