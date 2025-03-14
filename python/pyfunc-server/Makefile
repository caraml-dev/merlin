.PHONY: setup
setup:
	$(MAKE) install_sdk
	pip install pipenv
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

.PHONY: depp build_install_sdk

SDK_VER := $(shell grep '^merlin-sdk' requirements.txt | awk -F '[=<>]+' '{print $$2}')
# check if installation of merlin-sdk  from pypi is succesful, if not build from local (latest / unreleased version)
install_sdk:
	pip install merlin-sdk==${SDK_VER} || $(MAKE) build_install_sdk

build_install_sdk:
	cd ../sdk && pip install setuptools setuptools_scm twine wheel
	cd ../sdk && sed -i -e "s|VERSION = \".*\"|VERSION = \"${SDK_VER}\"|g" ./merlin/version.py
	cd ../sdk && python setup.py sdist bdist_wheel
	pip install ../sdk/dist/merlin_sdk-${SDK_VER}-py3-none-any.whl