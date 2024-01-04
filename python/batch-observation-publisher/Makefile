WORKFLOW_IMAGE ?= batch-observation-publisher:dev
WORKFLOW_VERSION ?= dev
FLYTE_PROJECT ?= flyteexamples
FLYTE_DOMAIN ?= development

.PHONY: setup
setup:
	@echo "Setting up environment..."
	@pip install -r requirements.txt --use-pep517
	@pip install -r requirements-dev.txt

.PHONY: pip-compile
pip-compile:
	@echo "Compiling requirements..."
	@python -m piptools compile

.PHONY: package
package:
	@echo "Packaging..."
	@pyflyte --pkgs publisher package -f --image ${WORKFLOW_IMAGE}

.PHONY: register
register: package
	@echo "Registering..."
	flytectl register files --version ${WORKFLOW_VERSION} --archive flyte-package.tgz --project ${FLYTE_PROJECT} --domain ${FLYTE_DOMAIN}
