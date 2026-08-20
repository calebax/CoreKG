ARG BUILDER_IMAGE=golang:1.24
ARG RUNTIME_IMAGE=debian:bookworm-slim
ARG TARGETOS=linux
ARG TARGETARCH

FROM ${BUILDER_IMAGE} AS building

ENV APP=corekg
ENV CGO_ENABLED=0
ENV DEPLOY_MODE=on_premise
# ENV DEPLOY_MODE=${DEPLOY_MODE}
# ENV GOPROXY=https://goproxy.cn,direct
# ENV GOCACHE=/go/src/owl/.gocache
# ENV GOMODCACHE=/go/src/owl/.gocache/mod
COPY . /go/src/roc

WORKDIR /go/src/roc

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make local

# FROM alpine:3.17
FROM ${RUNTIME_IMAGE}

# debian:bookworm-slim 不含 CA 证书，Go 出站 HTTPS（如访问 LLM/外呼）会因
# "x509: certificate signed by unknown authority" 失败；安装系统根证书。
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=building /go/src/roc/bundles/corekg /usr/local/bin/corekg

EXPOSE 8080
VOLUME [ "/data" ]

ENTRYPOINT ["/usr/local/bin/corekg"]