# ============================================================
# Build stage 1: Build web server
# ============================================================
FROM golang:1.20-alpine as go-builder

RUN mkdir -p /src/model
WORKDIR /src/model

COPY . .

RUN go build -o bin/model-server cmd/main.go


## ============================================================
## Build stage 2: Copy binary
## ============================================================
#
FROM debian:buster-slim

COPY --from=go-builder /src/model/bin/model-server /usr/bin/model-server

ENTRYPOINT ["model-server"]
