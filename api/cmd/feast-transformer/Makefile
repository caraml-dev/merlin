include .env.sample
export

BIN_NAME=feast-transformer

all: gen-spec-proto init-dep clean build run

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
	@go build -o bin/${BIN_NAME} ./main.go

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
clean:
	@echo "> Cleaning up existing compiled binary ..."
	@test ! -e bin || rm -r bin

.PHONY: gen-spec-proto
gen-spec-proto:
	@echo "> Generating specification configuration from Proto file..."
	@cd ../../pkg/transformers/feast && protoc -I=. --go_out=. --go-json_out=. --go_opt=paths=source_relative spec/*.proto
