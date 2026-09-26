# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS build
ARG GITHUB_TOKEN
WORKDIR /src
RUN apk add --no-cache git ca-certificates
ENV GOPRIVATE=github.com/ChristianDenniss/*
ENV GONOSUMDB=github.com/ChristianDenniss/*
RUN if [ -n "$GITHUB_TOKEN" ]; then \
  git config --global url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; \
  fi
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download
COPY cmd cmd
COPY internal internal
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api-engine ./cmd/api-engine

FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl
COPY --from=build /out/api-engine /usr/local/bin/api-engine
EXPOSE 8080 9090
HEALTHCHECK --interval=2s --timeout=3s --retries=20 --start-period=5s \
  CMD curl -fsS http://127.0.0.1:8080/health || exit 1
ENTRYPOINT ["api-engine"]
