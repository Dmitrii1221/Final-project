FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/budget-api ./cmd/budget-api && \
    go build -o /app/budget-consumer ./cmd/budget-consumer

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/budget-api /usr/local/bin/
COPY --from=builder /app/budget-consumer /usr/local/bin/
COPY --from=builder /app/migrations /migrations
