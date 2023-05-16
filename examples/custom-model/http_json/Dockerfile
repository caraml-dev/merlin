FROM golang:1.20-alpine as go-builder

RUN apk update && apk add --no-cache git ca-certificates bash
RUN mkdir -p src

WORKDIR /src
COPY model-server .

RUN go build -o bin/custom_model


FROM alpine:3.12

COPY --from=go-builder /src/bin/custom_model /usr/bin/custom_model
CMD ["custom_model"]
