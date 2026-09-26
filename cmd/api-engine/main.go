package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ChristianDenniss/api-engine/internal/grpcingest"
	"github.com/ChristianDenniss/api-engine/internal/health"
	"github.com/ChristianDenniss/api-engine/internal/httpx"
	menusvc "github.com/ChristianDenniss/go-data-model/menu/service"
	obssvc "github.com/ChristianDenniss/go-data-model/observation/service"
	offersvc "github.com/ChristianDenniss/go-data-model/offer/service"
	restaurantsvc "github.com/ChristianDenniss/go-data-model/restaurant/service"
	"github.com/ChristianDenniss/go-data-store"
	ingestv1 "github.com/ChristianDenniss/platform-contracts/gen/ingest/v1"
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

	restaurantSvc := restaurantsvc.New(postgres.NewRestaurantRepository(db))
	menuSvc := menusvc.New(postgres.NewMenuRepository(db))
	offerSvc := offersvc.New(postgres.NewOfferRepository(db))
	observationSvc := obssvc.New(postgres.NewObservationRepository(db))

	healthController := health.NewController(health.NewService(db))
	log.Printf("domains: restaurant, menu, offer, observation wired")

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	ingestv1.RegisterIngestServiceServer(grpcServer, grpcingest.NewServer(
		restaurantSvc, menuSvc, offerSvc, observationSvc,
	))
	go func() {
		log.Printf("grpc ingest listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	mux := http.NewServeMux()
	health.Mount(mux, healthController)
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
