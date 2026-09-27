package grpcingestv2

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ChristianDenniss/go-data-model/catalog"
	"log"
	"time"

	brandentity "github.com/ChristianDenniss/go-data-model/brand/entity"
	brandsvc "github.com/ChristianDenniss/go-data-model/brand/service"
	channelentity "github.com/ChristianDenniss/go-data-model/channel/entity"
	channelsvc "github.com/ChristianDenniss/go-data-model/channel/service"
	dishentity "github.com/ChristianDenniss/go-data-model/dish/entity"
	dishsvc "github.com/ChristianDenniss/go-data-model/dish/service"
	fulfillmententity "github.com/ChristianDenniss/go-data-model/fulfillment/entity"
	ingestentity "github.com/ChristianDenniss/go-data-model/ingest/entity"
	ingestsvc "github.com/ChristianDenniss/go-data-model/ingest/service"
	itempriceentity "github.com/ChristianDenniss/go-data-model/itemprice/entity"
	itempricesvc "github.com/ChristianDenniss/go-data-model/itemprice/service"
	location "github.com/ChristianDenniss/go-data-model/location/entity"
	marketentity "github.com/ChristianDenniss/go-data-model/market/entity"
	marketsvc "github.com/ChristianDenniss/go-data-model/market/service"
	mediaentity "github.com/ChristianDenniss/go-data-model/media/entity"
	money "github.com/ChristianDenniss/go-data-model/money/entity"
	placeentity "github.com/ChristianDenniss/go-data-model/place/entity"
	placesvc "github.com/ChristianDenniss/go-data-model/place/service"
	promotionentity "github.com/ChristianDenniss/go-data-model/promotion/entity"
	promotionsvc "github.com/ChristianDenniss/go-data-model/promotion/service"
	quoteentity "github.com/ChristianDenniss/go-data-model/quoteobs/entity"
	quoteobssvc "github.com/ChristianDenniss/go-data-model/quoteobs/service"
	resolutionentity "github.com/ChristianDenniss/go-data-model/resolution/entity"
	resolutionsvc "github.com/ChristianDenniss/go-data-model/resolution/service"
	sourceentity "github.com/ChristianDenniss/go-data-model/source/entity"
	sourcesvc "github.com/ChristianDenniss/go-data-model/source/service"
	ingestv2 "github.com/ChristianDenniss/platform-contracts/gen/ingest/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	ingestv2.UnimplementedIngestServiceServer
	channels   *channelsvc.Service
	ingest     *ingestsvc.Service
	source     *sourcesvc.Service
	brands     *brandsvc.Service
	dishes     *dishsvc.Service
	places     *placesvc.Service
	resolution *resolutionsvc.Service
	itemPrices *itempricesvc.Service
	quotes     *quoteobssvc.Service
	catalog    *catalog.Service
	markets    *marketsvc.Service
	promos     *promotionsvc.Service
}

func NewServer(
	channels *channelsvc.Service,
	ingest *ingestsvc.Service,
	source *sourcesvc.Service,
	brands *brandsvc.Service,
	dishes *dishsvc.Service,
	places *placesvc.Service,
	resolution *resolutionsvc.Service,
	itemPrices *itempricesvc.Service,
	quotes *quoteobssvc.Service,
	markets *marketsvc.Service,
	promos *promotionsvc.Service,
	catalogService ...*catalog.Service,
) *Server {
	var catalogSvc *catalog.Service
	if len(catalogService) > 0 {
		catalogSvc = catalogService[0]
	}
	return &Server{
		catalog:    catalogSvc,
		channels:   channels,
		ingest:     ingest,
		source:     source,
		brands:     brands,
		dishes:     dishes,
		places:     places,
		resolution: resolution,
		itemPrices: itemPrices,
		quotes:     quotes,
		markets:    markets,
		promos:     promos,
	}
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, channelentity.ErrIDRequired) ||
		errors.Is(err, ingestentity.ErrIDRequired) ||
		errors.Is(err, sourceentity.ErrIDRequired) ||
		errors.Is(err, itempriceentity.ErrIDRequired) ||
		errors.Is(err, quoteentity.ErrIDRequired) ||
		errors.Is(err, placeentity.ErrIDRequired) ||
		errors.Is(err, resolutionentity.ErrIDRequired) ||
		errors.Is(err, brandentity.ErrIDRequired) ||
		errors.Is(err, dishentity.ErrIDRequired) ||
		errors.Is(err, marketentity.ErrIDRequired) ||
		errors.Is(err, marketentity.ErrSlugRequired) ||
		errors.Is(err, marketentity.ErrStatusInvalid) ||
		errors.Is(err, promotionentity.ErrIDRequired) ||
		errors.Is(err, mediaentity.ErrImageURLInvalid) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.Is(err, channelentity.ErrNotFound) ||
		errors.Is(err, ingestentity.ErrNotFound) ||
		errors.Is(err, sourceentity.ErrNotFound) ||
		errors.Is(err, itempriceentity.ErrNotFound) ||
		errors.Is(err, quoteentity.ErrNotFound) ||
		errors.Is(err, placeentity.ErrNotFound) ||
		errors.Is(err, resolutionentity.ErrNotFound) ||
		errors.Is(err, brandentity.ErrNotFound) ||
		errors.Is(err, dishentity.ErrNotFound) ||
		errors.Is(err, marketentity.ErrNotFound) ||
		errors.Is(err, promotionentity.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Errorf(codes.Internal, "%v", err)
}

