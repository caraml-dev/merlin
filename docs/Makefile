.PHONY: docs
docs: setup format

.PHONY: setup
setup:
	@rm -rf mdformatter
	@git clone https://github.com/caraml-dev/mdformatter.git
	@pip install -r mdformatter/requirements.txt

.PHONY: format
format: 
	cd mdformatter && python -m mdformatter ../user/templates ../user/overrides ../user/generated ../user/values.json GITBOOK
