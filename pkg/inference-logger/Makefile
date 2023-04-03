#https://stackoverflow.com/questions/10858261/abort-makefile-if-variable-not-set
check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

VERSION := $(or ${VERSION}, $(shell git describe --tags --always --first-parent))
APP_NAME=merlin-logger
LOG_URL?=localhost:8002
TEST_TAGS?=

# ============================================================
# Testing recipes
# ============================================================
.PHONY: test
test: fmt vet
	@go test $(TEST_TAGS) -race ./pkg/... ./cmd/... -cover -coverprofile coverage.out
	@go tool cover -func coverage.out

# ============================================================
# Analyze source code recipes
# ============================================================
.PHONY: lint
lint:
	@echo "> Analyzing API source code..."
	@golangci-lint run

# Run go fmt against code
.PHONY: fmt
fmt:
	@go fmt ./pkg/... ./cmd/...

# Run go vet against code
.PHONY: vet
vet:
	@go vet ./pkg/... ./cmd/...

.PHONY: vendor
vendor:
	@echo "Fetching dependencies..."
	go mod vendor

# ============================================================
# Building recipes
# ============================================================
.PHONY: build
build: test
	@go build -o bin/${APP_NAME} ./cmd/main.go

.PHONY: docker-build
docker-build:
	$(call check_defined, DOCKER_REGISTRY, DOCKER_REGISTRY must be set)
	$(call check_defined, VERSION, VERSION must be set)
	@go mod vendor
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)merlin-logger:${VERSION})
	@docker build -t ${IMAGE_TAG} .

# ============================================================
# Run recipes
# ============================================================
.PHONY: run
run:
	@rm /tmp/agent.sock || true
	@SERVING_READINESS_PROBE='{"tcpSocket":{"port":8080,"host":"127.0.0.1"},"successThreshold":1}' UNIX_SOCKET_PATH="/tmp/agent.sock" go run $(TEST_TAGS) cmd/main.go -log-url="$(LOG_URL)"

.PHONY: run-mock
run-mock:
	@go run mock-server/mock_server.go

