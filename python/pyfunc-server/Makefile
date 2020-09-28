
dev_install:
	pipenv install --skip-lock -e .[test]

test: type_check
	pipenv run pytest -W ignore -s

type_check:
	pipenv run mypy --ignore-missing-imports pyfuncserver