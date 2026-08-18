FROM golang:1.24 AS building

ENV APP=corekg
ENV CGO_ENABLED=0
ENV DEPLOY_MODE=on_premise
# ENV DEPLOY_MODE=${DEPLOY_MODE}
# ENV GOPROXY=https://goproxy.cn,direct
# ENV GOCACHE=/go/src/owl/.gocache
# ENV GOMODCACHE=/go/src/owl/.gocache/mod
COPY . /go/src/roc

WORKDIR /go/src/roc

RUN make local

# FROM alpine:3.17
FROM debian:bookworm-slim


COPY --from=building /go/src/roc/bundles/corekg /usr/local/bin/corekg

EXPOSE 8080
VOLUME [ "/data" ]

ENTRYPOINT ["/usr/local/bin/corekg"]