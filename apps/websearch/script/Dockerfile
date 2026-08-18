ARG BUILDER_IMAGE=golang:1.24
ARG RUNTIME_IMAGE=debian:bookworm-slim
ARG TARGETOS=linux
ARG TARGETARCH

FROM ${BUILDER_IMAGE} AS building

ARG TARGETOS=linux
ARG TARGETARCH
ENV APP=websearch
ENV CGO_ENABLED=0

COPY . /go/src/roc
WORKDIR /go/src/roc
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make local

FROM ${RUNTIME_IMAGE}

COPY --from=building /go/src/roc/bundles/websearch /usr/local/bin/websearch

EXPOSE 8080
VOLUME [ "/app/config", "/data" ]
WORKDIR /app
ENTRYPOINT ["/usr/local/bin/websearch"]
