# Build the releases expected by chart 3.6.3; upstream container images are no
# longer publicly pullable. Use the original toolchain for these older sources.
FROM golang:1.17.13 AS builder

ARG COMPONENT
ARG RELEASE
ARG SOURCE_SHA256

WORKDIR /src
ADD https://github.com/minio/${COMPONENT}/archive/refs/tags/${RELEASE}.tar.gz /tmp/source.tar.gz
RUN echo "${SOURCE_SHA256}  /tmp/source.tar.gz" | sha256sum -c - && \
    tar -xzf /tmp/source.tar.gz --strip-components=1 -C /src

ENV CGO_ENABLED=0
RUN mkdir -p /out && \
    go build -mod=readonly -trimpath \
        -ldflags "-s -w -X github.com/minio/${COMPONENT}/cmd.ReleaseTag=${RELEASE}" \
        -o /out/${COMPONENT} . && \
    if [ "${COMPONENT}" = minio ]; then \
        cp dockerscripts/docker-entrypoint.sh /out/docker-entrypoint.sh; \
    fi

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/ /usr/bin/
COPY --from=builder /src/LICENSE /licenses/LICENSE
# The Helm chart supplies the commands for both the server and the client jobs.
