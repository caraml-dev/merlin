OBSERVATION_PUBLISHER_IMAGE_TAG = observation-publisher:dev

.PHONY: observation-publisher
observation-publisher:
	@echo "Building image for observation publisher..."
	@docker build -t ${OBSERVATION_PUBLISHER_IMAGE_TAG} -f observation-publisher/Dockerfile --progress plain .
