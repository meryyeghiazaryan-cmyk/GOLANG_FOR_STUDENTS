GOPATH      ?= $(HOME)/go
PROTO_DIR    = pkg/proto
PROTO_FILE   = $(PROTO_DIR)/location.proto

.PHONY: all proto build test test-integration lint docker-up docker-down tidy

all: proto build

## proto: regenerate Go code from the .proto definition
proto:
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_FILE)

## build: compile both services
build:
	go build ./cmd/loc-management
	go build ./cmd/loc-history

## test: run unit and functional tests
test:
	go test -race -count=1 ./...

## test-integration: run integration tests (requires running services and DBs)
test-integration:
	go test -race -tags integration ./test/integration/...

## run-management: start the location management service locally
run-management:
	go run ./cmd/loc-management

## run-history: start the location history service locally
run-history:
	go run ./cmd/loc-history

## docker-up: build and start all services with Docker Compose
docker-up:
	docker-compose up --build

## docker-down: tear down all Docker Compose resources
docker-down:
	docker-compose down -v

## tidy: tidy go module dependencies
tidy:
	go mod tidy
