
.PHONY: dev_install
dev_install:
	pipenv install --skip-lock -e .[test] --verbose

.PHONY: setup
setup: dev_install

.PHONY: test
test: type_check
	pipenv run pytest -W ignore -s

.PHONY: type_check
type_check:
	pipenv run mypy --ignore-missing-imports pyfuncserver

.PHONY: benchmark
benchmark:
	cd benchmark && ./benchmark.sh
