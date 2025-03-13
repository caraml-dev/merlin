OBSERVATION_PUBLISHER_IMAGE_TAG ?= observation-publisher:dev

.PHONY: observation-publisher
observation-publisher:
	@echo "Building image for observation publisher..."
	@docker build -t ${OBSERVATION_PUBLISHER_IMAGE_TAG} -f observation-publisher/Dockerfile --progress plain .

SDK_VER = $(subst v,,$(SDK_VERSION))

.PHONY: upgrade-sdk-version
upgrade-sdk-version:
	python -m pip install --upgrade pip
	pip install setuptools setuptools_scm twine wheel
	sed -i -e "s|merlin-sdk.*|merlin-sdk==$(SDK_VER)|g" ./batch-predictor/requirements.txt
	sed -i -e "s|merlin-sdk.*|merlin-sdk==$(SDK_VER)|g" ./pyfunc-server/requirements.txt
	cd sdk && sed -i -e "s|VERSION = \".*\"|VERSION = \"$(SDK_VER)\"|g" ./merlin/version.py
	cd sdk && python setup.py sdist bdist_wheel
	pip install sdk/dist/merlin_sdk-$(SDK_VER)-py3-none-any.whl
