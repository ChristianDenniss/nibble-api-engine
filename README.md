# nibble-api-engine

The running application: HTTP, gRPC ingest, health, config, and process wiring.

**Why this repo exists:** something has to listen on a port, open a store, and plug adapters into `nibble-go-data-model` services. That composition is an application, not the domain. Keeping it here means the domain library stays importable and the store stays swappable.

`cmd/api-engine` is the only place that knows about both `nibble-go-data-store` and the gRPC server. Domain packages do not import this repo.

## Provider catalog

`GET /v1/catalog` serves the Postgres-backed restaurant/menu comparison catalog.
The private v2 `RecordSourceSnapshot` RPC accepts the versioned catalog envelope
documented in `nibble-platform-contracts/catalog`. Acquisition owns extraction;
the model owns matching; the store owns persistence. Collector failures do not
remove the last imported catalog. Collection dates and status are never part
of the public response. Use the sibling `nibble-local-dev` workspace build;
standalone vendored CI includes the additions described in `VENDOR-PATCHES.md`.
