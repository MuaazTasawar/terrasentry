# --- Build stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /operator ./cmd/operator

# --- Runtime stage ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /server /app/server
COPY --from=builder /operator /app/operator
COPY migrations ./migrations

EXPOSE 8080

# Default entrypoint runs the API server. The operator is run as a
# separate container (see infra/docker-compose.yml or a K8s Deployment)
# using the same image with entrypoint overridden to /app/operator.
ENTRYPOINT ["/app/server"]