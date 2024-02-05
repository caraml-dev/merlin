ENVIRONMENT_CONFIG = "example-override"

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
	@python -m pytest -m "not integration"

.PHONY: test-integration
integration-test:
	@echo "Running integration tests..."
	@python -m pytest -m "integration"

.PHONY: run
run:
	@echo "Running observation publisher..."
	@python -m publisher +environment=${ENVIRONMENT_CONFIG}
