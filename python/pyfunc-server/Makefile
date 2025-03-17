.PHONY: setup
setup:
	pip install pipenv
	$(MAKE) build_install_sdk
	pipenv install --skip-lock -e .[test]

.PHONY: test
test: type_check
	pipenv run pytest -W ignore -s

.PHONY: type_check
type_check:
	pipenv run mypy --ignore-missing-imports pyfuncserver

.PHONY: benchmark
benchmark:
	cd benchmark && ./benchmark.sh

.PHONY: build_install_sdk

SDK_VER := $(shell grep '^merlin-sdk' requirements.txt | awk -F '[=<>]+' '{print $$2}')

build_install_sdk:
	cd ../sdk && pip install setuptools setuptools_scm twine wheel
	cd ../sdk && sed -i -e "s|VERSION = \".*\"|VERSION = \"${SDK_VER}\"|g" ./merlin/version.py
	cd ../sdk && python setup.py sdist bdist_wheel
	pipenv run pip install ../sdk/dist/merlin_sdk-${SDK_VER}-py3-none-any.whl