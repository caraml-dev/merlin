IMAGE_TAG=dev

.PHONY: observation-publisher
observation-publisher:
	@echo "Building image for observation publisher..."
	@docker build -t observation-publisher:${IMAGE_TAG} -f observation-publisher/Dockerfile .
