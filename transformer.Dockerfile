# Copyright 2020 The Merlin Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# ============================================================
# Build stage 1: Build web server
# ============================================================
FROM golang:1.17-alpine as go-builder

ARG BRANCH
ARG REVISION
ARG VERSION

RUN apk update && apk add --no-cache git ca-certificates bash
RUN mkdir -p src/api

WORKDIR /src/api
COPY api .
COPY python/batch-predictor ../python/batch-predictor

RUN go mod tidy
RUN go build -o bin/transformer \
    -ldflags="-X github.com/prometheus/common/version.Branch=${BRANCH} \
    -X github.com/prometheus/common/version.Revision=${REVISION} \
    -X github.com/prometheus/common/version.Version=${VERSION}" \
    ./cmd/transformer/main.go

# ============================================================
# Build stage 2: Copy binary
# ============================================================
FROM alpine:3.12

COPY --from=go-builder /src/api/bin/transformer /usr/bin/transformer
COPY --from=go-builder /usr/local/go/lib/time/zoneinfo.zip /zoneinfo.zip

ENV ZONEINFO=/zoneinfo.zip

ENTRYPOINT ["transformer"]
