include .env.sample
export

BIN_NAME=merlin
UI_PATH := ui
UI_BUILD_PATH := ${UI_PATH}/build
API_PATH=api
API_ALL_PACKAGES := $(shell cd ${API_PATH} && go list ./... | grep -v github.com/gojek/mlp/api/client | grep -v -e mocks -e client)

all: setup init-dep lint test clean build run

# ============================================================
# Initialize dependency recipes
# ============================================================
.PHONY: setup
setup:
	@echo "> Setting up tools ..."
	@test -x ${GOPATH}/bin/golint || go get -u golang.org/x/lint/golint

.PHONY: init-dep
init-dep: init-dep-ui init-dep-api

.PHONY: init-dep-ui
init-dep-ui:
	@echo "> Initializing UI dependencies ..."
	@cd ${UI_PATH} && yarn

.PHONY: init-dep-api
init-dep-api:
	@echo "> Initializing API dependencies ..."
	@cd ${API_PATH} && go mod tidy -v
	@cd ${API_PATH} && go get -v ./...

# ============================================================
# Analyze source code recipes
# ============================================================
.PHONY: lint
lint: lint-ui lint-api

.PHONY: lint-ui
lint-ui:
	@echo "> Linting the UI source code ..."
	@cd ${UI_PATH} && yarn lint

.PHONY: lint-api
lint-api:
	@echo "> Analyzing API source code..."
	@cd ${API_PATH} && golint ${API_ALL_PACKAGES}

# ============================================================
# Testing recipes
# ============================================================
.PHONY: test
test: test-api

.PHONY: test-api
test-api: init-dep-api
	@echo "> API unit testing ..."
	@cd ${API_PATH} && go test -v -race -cover -coverprofile cover.out -tags unit ${API_ALL_PACKAGES}
	@cd ${API_PATH} && go tool cover -func cover.out

.PHONY: it-test-api-local
it-test-api-local: local-db
	@echo "> API integration testing locally ..."
	@cd ${API_PATH} && go test -v -race -short -cover -coverprofile cover.out -tags unit,integration_local ${API_ALL_PACKAGES}
	@cd ${API_PATH} && go tool cover -func cover.out

.PHONY: it-test-api-ci
it-test-api-ci:
	@echo "> API integration testing ..."
	@cd ${API_PATH} && go test -v -race -short -cover -coverprofile cover.out -tags unit,integration ${API_ALL_PACKAGES}
	@cd ${API_PATH} && go tool cover -func cover.out

# ============================================================
# Building recipes
# ============================================================
.PHONY: build
build: build-ui build-api

.PHONY: build-ui
build-ui: clean-ui
	@echo "> Building UI static build ..."
	@cd ${UI_PATH} && npm run build

.PHONY: build-api
build-api: clean-bin
	@echo "> Building API binary ..."
	@cd ${API_PATH} && go build -o ../bin/${BIN_NAME} ./cmd/main.go

# ============================================================
# Run recipe
# ============================================================
.PHONY: run
run:
	@echo "> Running application ..."
	@./bin/${BIN_NAME}

# ============================================================
# Utility recipes
# ============================================================
.PHONY: clean
clean: clean-ui clean-bin

.PHONY: clean-ui
clean-ui:
	@echo "> Cleaning up existing UI static build ..."
	@test ! -e ${UI_BUILD_PATH} || rm -r ${UI_BUILD_PATH}

.PHONY: clean-bin
clean-bin:
	@echo "> Cleaning up existing compiled binary ..."
	@test ! -e bin || rm -r bin

.PHONY: local-db
local-db:
	@echo "> Starting up DB ..."
	@docker-compose up -d postgres && docker-compose run migrations

.PHONY: stop-docker
stop-docker:
	@echo "> Stopping Docker compose ..."
	@docker-compose down

.PHONY: swagger-ui
swagger-ui:
	@echo "Starting Swagger UI"
	@docker-compose up -d swagger-ui

# ============================================================
# Generate client recipes
# ============================================================
.PHONY: generate-client
generate-client: generate-client-go generate-client-python

CLIENT_GO_OUTPUT_DIR = ./api/client
TEMP_CLIENT_GO_OUTPUT_DIR = ./api/client_tmp
CLIENT_GO_EXAMPLES_DIR = ./api/client/examples
TEMP_CLIENT_GO_EXAMPLES_DIR  = ./api/client_examples_temp
.PHONY: generate-client-go
generate-client-go:
	@echo "Generating Go client from swagger.yaml"
	@mv ${CLIENT_GO_EXAMPLES_DIR} ${TEMP_CLIENT_GO_EXAMPLES_DIR}
	@rm -rf ${CLIENT_GO_OUTPUT_DIR}
	@swagger-codegen generate -i swagger.yaml -l go -o ${TEMP_CLIENT_GO_OUTPUT_DIR} -DpackageName=client
	@mkdir ${CLIENT_GO_OUTPUT_DIR}
	@mv ${TEMP_CLIENT_GO_OUTPUT_DIR}/*.go ${CLIENT_GO_OUTPUT_DIR}
	@rm -rf ${TEMP_CLIENT_GO_OUTPUT_DIR}
	@mv ${TEMP_CLIENT_GO_EXAMPLES_DIR} ${CLIENT_GO_EXAMPLES_DIR}
	@goimports -w ${CLIENT_GO_OUTPUT_DIR}

CLIENT_PYTHON_OUTPUT_DIR = ./python/sdk/client
TEMP_CLIENT_PYTHON_OUTPUT_DIR = ./python/sdk/client_tmp
.PHONY: generate-client-python
generate-client-python:
	@echo "Generating Python client from swagger.yaml"
	@rm -rf ${CLIENT_PYTHON_OUTPUT_DIR}
	@swagger-codegen generate -i swagger.yaml -l python -o ${TEMP_CLIENT_PYTHON_OUTPUT_DIR} -DpackageName=client
	@mv ${TEMP_CLIENT_PYTHON_OUTPUT_DIR}/client ${CLIENT_PYTHON_OUTPUT_DIR}
	@rm -rf ${TEMP_CLIENT_PYTHON_OUTPUT_DIR}
