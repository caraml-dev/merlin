ARG PYTHON_VERSION=3.8
FROM python:${PYTHON_VERSION}-slim-buster

LABEL org.opencontainers.image.source https://github.com/caraml-dev/merlin

WORKDIR /root
ENV PYTHONPATH /root

RUN apt-get update && apt-get install build-essential curl vim wget -y

COPY . .${WORKDIR}

RUN pip install .
RUN pip install ".[test]"
