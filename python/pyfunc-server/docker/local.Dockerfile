ARG BASE_IMAGE
FROM ${BASE_IMAGE}

# Download and install user model dependencies
ARG MODEL_DEPENDENCIES_URL
COPY ${MODEL_DEPENDENCIES_URL} conda.yaml

ARG MERLIN_DEP_CONSTRAINT
RUN process_conda_env.sh conda.yaml "merlin-pyfunc-server" "${MERLIN_DEP_CONSTRAINT}"
RUN conda env create --name merlin-model --file conda.yaml

# Download and dry-run user model artifacts and code
ARG MODEL_ARTIFACTS_URL
COPY ${MODEL_ARTIFACTS_URL} model
RUN /bin/bash -c ". activate merlin-model && merlin-pyfunc-server --model_dir model --dry_run"

CMD ["run.sh"]
