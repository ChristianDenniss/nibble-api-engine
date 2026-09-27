module github.com/ChristianDenniss/api-engine

go 1.23

require (
	github.com/ChristianDenniss/go-data-model v1.8.1-0.20260927010316-73a593a8abe3
	github.com/ChristianDenniss/go-data-store v1.5.1-0.20260927010320-f75dfbaef8e1
	github.com/ChristianDenniss/platform-contracts v1.6.1-0.20260927010323-c47ac70f1f59
	golang.org/x/crypto v0.31.0
	golang.org/x/oauth2 v0.24.0
	google.golang.org/api v0.214.0
	google.golang.org/grpc v1.68.1
)

// Local workspace modules keep the API, promotion model, and SMS persistence
// changes in lockstep during development.
replace github.com/ChristianDenniss/go-data-model => ../nibble-go-data-model

replace github.com/ChristianDenniss/go-data-store => ../nibble-go-data-store

replace github.com/ChristianDenniss/platform-contracts => ../nibble-platform-contracts

require (
	cloud.google.com/go/auth v0.13.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.6 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.14.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.2 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.54.0 // indirect
	go.opentelemetry.io/otel v1.29.0 // indirect
	go.opentelemetry.io/otel/metric v1.29.0 // indirect
	go.opentelemetry.io/otel/trace v1.29.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241209162323-e6fa225c2576 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)
