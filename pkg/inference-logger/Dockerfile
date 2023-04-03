FROM golang:1.20-alpine as go-builder

ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64

RUN apk update && apk add --no-cache git ca-certificates bash musl-dev gcc
RUN mkdir -p /src/log-collector

WORKDIR /src/log-collector
COPY . .

RUN go build \
    -mod=vendor \
    -tags musl \
    -o bin/log-collector ./cmd/main.go

FROM alpine:3.12

COPY --from=go-builder /src/log-collector/bin/log-collector /log-collector

ENTRYPOINT ["/log-collector"]
