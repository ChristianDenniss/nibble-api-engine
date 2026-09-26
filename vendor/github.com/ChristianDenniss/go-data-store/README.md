# nibble-go-data-store

The persistence adapter. Today that is Postgres: connection, migrations, and repository implementations that satisfy `nibble-go-data-model` contracts.

**Why this repo exists:** SQL, row structs, and schema are not domain. They also are not “the API.” Naming this `go-postgres` would freeze the role to one engine. Repos like this are usually called **storage** or **persistence** — the driven adapter behind a repository interface. This one is `nibble-go-data-store` so it sits next to `nibble-go-data-model`.

`nibble-api-engine` opens a store and injects the concrete repos into domain services. It never writes SQL.

If a second engine is added later, it belongs here as another package (`sqlite`, …), not inside the domain module and not as its own network service.
