.PHONY: docs
docs: setup format

.PHONY: setup
setup:
	@rm -rf mdformatter
	@git clone https://github.com/caraml-dev/mdformatter.git
	@pip install -r mdformatter/requirements.txt

# The target below uses a non-existent doc overrides folder name to generate the final docs,
# as there are no overrides.
.PHONY: format
format:
	@echo "Formatting maintainer docs ..."
	@cd mdformatter && python -m mdformatter ../maintainer/templates ../maintainer/overrides ../maintainer/generated ../maintainer/values.json GITBOOK
	@echo "Formatting user docs ..."
	@cd mdformatter && python -m mdformatter ../user/templates ../user/overrides ../user/generated ../user/values.json GITBOOK
