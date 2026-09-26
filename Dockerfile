FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY go-data-model ./go-data-model
COPY platform-contracts ./platform-contracts
COPY api-engine ./api-engine
RUN printf 'go 1.23\n\nuse (\n\t./go-data-model\n\t./platform-contracts\n\t./api-engine\n)\n' > go.work
WORKDIR /src/api-engine
RUN GOWORK=/src/go.work go mod tidy && CGO_ENABLED=0 GOWORK=/src/go.work go build -o /out/api-engine ./cmd/api-engine

FROM debian:bookworm-slim
COPY --from=build /out/api-engine /usr/local/bin/api-engine
EXPOSE 8080 9090
ENTRYPOINT ["api-engine"]
