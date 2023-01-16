
.PHONY: dev_install
dev_install:
	python -m pipenv install --skip-lock -e .[test] --verbose

.PHONY: setup
setup: dev_install

.PHONY: test
test: type_check
	python -m pipenv run pytest -W ignore -s

.PHONY: type_check
type_check:
	python -m pipenv run mypy --ignore-missing-imports pyfuncserver

.PHONY: benchmark
benchmark:
	cd benchmark && ./benchmark.sh
