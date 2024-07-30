FROM golang:1.22.4 as builder

COPY ../.. /src
WORKDIR /src

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/user-producer cmd/user-producer/main.go

FROM debian:stable-slim

COPY --from=builder /src/bin/user-producer /app/bin/user-producer

WORKDIR /app

ENTRYPOINT ["/app/bin/user-producer"]