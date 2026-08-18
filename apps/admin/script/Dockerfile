# FROM golang:1.20 AS building
FROM golang:1.24 AS building

ENV APP=admin
ENV CGO_ENABLED=0
# ENV GOPROXY=https://goproxy.cn,direct
# ENV GOCACHE=/go/src/owl/.gocache
# ENV GOMODCACHE=/go/src/owl/.gocache/mod

COPY . /go/src/roc

WORKDIR /go/src/roc
RUN make local

FROM debian:bookworm-slim


COPY --from=building /go/src/roc/bundles/admin /usr/local/bin/admin

EXPOSE 8080
VOLUME [ "/data" ]

ENTRYPOINT ["/usr/local/bin/admin"]