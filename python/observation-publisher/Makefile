IMAGE_NAME = "observation-publisher"
IMAGE_TAG = "dev"
ENVIRONMENT_CONFIG = "example-override"

.PHONY: test
test:
	@echo "Running tests..."
	@python -m pytest

.PHONY: image
image:
	@echo "Building image..."
	@docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

.PHONY: run
run:
	@echo "Running observation publisher..."
	@python -m observation_publisher +environment=${ENVIRONMENT_CONFIG}