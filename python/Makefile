OBSERVATION_PUBLISHER_IMAGE_TAG ?= observation-publisher:dev
BATCH_OBSERVATION_PUBLISHER_IMAGE_TAG ?= batch-observation-publisher:dev

.PHONY: observation-publisher
observation-publisher:
	@echo "Building image for observation publisher..."
	@docker build -t ${OBSERVATION_PUBLISHER_IMAGE_TAG} -f observation-publisher/Dockerfile --progress plain .

.PHONY: batch-observation-publisher
batch-observation-publisher:
	@echo "Building image for batch observation publisher..."
	@docker build -t ${BATCH_OBSERVATION_PUBLISHER_IMAGE_TAG} -f batch-observation-publisher/Dockerfile --progress plain .
