.PHONY: build test lint fmt clean install-tools mod-tidy

build:
	go build ./...

test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint fmt && golangci-lint run ./...

mod-tidy:
	go mod tidy
	go mod download

