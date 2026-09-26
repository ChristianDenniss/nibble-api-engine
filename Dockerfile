FROM golang:1.23-bookworm AS build
ARG GITHUB_TOKEN
WORKDIR /src
RUN apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq git >/dev/null
ENV GOPRIVATE=github.com/ChristianDenniss/*
RUN if [ -n "$GITHUB_TOKEN" ]; then \
  git config --global url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; \
  fi
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api-engine ./cmd/api-engine

FROM debian:bookworm-slim
COPY --from=build /out/api-engine /usr/local/bin/api-engine
EXPOSE 8080 9090
ENTRYPOINT ["api-engine"]
