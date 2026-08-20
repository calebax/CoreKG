ARG BUILDER_IMAGE=golang:1.24
ARG RUNTIME_IMAGE=debian:bookworm-slim
ARG TARGETOS=linux
ARG TARGETARCH

FROM ${BUILDER_IMAGE} AS building

ENV APP=keinit
ENV CGO_ENABLED=0
# ENV GOPROXY=https://goproxy.cn,direct
# ENV GOCACHE=/go/src/owl/.gocache
# ENV GOMODCACHE=/go/src/owl/.gocache/mod

COPY . /go/src/roc

WORKDIR /go/src/roc
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make local

# FROM alpine:3.17
FROM ${RUNTIME_IMAGE}

RUN apt-get update && apt-get install -y gettext-base && rm -rf /var/lib/apt/lists/*

COPY --from=building /go/src/roc/bundles/keinit /usr/local/bin/keinit

EXPOSE 8080
VOLUME [ "/app/config" ]

COPY ./scripts/mysql /app/scripts/mysql
COPY ./scripts/mysqlcoze /app/scripts/mysqlcoze
COPY ./scripts/es /app/scripts/es
# 原私有镜像 minio-init 初始化脚本改为由 docker-compose 中的 minio/mc init 容器完成。
# 见仓库根 docker-compose.yml 的 minio-init 服务。
WORKDIR /app/

ENTRYPOINT ["/usr/local/bin/keinit"]