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
# Build stage 1: Build API
# ============================================================
FROM golang:1.14-alpine as go-builder

RUN apk update && apk add --no-cache git ca-certificates bash
RUN mkdir -p src/api

WORKDIR /src/api
COPY api .
COPY python/batch-predictor ../python/batch-predictor
COPY db-migrations ./db-migrations

RUN go build -o bin/merlin_api ./cmd/api

# ============================================================
# Build stage 2: Build UI
# ============================================================
FROM node:14 as node-builder
WORKDIR /src/ui
COPY ui .
RUN yarn
RUN yarn run build

# ============================================================
# Build stage 3: Copy binary and config file
# ============================================================
FROM alpine:3.12

COPY --from=go-builder /src/api/bin/merlin_api /usr/bin/merlin_api
COPY --from=go-builder /src/api/db-migrations ./db-migrations
COPY --from=node-builder /src/ui/build ./ui/build

CMD ["merlin_api"]
