include ../.env.sample
export

BIN_NAME=merlin-api
ALL_PACKAGES=$(shell go list ./...)
TEST_PACKAGES=$(shell go list ./... | grep -v github.com/gojek/merlin/client | grep -v mocks)

.PHONY: default test it it-local race dep build clean setup coverage coverhtml lint

default: test

clean:
	@test ! -e bin || rm -r bin

run: build local-db
	go run cmd/main.go

stop:
	docker-compose down

dep:
	@echo "Fetching dependencies..."
	@GO111MODULE=on go mod tidy

build: dep
	@echo "GOPATH=${GOPATH}"
	@test ! -e bin || rm -r bin
	@echo "Building binary..."
	@GO111MODULE=on go build -o ./bin/${BIN_NAME} cmd/main.go

local-db:
	@docker-compose up -d db
	@echo "Starting up DB..."
	@sleep 5
	@docker-compose run migrations

lint: ## Lint all go files
	@echo "Analyzing source code..."
	@GO111MODULE=on golint -set_exit_status $(ALL_PACKAGES)

test: dep
	@echo "Testing source code..."
	@GO111MODULE=on go test -race -short -cover -coverprofile cover.out $(TEST_PACKAGES) -tags unit
	@GO111MODULE=on go tool cover -func cover.out

it: local-db dep
	@GO111MODULE=on go test -race -short -cover -coverprofile cover.out $(TEST_PACKAGES) -tags integration_local

it-ci: dep
	@echo "Running integration tests..."
	@GO111MODULE=on go test -race -short -cover -coverprofile cover.out $(TEST_PACKAGES) -tags integration

test-all: local-db dep
	@GO111MODULE=on go test -race -short -cover -coverprofile cover.out $(TEST_PACKAGES) -tags unit,integration_local
	@GO111MODULE=on go tool cover -func cover.out

test-all-ci: dep
	@echo "Running integration tests..."
	@GO111MODULE=on go test -race -short -cover -coverprofile cover.out $(TEST_PACKAGES) -tags unit,integration
	@GO111MODULE=on go tool cover -func cover.out

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
