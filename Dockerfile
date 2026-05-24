# ── Stage 1: Build ──────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api ./cmd/api/main.go

# ── Stage 2: Run ────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/api        ./api
COPY --from=builder /app/config     ./config
COPY --from=builder /app/migrations ./migrations

EXPOSE 8081

CMD ["./api"]
