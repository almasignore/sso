FROM golang:1.26.0-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build -o /app/sso ./cmd/sso

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build -o /app/migrator ./cmd/migrator


FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/sso ./sso
COPY --from=builder /app/migrator ./migrator
COPY config/ ./config/
COPY migrations/ ./migrations/

RUN mkdir -p /app/storage

EXPOSE 44044

ENTRYPOINT ["./sso"]
CMD ["--config=./config/local.yaml"]