func (s *Server) RecordChannel(ctx context.Context, req *ingestv2.RecordChannelRequest) (*ingestv2.RecordChannelResponse, error) {
	ch := req.GetChannel()
	if ch == nil {
		return nil, mapError(channelentity.ErrIDRequired)
	}
	err := s.channels.Record(ctx, channelentity.Channel{
		ID: ch.GetId(), Slug: ch.GetSlug(), Kind: ch.GetKind(), Name: ch.GetName(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordChannelResponse{Id: ch.GetId()}, nil
}

func (s *Server) RecordIngestRun(ctx context.Context, req *ingestv2.RecordIngestRunRequest) (*ingestv2.RecordIngestRunResponse, error) {
	run := req.GetRun()
	if run == nil {
		return nil, mapError(ingestentity.ErrIDRequired)
	}
	var finished *time.Time
	if run.GetFinishedAtUnixMs() != 0 {
		t := time.UnixMilli(run.GetFinishedAtUnixMs())
		finished = &t
	}
	err := s.ingest.StartRun(ctx, ingestentity.IngestRun{
		ID:            run.GetId(),
		JobType:       run.GetJobType(),
		ChannelID:     run.GetChannelId(),
		StartedAt:     time.UnixMilli(run.GetStartedAtUnixMs()),
		FinishedAt:    finished,
		ParserVersion: run.GetParserVersion(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordIngestRunResponse{Id: run.GetId()}, nil
}

func (s *Server) RecordSourceSnapshot(ctx context.Context, req *ingestv2.RecordSourceSnapshotRequest) (*ingestv2.RecordSourceSnapshotResponse, error) {
	snap := req.GetSnapshot()
	if snap == nil {
		return nil, mapError(ingestentity.ErrIDRequired)
	}
	if snap.GetContentType() == catalog.ContentType {
		if s.catalog == nil {
			return nil, status.Error(codes.Unavailable, "catalog ingestion unavailable")
		}
		if len(snap.GetRawJson()) > 16<<20 {
			return nil, status.Error(codes.ResourceExhausted, "catalog too large")
		}
		var bundle catalog.Bundle
		if err := json.Unmarshal(snap.GetRawJson(), &bundle); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid catalog JSON")
		}
		if err := bundle.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if err := s.catalog.Import(ctx, snap.GetRawJson()); err != nil {
			log.Printf("catalog import: %v", err)
			return nil, status.Error(codes.Internal, "catalog import failed")
		}
		return &ingestv2.RecordSourceSnapshotResponse{Id: snap.GetId()}, nil
	}
	err := s.ingest.RecordSnapshot(ctx, ingestentity.SourceSnapshot{
		ID:              snap.GetId(),
		IngestRunID:     snap.GetIngestRunId(),
		ExternalStoreID: snap.GetExternalStoreId(),
		ContentType:     snap.GetContentType(),
		RawJSON:         json.RawMessage(snap.GetRawJson()),
		Checksum:        snap.GetChecksum(),
		ObservedAt:      time.UnixMilli(snap.GetObservedAtUnixMs()),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordSourceSnapshotResponse{Id: snap.GetId()}, nil
}

func (s *Server) RecordSourceStore(ctx context.Context, req *ingestv2.RecordSourceStoreRequest) (*ingestv2.RecordSourceStoreResponse, error) {
	st := req.GetStore()
	if st == nil {
		return nil, mapError(sourceentity.ErrIDRequired)
	}
	loc := st.GetLocation()
	err := s.source.RecordStore(ctx, sourceentity.Store{
		ID: st.GetId(), ChannelID: st.GetChannelId(), ExternalStoreID: st.GetExternalStoreId(),
		Name: st.GetName(), Phone: st.GetPhone(),
		Location: location.Location{
			Latitude: loc.GetLatitude(), Longitude: loc.GetLongitude(), Address: loc.GetAddress(),
			City: loc.GetCity(), Region: loc.GetRegion(), PostalCode: loc.GetPostalCode(),
		},
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordSourceStoreResponse{Id: st.GetId()}, nil
}

func (s *Server) RecordSourceMenu(ctx context.Context, req *ingestv2.RecordSourceMenuRequest) (*ingestv2.RecordSourceMenuResponse, error) {
	m := req.GetMenu()
	if m == nil {
		return nil, mapError(sourceentity.ErrIDRequired)
	}
	kind, err := s.channelKindForStore(ctx, m.GetSourceStoreId())
	if err != nil {
		return nil, mapError(err)
	}
	f, e := canonicalPath(m.GetFulfillmentMode(), m.GetDeliveryExecutor(), kind)
	err = s.source.RecordMenu(ctx, sourceentity.Menu{
		ID: m.GetId(), SourceStoreID: m.GetSourceStoreId(),
		FulfillmentMode: f, DeliveryExecutor: e, ExternalMenuID: m.GetExternalMenuId(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordSourceMenuResponse{Id: m.GetId()}, nil
}

func (s *Server) RecordSourceCategory(ctx context.Context, req *ingestv2.RecordSourceCategoryRequest) (*ingestv2.RecordSourceCategoryResponse, error) {
	c := req.GetCategory()
	if c == nil {
		return nil, mapError(sourceentity.ErrIDRequired)
	}
	err := s.source.RecordCategory(ctx, sourceentity.Category{
		ID: c.GetId(), SourceMenuID: c.GetSourceMenuId(),
		ExternalCategoryID: c.GetExternalCategoryId(), Name: c.GetName(), SortOrder: int(c.GetSortOrder()),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordSourceCategoryResponse{Id: c.GetId()}, nil
}

func (s *Server) RecordSourceItem(ctx context.Context, req *ingestv2.RecordSourceItemRequest) (*ingestv2.RecordSourceItemResponse, error) {
	it := req.GetItem()
	if it == nil {
		return nil, mapError(sourceentity.ErrIDRequired)
	}
	err := s.source.RecordItem(ctx, sourceentity.Item{
		ID: it.GetId(), SourceCategoryID: it.GetSourceCategoryId(),
		ExternalItemID: it.GetExternalItemId(), Name: it.GetName(),
		Description: it.GetDescription(), Available: it.GetAvailable(),
		ImageURL: it.GetImageUrl(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordSourceItemResponse{Id: it.GetId()}, nil
}

func (s *Server) RecordItemPriceObservation(ctx context.Context, req *ingestv2.RecordItemPriceObservationRequest) (*ingestv2.RecordItemPriceObservationResponse, error) {
	obs := req.GetObservation()
	if obs == nil {
		return nil, mapError(itempriceentity.ErrIDRequired)
	}
	price := obs.GetPrice()
	f, e := fulfillmententity.Canonicalize(obs.GetFulfillmentMode(), obs.GetDeliveryExecutor())
	err := s.itemPrices.Record(ctx, itempriceentity.Observation{
		ID: obs.GetId(), SourceItemID: obs.GetSourceItemId(), IngestRunID: obs.GetIngestRunId(),
		Price:           money.Money{AmountCents: price.GetAmountCents(), Currency: price.GetCurrency()},
		FulfillmentMode: f, DeliveryExecutor: e,
		ObservedAt: time.UnixMilli(obs.GetObservedAtUnixMs()),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordItemPriceObservationResponse{Id: obs.GetId()}, nil
}

func (s *Server) RecordQuoteObservation(ctx context.Context, req *ingestv2.RecordQuoteObservationRequest) (*ingestv2.RecordQuoteObservationResponse, error) {
	obs := req.GetObservation()
	if obs == nil {
		return nil, mapError(quoteentity.ErrIDRequired)
	}
	feeLines := make([]quoteentity.FeeLine, 0, len(obs.GetFeeLines()))
	for _, fl := range obs.GetFeeLines() {
		amt := fl.GetAmount()
		feeLines = append(feeLines, quoteentity.FeeLine{
			ID: fl.GetId(), Kind: fl.GetKind(),
			Amount:  money.Money{AmountCents: amt.GetAmountCents(), Currency: amt.GetCurrency()},
			Percent: fl.GetPercent(), ThresholdCents: fl.GetThresholdCents(),
		})
	}
	kind, err := s.channelKind(ctx, obs.GetChannelId())
	if err != nil {
		return nil, mapError(err)
	}
	qf, qe := canonicalPath(obs.GetFulfillmentMode(), obs.GetDeliveryExecutor(), kind)
	err = s.quotes.Record(ctx, quoteentity.Observation{
		ID: obs.GetId(), SourceStoreID: obs.GetSourceStoreId(), ChannelID: obs.GetChannelId(),
		FulfillmentMode: qf, DeliveryExecutor: qe, DropoffGeohash: obs.GetDropoffGeohash(),
		MembershipTier: obs.GetMembershipTier(), QuoteKind: obs.GetQuoteKind(),
		BasketSubtotalCents: obs.GetBasketSubtotalCents(),
		ObservedAt:          time.UnixMilli(obs.GetObservedAtUnixMs()),
		IngestRunID:         obs.GetIngestRunId(),
		FeeLines:            feeLines,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordQuoteObservationResponse{Id: obs.GetId()}, nil
}

func (s *Server) RecordPlace(ctx context.Context, req *ingestv2.RecordPlaceRequest) (*ingestv2.RecordPlaceResponse, error) {
	p := req.GetPlace()
	if p == nil {
		return nil, mapError(placeentity.ErrIDRequired)
	}
	loc := p.GetLocation()
	err := s.places.RecordPlace(ctx, placeentity.Place{
		ID: p.GetId(), BrandID: p.GetBrandId(), Name: p.GetName(),
		Location: location.Location{
			Latitude: loc.GetLatitude(), Longitude: loc.GetLongitude(), Address: loc.GetAddress(),
			City: loc.GetCity(), Region: loc.GetRegion(), PostalCode: loc.GetPostalCode(),
		},
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordPlaceResponse{Id: p.GetId()}, nil
}

func (s *Server) RecordPurchaseOption(ctx context.Context, req *ingestv2.RecordPurchaseOptionRequest) (*ingestv2.RecordPurchaseOptionResponse, error) {
	opt := req.GetOption()
	if opt == nil {
		return nil, mapError(placeentity.ErrIDRequired)
	}
	kind, err := s.channelKind(ctx, opt.GetChannelId())
	if err != nil {
		return nil, mapError(err)
	}
	pf, pe := canonicalPath(opt.GetFulfillmentMode(), opt.GetDeliveryExecutor(), kind)
	err = s.places.RecordPurchaseOption(ctx, placeentity.PurchaseOption{
		ID: opt.GetId(), PlaceID: opt.GetPlaceId(), ChannelID: opt.GetChannelId(),
		FulfillmentMode: pf, DeliveryExecutor: pe, SourceStoreID: opt.GetSourceStoreId(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordPurchaseOptionResponse{Id: opt.GetId()}, nil
}

func (s *Server) RecordStoreMatch(ctx context.Context, req *ingestv2.RecordStoreMatchRequest) (*ingestv2.RecordStoreMatchResponse, error) {
	m := req.GetMatch()
	if m == nil {
		return nil, mapError(resolutionentity.ErrIDRequired)
	}
	err := s.resolution.RecordStoreMatch(ctx, resolutionentity.StoreMatch{
		ID: m.GetId(), SourceStoreID: m.GetSourceStoreId(), PlaceID: m.GetPlaceId(),
		Confidence: m.GetConfidence(), Status: m.GetStatus(), Method: m.GetMethod(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordStoreMatchResponse{Id: m.GetId()}, nil
}

func (s *Server) RecordItemMatch(ctx context.Context, req *ingestv2.RecordItemMatchRequest) (*ingestv2.RecordItemMatchResponse, error) {
	m := req.GetMatch()
	if m == nil {
		return nil, mapError(resolutionentity.ErrIDRequired)
	}
	err := s.resolution.RecordItemMatch(ctx, resolutionentity.ItemMatch{
		ID: m.GetId(), SourceItemID: m.GetSourceItemId(), DishID: m.GetDishId(),
		Confidence: m.GetConfidence(), Status: m.GetStatus(), Method: m.GetMethod(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordItemMatchResponse{Id: m.GetId()}, nil
}

func (s *Server) RecordBrand(ctx context.Context, req *ingestv2.RecordBrandRequest) (*ingestv2.RecordBrandResponse, error) {
	b := req.GetBrand()
	if b == nil {
		return nil, mapError(brandentity.ErrIDRequired)
	}
	err := s.brands.Record(ctx, brandentity.Brand{ID: b.GetId(), Slug: b.GetSlug(), Name: b.GetName()})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordBrandResponse{Id: b.GetId()}, nil
}

func (s *Server) RecordDish(ctx context.Context, req *ingestv2.RecordDishRequest) (*ingestv2.RecordDishResponse, error) {
	d := req.GetDish()
	if d == nil {
		return nil, mapError(dishentity.ErrIDRequired)
	}
	err := s.dishes.Record(ctx, dishentity.Dish{
		ID: d.GetId(), BrandID: d.GetBrandId(), Name: d.GetName(), Description: d.GetDescription(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordDishResponse{Id: d.GetId()}, nil
}

func (s *Server) RecordMarket(ctx context.Context, req *ingestv2.RecordMarketRequest) (*ingestv2.RecordMarketResponse, error) {
	m := req.GetMarket()
	if m == nil {
		return nil, mapError(marketentity.ErrIDRequired)
	}
	err := s.markets.RecordMarket(ctx, marketentity.Market{
		ID: m.GetId(), Slug: m.GetSlug(), Name: m.GetName(), Country: m.GetCountry(),
		Region: m.GetRegion(), Currency: m.GetCurrency(), Timezone: m.GetTimezone(),
		Status: m.GetStatus(), GeohashPrefixes: m.GetGeohashPrefixes(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordMarketResponse{Id: m.GetId()}, nil
}

func (s *Server) RecordProbeDropoff(ctx context.Context, req *ingestv2.RecordProbeDropoffRequest) (*ingestv2.RecordProbeDropoffResponse, error) {
	d := req.GetDropoff()
	if d == nil {
		return nil, mapError(marketentity.ErrIDRequired)
	}
	loc := d.GetLocation()
	err := s.markets.RecordProbeDropoff(ctx, marketentity.ProbeDropoff{
		ID: d.GetId(), MarketID: d.GetMarketId(), Label: d.GetLabel(), Geohash: d.GetGeohash(),
		Location: location.Location{
			Latitude: loc.GetLatitude(), Longitude: loc.GetLongitude(), Address: loc.GetAddress(),
			City: loc.GetCity(), Region: loc.GetRegion(), PostalCode: loc.GetPostalCode(),
		},
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordProbeDropoffResponse{Id: d.GetId()}, nil
}

func (s *Server) RecordChannelMarketCoverage(ctx context.Context, req *ingestv2.RecordChannelMarketCoverageRequest) (*ingestv2.RecordChannelMarketCoverageResponse, error) {
	c := req.GetCoverage()
	if c == nil {
		return nil, mapError(marketentity.ErrIDRequired)
	}
	var observed *time.Time
	if c.GetLastObservedAtUnixMs() != 0 {
		t := time.UnixMilli(c.GetLastObservedAtUnixMs())
		observed = &t
	}
	err := s.markets.RecordCoverage(ctx, marketentity.ChannelCoverage{
		ID: c.GetId(), ChannelID: c.GetChannelId(), MarketID: c.GetMarketId(),
		Status: c.GetStatus(), StoreCount: int(c.GetStoreCount()),
		IngestRunID: c.GetIngestRunId(), Note: c.GetNote(), LastObservedAt: observed,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordChannelMarketCoverageResponse{Id: c.GetId()}, nil
}

func (s *Server) RecordMembershipProduct(ctx context.Context, req *ingestv2.RecordMembershipProductRequest) (*ingestv2.RecordMembershipProductResponse, error) {
	p := req.GetProduct()
	if p == nil {
		return nil, mapError(promotionentity.ErrIDRequired)
	}
	err := s.promos.RecordMembershipProduct(ctx, promotionentity.MembershipProduct{
		ID: p.GetId(), ChannelID: p.GetChannelId(), Name: p.GetName(), Slug: p.GetSlug(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &ingestv2.RecordMembershipProductResponse{Id: p.GetId()}, nil
}
