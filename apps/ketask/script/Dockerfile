FROM golang:1.24 AS building

ENV APP=ketask
ENV CGO_ENABLED=0

COPY . /go/src/roc

WORKDIR /go/src/roc
RUN make local

FROM debian:bookworm-slim

COPY --from=building /go/src/roc/bundles/ketask /usr/local/bin/ketask

EXPOSE 8080
VOLUME [ "/data" ]

ENTRYPOINT ["/usr/local/bin/ketask"]
