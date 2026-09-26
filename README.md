# nibble-api-engine

The running application: HTTP, gRPC ingest, health, config, and process wiring.

**Why this repo exists:** something has to listen on a port, open a store, and plug adapters into `nibble-go-data-model` services. That composition is an application, not the domain. Keeping it here means the domain library stays importable and the store stays swappable.

`cmd/api-engine` is the only place that knows about both `nibble-go-data-store` and the gRPC server. Domain packages do not import this repo.
