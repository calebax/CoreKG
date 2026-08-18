ARG BUILDER_IMAGE=golang:1.24
ARG RUNTIME_IMAGE=debian:bookworm-slim
ARG TARGETOS=linux
ARG TARGETARCH

FROM ${BUILDER_IMAGE} AS building

ENV APP=ketask
ENV CGO_ENABLED=0

COPY . /go/src/roc

WORKDIR /go/src/roc
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make local

FROM ${RUNTIME_IMAGE}

COPY --from=building /go/src/roc/bundles/ketask /usr/local/bin/ketask

EXPOSE 8080
VOLUME [ "/data" ]

ENTRYPOINT ["/usr/local/bin/ketask"]
