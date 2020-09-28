{%- macro model_class() %}
{%- set split_model_slug  = cookiecutter.model_slug.split("_") %}
{%- set class_components = [] %}
{%- for x in split_model_slug %}
{{- class_components.append(x.capitalize()) or ""}}
{%- endfor %}
{%- set class_name = class_components|join %}
{{- class_name ~ "Model" }}
{%- endmacro %}

import merlin
from merlin.model import ModelType, PyFuncModel
from src.{{ cookiecutter.model_slug }} import {{ model_class() }}
import os
import src.constant as constant
import sys

model_name = os.getenv("MODEL_NAME")
model_path = os.getenv("MODEL_PATH")
merlin_url = os.getenv("MERLIN_URL")
merlin_project = os.getenv("MERLIN_PROJECT")
merlin_env_name = os.getenv("MERLIN_ENV_NAME")

if model_name is None or model_name == "":
    sys.exit("Model name could not be empty")
if model_path is None or model_path == "":
    sys.exit("Model path could not be empty")
if merlin_url is None or merlin_url == "":
    sys.exit("Merlin URL could not be empty")
if merlin_project is None or merlin_project == "":
    sys.exit("Merlin project could not be empty")
if merlin_env_name is None or merlin_env_name == "":
    sys.exit("Environment name could not be empty")

merlin.set_url(merlin_url)
merlin.set_project(merlin_project)
merlin.set_model(model_name, ModelType.PYFUNC)

with merlin.new_model_version() as v:
    merlin.log_pyfunc_model(
        model_instance={{ model_class() }}(merlin_project, model_name),
        conda_env="env/conda.yaml",
        code_dir=["src"],
        artifacts={
            constant.ARTIFACT_MODEL_PATH: model_path,
        },
    )
    merlin.deploy(v, merlin_env_name)
