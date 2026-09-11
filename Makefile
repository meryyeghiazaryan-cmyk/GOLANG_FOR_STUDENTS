GOPATH      ?= $(HOME)/go
PROTO_DIR    = pkg/proto
PROTO_FILE   = $(PROTO_DIR)/location.proto
MGMT_DB      = postgres://postgres:password@localhost:5432/loc_management?sslmode=disable
HIST_DB      = postgres://postgres:password@localhost:5433/loc_history?sslmode=disable

.PHONY: all proto build test test-integration lint docker-up docker-down tidy migrate-up migrate-down

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

## migrate-up: apply pending SQL migrations (local Postgres on 5432/5433)
migrate-up:
	migrate -path migrations/loc-management -database "$(MGMT_DB)" up
	migrate -path migrations/loc-history -database "$(HIST_DB)" up

## migrate-down: roll back the latest migration on both databases
migrate-down:
	migrate -path migrations/loc-management -database "$(MGMT_DB)" down 1
	migrate -path migrations/loc-history -database "$(HIST_DB)" down 1

## tidy: tidy go module dependencies
tidy:
	go mod tidy
