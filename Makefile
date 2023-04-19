include .env.sample
export

BIN_NAME=merlin
TRANSFORMER_BIN_NAME=merlin-transformer
INFERENCE_LOGGER_BIN_NAME=merlin-logger
UI_PATH := ui
UI_BUILD_PATH := ${UI_PATH}/build
API_PATH=api
API_ALL_PACKAGES := $(shell cd ${API_PATH} && go list ./... | grep -v github.com/caraml-dev/mlp/api/client | grep -v -e mocks -e client)
VERSION := $(or ${VERSION}, $(shell git describe --tags --always --first-parent))
LOG_URL?=localhost:8002
TEST_TAGS?=

GOLANGCI_LINT_VERSION="v1.51.2"
PROTOC_GEN_GO_JSON_VERSION="v1.1.0"
PROTOC_GEN_GO_VERSION="v1.26"
PYTHON_VERSION ?= "37"	#set as 37 38 39 310 for 3.7-3.10 respectively

all: setup init-dep lint test clean build run

# ============================================================
# Initialize dependency recipes
# ============================================================
.PHONY: setup
setup:
	@echo "> Setting up tools ..."
	@test -x ${GOPATH}/bin/goimports || go install golang.org/x/tools/cmd/goimports@latest
	@test -x ${GOPATH}/bin/gotest || go install github.com/rakyll/gotest@latest
	@test -x ${GOPATH}/bin/protoc-gen-go-json || go install github.com/mitchellh/protoc-gen-go-json@${PROTOC_GEN_GO_JSON_VERSION}
	@test -x ${GOPATH}/bin/protoc-gen-go || go install google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}
	@command -v golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin ${GOLANGCI_LINT_VERSION}

.PHONY: init-dep
init-dep: init-dep-ui init-dep-api

.PHONY: init-dep-ui
init-dep-ui:
	@echo "> Initializing UI dependencies ..."
	@cd ${UI_PATH} && yarn install --network-concurrency 1

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
	@cd ${API_PATH} && golangci-lint run

# ============================================================
# Testing recipes
# ============================================================
.PHONY: test
test: test-api

.PHONY: test-ui
test-ui:
	@echo "> UI unit testing ..."
	@cd ${UI_PATH} && yarn test-ci

# Added -gcflags=all=-d=checkptr=0 flag to disable unsafe.Pointer checking (https://go.dev/doc/go1.14#compiler).
# This is a workaround to avoid error from murmur3 library (https://github.com/pingcap/tidb/issues/29086#issuecomment-951787585)
.PHONY: test-api
test-api: init-dep-api
	@echo "> API unit testing ..."
	@cd ${API_PATH} && gotest -race -cover -coverprofile cover.out -tags unit ${API_ALL_PACKAGES} -gcflags=all=-d=checkptr=0
	@cd ${API_PATH} && go tool cover -func cover.out

.PHONY: it-test-api-local
it-test-api-local: local-db
	@echo "> API integration testing locally ..."
	@cd ${API_PATH} && gotest -race -short -cover -coverprofile cover.out -tags unit,integration_local ${API_ALL_PACKAGES} -gcflags=all=-d=checkptr=0
	@cd ${API_PATH} && go tool cover -func cover.out

.PHONY: it-test-api-ci
it-test-api-ci:
	@echo "> API integration testing ..."
	@cd ${API_PATH} && gotest -race -short -cover -coverprofile cover.out -tags unit,integration ${API_ALL_PACKAGES} -gcflags=all=-d=checkptr=0
	@cd ${API_PATH} && go tool cover -func cover.out

.PHONY:
bench:
	@echo "> Running Benchmark ..."
	@cd ${API_PATH} && gotest ${API_ALL_PACKAGES} -bench=. -run=^$$

# ============================================================
# Building recipes
# ============================================================
.PHONY: build
build: build-ui build-api

.PHONY: build-ui
build-ui:
	@echo "> Building UI static build ..."
	@cd ${UI_PATH} && npm run build

.PHONY: build-api
build-api:
	@echo "> Building API binary ..."
	@cd ${API_PATH} && go build -o ../bin/${BIN_NAME} ./cmd/api

.PHONY: build-transformer
build-transformer:
	@echo "> Building Transformer binary ..."
	@cd ${API_PATH} && go build -o ../bin/${TRANSFORMER_BIN_NAME} ./cmd/transformer

.PHONY: build-inference-logger
build-inference-logger:
	@echo "> Building Inference Logger binary ..."
	@cd ${API_PATH} && go build -o ../bin/${INFERENCE_LOGGER_BIN_NAME} ./cmd/inference-logger

# ============================================================
# Run recipe
# ============================================================
.PHONY: run
run:
	@echo "> Running application ..."
	@./bin/${BIN_NAME}

.PHONY: run-ui
run-ui:
	@echo "> Running UI ..."
	@cd ui && yarn start

.PHONY: run-inference-logger
run-inference-logger:
	@echo "> Running Inference Logger ..."
	@rm /tmp/agent.sock || true
	@cd api && SERVING_READINESS_PROBE='{"tcpSocket":{"port":8080,"host":"127.0.0.1"},"successThreshold":1}' UNIX_SOCKET_PATH="/tmp/agent.sock" go run $(TEST_TAGS) cmd/inference-logger/main.go -log-url="$(LOG_URL)"

.PHONY: run-mock-model-server
run-mock-model-server:
	@cd api && go run pkg/inference-logger/mock-server/mock_server.go

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
# Generate code recipes
# ============================================================
.PHONY: generate
generate: generate-client generate-proto

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


.PHONY: generate-proto
generate-proto:
	@echo "> Generating specification configuration from Proto file..."
	@cd protos/merlin && \
		protoc -I=. \
		--go_out=../../api \
		--go_opt=module=github.com/caraml-dev/merlin \
		--go-json_out=../../api \
		--go-json_opt=module=github.com/caraml-dev/merlin \
		transformer/**/*.proto log/*.proto

# ============================================================
# Docker build
# ============================================================
.PHONY: docker-build
docker-build: docker-build-transformer docker-build-api docker-build-pyfunc docker-build-batch-predictor

.PHONY: docker-build-api
docker-build-api: build-ui
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)merlin:${VERSION})
	@DOCKER_BUILDKIT=1 docker build -t ${IMAGE_TAG} -f Dockerfile .
	@rm -rf build

.PHONY: docker-build-transformer
docker-build-transformer:
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)merlin-transformer:${VERSION})
	@DOCKER_BUILDKIT=1 docker build -t ${IMAGE_TAG} -f transformer.Dockerfile .

.PHONY: docker-build-pyfunc
docker-build-pyfunc:
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)merlin/merlin-pyfunc-base-py${PYTHON_VERSION}:${VERSION})
	@DOCKER_BUILDKIT=1 docker build -t ${IMAGE_TAG} \
			--build-arg PYTHON_VERSION=${PYTHON_VERSION} \
			-f python/pyfunc-server/docker/base.Dockerfile python

.PHONY: docker-build-batch-predictor
docker-build-batch-predictor:
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)merlin-pyspark-base:${VERSION})
	@DOCKER_BUILDKIT=1 docker build -t ${IMAGE_TAG} -f python/batch-predictor/docker/base.Dockerfile python
