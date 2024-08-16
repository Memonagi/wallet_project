FROM golang:1.22.4 AS builder

COPY ../.. /src

WORKDIR /src

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/wallet-service cmd/wallet-service/main.go

FROM debian:stable-slim

COPY --from=builder /src/bin/wallet-service /app/bin/wallet-service

WORKDIR /app

EXPOSE 8080

ENTRYPOINT ["./bin/wallet-service"]