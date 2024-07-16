build:
	 go build -o bin/wallet-service cmd/wallet-service/main.go

tidy:
	go mod tidy

fmt:
	gofumpt -w .
	gci write . --skip-generated -s standard -s default

lint: tidy fmt build
	golangci-lint run

test:
	go test -v ./...