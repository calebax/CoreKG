# FROM golang:1.20 AS building
FROM golang:1.24 AS building

ENV APP=kecore
ENV CGO_ENABLED=0
# ENV GOPROXY=https://goproxy.cn,direct
# ENV GOCACHE=/go/src/owl/.gocache
# ENV GOMODCACHE=/go/src/owl/.gocache/mod

COPY . /go/src/roc

WORKDIR /go/src/roc
RUN make local

# FROM alpine:3.17
FROM debian:bookworm-slim


COPY --from=building /go/src/roc/bundles/kecore /usr/local/bin/kecore

EXPOSE 8080
VOLUME [ "/data" ]

ENTRYPOINT ["/usr/local/bin/kecore"]
