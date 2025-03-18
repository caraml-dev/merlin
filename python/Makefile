OBSERVATION_PUBLISHER_IMAGE_TAG ?= observation-publisher:dev

.PHONY: observation-publisher
observation-publisher:
	@echo "Building image for observation publisher..."
	@docker build -t ${OBSERVATION_PUBLISHER_IMAGE_TAG} -f observation-publisher/Dockerfile --progress plain .

VER = $(subst v,,$(VERSION))

# to update merlin-sdk version in packages that depend on it
.PHONY: update-sdk-version
update-sdk-version:
	sed -i -e "s|merlin-sdk.*|merlin-sdk==${VER}|g" ./batch-predictor/requirements.txt
	sed -i -e "s|merlin-sdk.*|merlin-sdk==${VER}|g" ./pyfunc-server/requirements.txt
