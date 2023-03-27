export OBJC_DISABLE_INITIALIZE_FORK_SAFETY=YES

.PHONY: setup
setup:
	@pipenv install --skip-lock -e .[test]

.PHONY: type_check
type_check:
	@pipenv run mypy --ignore-missing-imports --allow-untyped-globals  merlinpyspark

.PHONY: test
test: type_check
	@gcloud config set project project-name
	@pipenv run pytest --cov=merlinpyspark -W ignore

.PHONY: unit-test
unit-test: type_check
	@gcloud config set project project-name
	@pipenv run pytest --cov=merlinpyspark -W ignore -m "not ci"

.PHONY: proto
proto:
	go get github.com/mitchellh/protoc-gen-go-json
	protoc -I=. --python_out=merlinpyspark --mypy_out=merlinpyspark --go_out=pkg/ --go-json_out=pkg/ --go_opt=paths=source_relative spec/*.proto
