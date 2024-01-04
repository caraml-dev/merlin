FROM python:3.10-slim-buster

WORKDIR /root
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8

RUN apt-get update && apt-get install -y build-essential curl
RUN curl -sL https://ctl.flyte.org/install | bash -s -- -b /usr/local/bin v0.8.5

COPY sdk/ ./sdk
WORKDIR batch-observation-publisher
COPY batch-observation-publisher/requirements.txt requirements.txt
RUN pip install -r requirements.txt
COPY batch-observation-publisher/publisher publisher
COPY batch-observation-publisher/Makefile Makefile
ENV PYTHONPATH /root/batch-observation-publisher
ARG tag
ENV FLYTE_INTERNAL_IMAGE $tag