# Copyright 2023 The Merlin Authors
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
FROM golang:1.20-alpine as go-builder

ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64

RUN apk update && apk add --no-cache git ca-certificates bash musl-dev gcc
RUN mkdir -p /src/log-collector

WORKDIR /src/log-collector

# Caching dependencies
COPY api/go.mod .
COPY api/go.sum .
COPY python/batch-predictor/go.mod ../python/batch-predictor/go.mod
COPY python/batch-predictor/go.sum ../python/batch-predictor/go.sum

RUN go mod download

COPY api .
COPY python/batch-predictor ../python/batch-predictor

RUN go mod tidy
RUN go build \
    -tags musl \
    -o bin/log-collector ./cmd/inference-logger/main.go

# ============================================================
# Build stage 2: Copy binary
# ============================================================
FROM alpine:3.12

COPY --from=go-builder /src/log-collector/bin/log-collector /log-collector

ENTRYPOINT ["/log-collector"]
