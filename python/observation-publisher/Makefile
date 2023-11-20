ENVIRONMENT_CONFIG = "example-override"

.PHONY: setup
setup:
	@echo "Setting up environment..."
	@pip install -r requirements.txt --use-pep517

.PHONY: test
test:
	@echo "Running tests..."
	@python -m pytest

.PHONY: run
run:
	@echo "Running observation publisher..."
	@python -m observation_publisher +environment=${ENVIRONMENT_CONFIG}