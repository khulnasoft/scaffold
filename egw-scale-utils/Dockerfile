ARG BASE_IMAGE=docker.io/library/alpine:3.21.3@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c
ARG GOLANG_IMAGE=docker.io/library/golang:1.24.1@sha256:fa145a3c13f145356057e00ed6f66fbd9bf017798c9d7b2b8e956651fe4f52da

FROM ${GOLANG_IMAGE} AS builder
ADD . /go/src/github.com/khulnasoft/scaffold/egw-scale-utils
WORKDIR /go/src/github.com/khulnasoft/scaffold/egw-scale-utils
RUN make local

FROM ${BASE_IMAGE}

COPY --from=builder /go/src/github.com/khulnasoft/scaffold/egw-scale-utils/egw-scale-utils /usr/bin/egw-scale-utils

ENTRYPOINT ["/usr/bin/egw-scale-utils"]