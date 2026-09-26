package service

import (
	"context"

	"github.com/ChristianDenniss/go-data-model/serviceability/entity"
	"github.com/ChristianDenniss/go-data-model/serviceability/repository"
)

type Service struct {
	status repository.StoreStatusRepository
	areas  repository.ServiceAreaRepository
}

func New(status repository.StoreStatusRepository, areas repository.ServiceAreaRepository) *Service {
	return &Service{status: status, areas: areas}
}

func (s *Service) RecordStatus(ctx context.Context, st entity.StoreStatus) error {
	if st.SourceStoreID == "" {
		return entity.ErrIDRequired
	}
	return s.status.Upsert(ctx, st)
}
