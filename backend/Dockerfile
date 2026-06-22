# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -o bin/worker ./cmd/worker

FROM alpine:3.21 AS api
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/bin/api /app/api
EXPOSE 8080
CMD ["/app/api"]

FROM alpine:3.21 AS worker
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/bin/worker /app/worker
CMD ["/app/worker"]
