ENVIRONMENT_CONFIG = "example-override"
VERSION=$(shell git describe --tags --always --first-parent)
ifeq ($(strip $(VERSION)),)
VERSION=0.1.0
endif

.PHONY: setup
setup:
	@echo "Setting up environment..."
	@pip install -r requirements.txt --use-pep517
	@pip install -r requirements-dev.txt

.PHONY: pip-compile
pip-compile:
	@echo "Compiling requirements..."
	@python -m piptools compile

.PHONY: test
test:
	@echo "Running tests..."
	@python -m pytest

.PHONY: run
run:
	@echo "Running observation publisher..."
	@python -m publisher +environment=${ENVIRONMENT_CONFIG}
