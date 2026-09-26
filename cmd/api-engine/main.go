package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ChristianDenniss/api-engine/internal/compare"
	"github.com/ChristianDenniss/api-engine/internal/grpcingest"
	"github.com/ChristianDenniss/api-engine/internal/storefront"
	grpcingestv2 "github.com/ChristianDenniss/api-engine/internal/grpcingest/v2"
	"github.com/ChristianDenniss/api-engine/internal/health"
	"github.com/ChristianDenniss/api-engine/internal/httpx"
	brandsvc "github.com/ChristianDenniss/go-data-model/brand/service"
	channelsvc "github.com/ChristianDenniss/go-data-model/channel/service"
	comparesvc "github.com/ChristianDenniss/go-data-model/compare/service"
	dishsvc "github.com/ChristianDenniss/go-data-model/dish/service"
	ingestsvc "github.com/ChristianDenniss/go-data-model/ingest/service"
	itempricesvc "github.com/ChristianDenniss/go-data-model/itemprice/service"
	menusvc "github.com/ChristianDenniss/go-data-model/menu/service"
	obssvc "github.com/ChristianDenniss/go-data-model/observation/service"
	offersvc "github.com/ChristianDenniss/go-data-model/offer/service"
	placesvc "github.com/ChristianDenniss/go-data-model/place/service"
	promotionsvc "github.com/ChristianDenniss/go-data-model/promotion/service"
	quoteobssvc "github.com/ChristianDenniss/go-data-model/quoteobs/service"
	resolutionsvc "github.com/ChristianDenniss/go-data-model/resolution/service"
	restaurantsvc "github.com/ChristianDenniss/go-data-model/restaurant/service"
	storefrontsvc "github.com/ChristianDenniss/go-data-model/storefront/service"
	sourcesvc "github.com/ChristianDenniss/go-data-model/source/service"
	serviceabilitysvc "github.com/ChristianDenniss/go-data-model/serviceability/service"
	usersvc "github.com/ChristianDenniss/go-data-model/user/service"
	"github.com/ChristianDenniss/go-data-store"
	ingestv1 "github.com/ChristianDenniss/platform-contracts/gen/ingest/v1"
	ingestv2 "github.com/ChristianDenniss/platform-contracts/gen/ingest/v2"
	"google.golang.org/grpc"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	databaseURL := getenv("DATABASE_URL", "postgres://doordash:doordash@localhost:5432/doordash?sslmode=disable")
	httpAddr := getenv("HTTP_ADDR", ":8080")
	grpcAddr := getenv("GRPC_ADDR", ":9090")

	log.Printf("api-engine starting HTTP_ADDR=%s GRPC_ADDR=%s", httpAddr, grpcAddr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer db.Close()
	log.Printf("postgres: ready")

	// Legacy ingest (v1)
	restaurantSvc := restaurantsvc.New(postgres.NewRestaurantRepository(db))
	menuSvc := menusvc.New(postgres.NewMenuRepository(db))
	offerSvc := offersvc.New(postgres.NewOfferRepository(db))
	observationSvc := obssvc.New(postgres.NewObservationRepository(db))

	// Target domain
	channelSvc := channelsvc.New(postgres.NewChannelRepository(db))
	ingestDomainSvc := ingestsvc.New(postgres.NewIngestRunRepository(db), postgres.NewSourceSnapshotRepository(db))
	sourceSvc := sourcesvc.New(
		postgres.NewSourceStoreRepository(db),
		postgres.NewSourceMenuRepository(db),
		postgres.NewSourceCategoryRepository(db),
		postgres.NewSourceItemRepository(db),
	)
	brandSvc := brandsvc.New(postgres.NewBrandRepository(db))
	placeSvc := placesvc.New(postgres.NewPlaceRepository(db), postgres.NewPurchaseOptionRepository(db))
	dishSvc := dishsvc.New(postgres.NewDishRepository(db))
	resolutionSvc := resolutionsvc.New(
		postgres.NewStoreMatchRepository(db),
		postgres.NewItemMatchRepository(db),
		postgres.NewMatchEvidenceRepository(db),
	)
	itemPriceSvc := itempricesvc.New(postgres.NewItemPriceObservationRepository(db))
	quoteObsSvc := quoteobssvc.New(postgres.NewQuoteObservationRepository(db))
	_ = promotionsvc.New(postgres.NewPromotionRepository(db), postgres.NewMembershipProductRepository(db))
	serviceabilitySvc := serviceabilitysvc.New(
		postgres.NewSourceStoreStatusRepository(db),
		postgres.NewServiceAreaRepository(db),
	)
	userSvc := usersvc.New(
		postgres.NewUserRepository(db),
		postgres.NewUserSettingsRepository(db),
		postgres.NewCompareSessionRepository(db),
		postgres.NewUserMembershipRepository(db),
	)
	compareSvc := comparesvc.New(placeSvc, userSvc, channelSvc, resolutionSvc, itemPriceSvc, quoteObsSvc, serviceabilitySvc)

	healthController := health.NewController(health.NewService(db))
	compareController := compare.NewController(compareSvc, userSvc)
	storefrontController := storefront.NewController(storefrontsvc.New(postgres.NewStorefrontRepository(db)))
	log.Printf("domains: legacy ingest + target catalog/pricing/compare wired")

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	ingestv1.RegisterIngestServiceServer(grpcServer, grpcingest.NewServer(
		restaurantSvc, menuSvc, offerSvc, observationSvc,
	))
	ingestv2.RegisterIngestServiceServer(grpcServer, grpcingestv2.NewServer(
		channelSvc, ingestDomainSvc, sourceSvc, brandSvc, dishSvc, placeSvc, resolutionSvc, itemPriceSvc, quoteObsSvc,
	))
	go func() {
		log.Printf("grpc ingest listening on %s (v1 + v2)", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	mux := http.NewServeMux()
	health.Mount(mux, healthController)
	compare.Mount(mux, compareController)
	storefront.Mount(mux, storefrontController)
	httpServer := &http.Server{Addr: httpAddr, Handler: httpx.Wrap(mux)}
	go func() {
		log.Printf("http listening on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("api-engine shutting down")
	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(context.Background())
	log.Printf("api-engine stopped")
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
