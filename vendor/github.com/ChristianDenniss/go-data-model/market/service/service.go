package service

import (
	"context"
	"strings"

	"github.com/ChristianDenniss/go-data-model/market/entity"
	"github.com/ChristianDenniss/go-data-model/market/repository"
)

type Service struct {
	markets  repository.MarketRepository
	dropoffs repository.ProbeDropoffRepository
	coverage repository.CoverageRepository
}

func New(
	markets repository.MarketRepository,
	dropoffs repository.ProbeDropoffRepository,
	coverage repository.CoverageRepository,
) *Service {
	return &Service{markets: markets, dropoffs: dropoffs, coverage: coverage}
}

func (s *Service) RecordMarket(ctx context.Context, m entity.Market) error {
	if m.ID == "" {
		return entity.ErrIDRequired
	}
	if strings.TrimSpace(m.Slug) == "" {
		return entity.ErrSlugRequired
	}
	if m.Status == "" {
		m.Status = entity.StatusPlanned
	}
	return s.markets.Upsert(ctx, m)
}

func (s *Service) RecordProbeDropoff(ctx context.Context, d entity.ProbeDropoff) error {
	if d.ID == "" || d.MarketID == "" {
		return entity.ErrIDRequired
	}
	return s.dropoffs.Upsert(ctx, d)
}

func (s *Service) RecordCoverage(ctx context.Context, c entity.ChannelCoverage) error {
	if c.ID == "" || c.ChannelID == "" || c.MarketID == "" {
		return entity.ErrIDRequired
	}
	if !validCoverage(c.Status) {
		return entity.ErrStatusInvalid
	}
	return s.coverage.Upsert(ctx, c)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (entity.Market, error) {
	if strings.TrimSpace(slug) == "" {
		return entity.Market{}, entity.ErrSlugRequired
	}
	return s.markets.GetBySlug(ctx, slug)
}

func (s *Service) ListProbeDropoffs(ctx context.Context, marketID string) ([]entity.ProbeDropoff, error) {
	if marketID == "" {
		return nil, entity.ErrIDRequired
	}
	return s.dropoffs.ListByMarket(ctx, marketID)
}

func (s *Service) ListCoverage(ctx context.Context, marketID string) ([]entity.ChannelCoverage, error) {
	if marketID == "" {
		return nil, entity.ErrIDRequired
	}
	return s.coverage.ListByMarket(ctx, marketID)
}

func validCoverage(status string) bool {
	switch status {
	case entity.CoverageExpected, entity.CoverageObserved, entity.CoverageAbsent, entity.CoverageUnknown:
		return true
	default:
		return false
	}
}
