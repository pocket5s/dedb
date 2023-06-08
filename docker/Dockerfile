FROM golang:1.19 as base

RUN mkdir -p /go/src/github.com/pocket5s/dedb
WORKDIR /go/src/github.com/pocket5s/dedb
RUN apt-get update
RUN apt -y install protobuf-compiler
RUN git config --global --add url."git@github.com:".insteadOf "https://github.com/"
RUN mkdir /root/.ssh/
RUN echo "StrictHostKeyChecking no" >> /root/.ssh/config
RUN go env -w GOPRIVATE=github.com/pocket5s/*

#RUN --mount=type=secret,id=API_TOKEN_GITHUB git config --global url."https://$(cat /run/secrets/API_TOKEN_GITHUB)@github.com/nxcr-org".insteadOf "https://github.com/nxcr-org"

ADD id_rsa /root/.ssh/

COPY go.mod .
COPY go.sum .

COPY . .
RUN make build

### Certs
#FROM alpine:latest as certs

### App
FROM alpine:latest as app
COPY --from=base /go/src/github.com/pocket5s/dedb/app /

#RUN GRPC_HEALTH_PROBE_VERSION=v0.3.1 && \
#  wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
#  chmod +x /bin/grpc_health_probe

ENTRYPOINT ["/app"]
