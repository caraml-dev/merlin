export OBJC_DISABLE_INITIALIZE_FORK_SAFETY=YES

PROTOC_GEN_GO_JSON_VERSION="v1.1.0"
PROTOC_GEN_GO_VERSION="v1.26"
PROTOC_PATH ?= protoc  # Default to 'protoc' if PROTOC_PATH is not set

.PHONY: setup
setup:
	@pip install pipenv
	$(MAKE) build_install_sdk
	@pipenv install --skip-lock -e .[test]

.PHONY: type_check
type_check:
	@pipenv run mypy --ignore-missing-imports --allow-untyped-globals merlinpyspark

.PHONY: test
test: type_check
	@gcloud config set project project-name
	@pipenv run pytest --cov=merlinpyspark -W ignore

.PHONY: unit-test
unit-test: type_check
	@gcloud config set project project-name
	@pipenv run pytest --cov=merlinpyspark -W ignore -m "not ci"

.PHONY: proto
proto:
	@echo "Checking libprotoc version..."
	@echo "protocPath=$(PROTOC_PATH)"
	@PROTOC_VERSION=$$($(PROTOC_PATH) --version | grep -o '[0-9]\+\.[0-9]\+\(\.[0-9]\+\)\?'); \
	if [ "$$PROTOC_VERSION" != "25.6" ]; then \
		echo "Error: libprotoc version 25.6 is required. Found version $$PROTOC_VERSION."; \
		echo "Download protoc version from https://repo1.maven.org/maven2/com/google/protobuf/protoc/3.25.6/"; \
		exit 1; \
	fi
	go install github.com/mitchellh/protoc-gen-go-json@${PROTOC_GEN_GO_JSON_VERSION}
	@$(PROTOC_PATH) -I=. --python_out=merlinpyspark --mypy_out=merlinpyspark --go_out=pkg/ --go-json_out=pkg/ --go_opt=paths=source_relative spec/*.proto


.PHONY: build_install_sdk

SDK_VER := $(shell grep '^merlin-sdk' requirements.txt | awk -F '[=<>]+' '{print $$2}')

build_install_sdk:
	cd ../sdk && pip install setuptools setuptools_scm twine wheel
	cd ../sdk && sed -i -e "s|VERSION = \".*\"|VERSION = \"${SDK_VER}\"|g" ./merlin/version.py
	cd ../sdk && python setup.py sdist bdist_wheel
	pipenv run pip install ../sdk/dist/merlin_sdk-${SDK_VER}-py3-none-any.whl