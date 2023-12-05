.PHONY: setup
setup:
	@pipenv install --dev --skip-lock

.PHONY: type-check
type-check:
	@pipenv run mypy --ignore-missing-imports --allow-untyped-globals --implicit-optional merlin

.PHONY: test
test: type-check
	@pipenv run pytest -n 8 --cov=merlin -W ignore

.PHONY: integration-test
integration-test: type-check
	@pipenv run pytest -n 8 -m integration -W ignore

.PHONY: integration-test-no-feast
integration-test-no-feast: type-check
	@pipenv run pytest -n 8 -m "not feast" -W ignore

.PHONY: local-server-test
local-server-test: type-check
	@pipenv run pytest -n 8 -m local_server_test -W ignore

.PHONY: unit-test
unit-test: type-check
	@pipenv run pytest -n 1 -m "not integration" -W ignore
