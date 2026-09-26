package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ChristianDenniss/api-engine/internal/ingest"
	"github.com/ChristianDenniss/api-engine/internal/store"
	ingestv1 "github.com/ChristianDenniss/platform-contracts/gen/ingest/v1"
	"google.golang.org/grpc"
)

func main() {
	databaseURL := getenv("DATABASE_URL", "postgres://doordash:doordash@localhost:5432/doordash?sslmode=disable")
	httpAddr := getenv("HTTP_ADDR", ":8080")
	grpcAddr := getenv("GRPC_ADDR", ":9090")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st, err := store.Open(ctx, databaseURL)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer st.Close()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	ingestv1.RegisterIngestServiceServer(grpcServer, ingest.NewServer(st))
	go func() {
		log.Printf("grpc ingest listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	httpServer := &http.Server{Addr: httpAddr, Handler: mux}
	go func() {
		log.Printf("http listening on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http serve: %v", err)
		}
	}()

	<-ctx.Done()
	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(context.Background())
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
