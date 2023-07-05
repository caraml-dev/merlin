BIN_NAME=feast-transformer
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
REVISION=$(shell git rev-parse HEAD)
VERSION=$(shell git describe --tags --always --dirty)

all: init-dep clean build run

# ============================================================
# Initialize dependency recipes
# ============================================================
.PHONY: init-dep
init-dep:
	@echo "> Initializing dependencies ..."
	@go mod tidy -v
	@go get -v ./...

# ============================================================
# Building recipes
# ============================================================
.PHONY: build
build: clean
	@echo "> Building API binary ..."
	@go build -o bin/${BIN_NAME} \
		-ldflags="-X github.com/prometheus/common/version.Branch=${BRANCH} \
					-X github.com/prometheus/common/version.Revision=${REVISION} \
					-X github.com/prometheus/common/version.Version=${VERSION}" \
		./main.go

# ============================================================
# Run recipe
# ============================================================
.PHONY: run
run:
	@echo "> Running application ..."
	@go run main.go

# ============================================================
# Utility recipes
# ============================================================
.PHONY: clean
clean:
	@echo "> Cleaning up existing compiled binary ..."
	@test ! -e bin || rm -r bin
