package grpcingestv2

import (
	"context"

	fulfillmententity "github.com/ChristianDenniss/go-data-model/fulfillment/entity"
)

func (s *Server) channelKind(ctx context.Context, channelID string) (string, error) {
	if channelID == "" {
		return "", nil
	}
	ch, err := s.channels.GetByID(ctx, channelID)
	if err != nil {
		return "", err
	}
	return ch.Kind, nil
}

func (s *Server) channelKindForStore(ctx context.Context, sourceStoreID string) (string, error) {
	if sourceStoreID == "" {
		return "", nil
	}
	st, err := s.source.GetStore(ctx, sourceStoreID)
	if err != nil {
		return "", err
	}
	return s.channelKind(ctx, st.ChannelID)
}

func canonicalPath(fulfillmentMode, deliveryExecutor, channelKind string) (string, string) {
	return fulfillmententity.CanonicalizeForChannel(fulfillmentMode, deliveryExecutor, channelKind)
}
